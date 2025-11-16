//go:generate mockgen --source=usecase.go --destination=mocks/usecase.go -package=mocks

package pullrequests

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	repository "github.com/qwerty268/pull_request_service/internal/usecases/pullrequests/storage"
)

var (
	ErrAlreadyExists = errors.New("already exists")
	ErrNotFound      = errors.New("not found")
	ErrPRMerged      = errors.New("already merged")
	ErrNotAssigned   = errors.New("not assigned")
	ErrNoCandidate   = errors.New("no condidate")
)

var GetRandomReviewer = getRandomReviewer // Чтобы тестить.

const (
	statusOpen   = "OPEN"
	statusMerged = "MERGED"
)

type prStorage interface {
	AddPr(pr repository.PullRequest) error
	SetPrMerged(prID string) (*repository.PullRequest, error)
	// CheckUserInPr проаеряет, что есть запись в таблице pr_user_map
	CheckUserInPr(prID, userID string) (bool, error)
	GetPrByID(prID string) (*repository.PullRequest, error)
	ResetPrMember(filter repository.ResetReviewerFilter) error
}

type teamStorage interface {
	// GetUserActiveTeammates выдает активных сокомандников, не включая самого пользователя.
	GetUserActiveTeammates(userID string) ([]string, error)
	// CheckUserInCommand смотрит существует ли запись про юзера в таблице team_user_map.
	// Подразумевается, что у пользователя одна команда.
	CheckUserInCommand(userID string) (bool, error)
}

type userStorage interface {
	CheckUserExists(userID string) (bool, error)
}

type Usecase struct {
	prStorage   prStorage
	teamStorage teamStorage
	userStorage userStorage
}

func NewUsecase(prStorage prStorage, teamStorage teamStorage, userStorage userStorage) Usecase {
	return Usecase{
		prStorage:   prStorage,
		teamStorage: teamStorage,
		userStorage: userStorage,
	}
}

func (u Usecase) CreatePR(_ context.Context, pr CreatePROpst) (*PullRequest, error) {
	newPr := &PullRequest{
		PullRequestID:   pr.PullRequestID,
		PullRequestName: pr.PullRequestName,
		AuthorID:        pr.AuthorID,
		Status:          statusOpen,
		CreatedAt:       time.Now(),
	}

	ok, err := u.userStorage.CheckUserExists(pr.AuthorID)
	if err != nil {
		return nil, fmt.Errorf("failed to check if users exists: %v", err)
	}
	if !ok {
		return nil, fmt.Errorf("user or team not exists: %w", ErrNotFound)
	}

	activeTeammates, err := u.teamStorage.GetUserActiveTeammates(pr.AuthorID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user teammates: %v", err)
	}
	reviewers := getRandomReviewers(activeTeammates)

	newPr.AssignedReviewers = reviewers

	err = u.prStorage.AddPr(repository.PullRequest(toStoragePR(newPr)))
	if err != nil {
		if errors.Is(err, repository.ErrAlreadyExists) {
			return nil, fmt.Errorf("failed to save pr: %w", ErrAlreadyExists)
		}
		return nil, fmt.Errorf("failed to save pr: %v", err)
	}

	return newPr, nil
}

func toStoragePR(pr *PullRequest) repository.PullRequest {
	repositoryPr := repository.PullRequest{
		PullRequestID:     pr.PullRequestID,
		PullRequestName:   pr.PullRequestName,
		AuthorID:          pr.AuthorID,
		AssignedReviewers: pr.AssignedReviewers,
		CreatedAt:         pr.CreatedAt,
		MergedAt:          pr.MergedAt,
	}
	if pr.Status == statusMerged {
		repositoryPr.IsMerged = true
	}
	return repositoryPr
}

func getRandomReviewers(activeTeammates []string) []string {
	if len(activeTeammates) <= 2 {
		return activeTeammates
	}
	n := len(activeTeammates)
	i := rand.Intn(n)
	j := rand.Intn(n - 1)
	if j >= i {
		j++
	}
	return []string{activeTeammates[i], activeTeammates[j]}
}

func (u Usecase) MergePR(ctx context.Context, prID string) (*PullRequest, error) {
	storagePr, err := u.prStorage.SetPrMerged(prID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, fmt.Errorf("failed to set merged flag: %w", ErrNotFound)
		}
		return nil, fmt.Errorf("failed to set merged flag: %v", err)
	}
	return fromStoragePr(storagePr), nil
}

func fromStoragePr(storagePr *repository.PullRequest) *PullRequest {
	pr := &PullRequest{
		PullRequestID:     storagePr.PullRequestID,
		PullRequestName:   storagePr.PullRequestName,
		AuthorID:          storagePr.AuthorID,
		AssignedReviewers: storagePr.AssignedReviewers,
		CreatedAt:         storagePr.CreatedAt,
		MergedAt:          storagePr.MergedAt,
	}
	if storagePr.IsMerged {
		pr.Status = statusMerged
	} else {
		pr.Status = statusOpen
	}
	return pr
}

func (u Usecase) ReassignReviewer(ctx context.Context, prID, oldUserID string) (*ReassignedRewiew, error) {
	// 1. Проверяем есть ли юезр и пр.
	storagePr, err := u.prStorage.GetPrByID(prID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, fmt.Errorf("failed to get pr from storage: %w", ErrNotFound)
		}
		return nil, fmt.Errorf("failed to get pr from storage: %v", err)
	}
	if storagePr.IsMerged {
		return nil, ErrPRMerged
	}

	userExists, err := u.userStorage.CheckUserExists(oldUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to check user exists: %v", err)
	}
	if !userExists {
		return nil, fmt.Errorf("check user exists: %w", ErrNotFound)
	}

	// 2. Проверяем, что пользователь прикреплен к пр.
	ok, err := u.prStorage.CheckUserInPr(prID, oldUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to check user in pr: %v", err)
	}
	if !ok {
		return nil, fmt.Errorf("failed to check user in pr: %w", ErrNotAssigned)
	}

	// 3. Выделяем чела, который должен остаться. Рассчет на то что ревьюеров может быть максимум 2.
	newReviewers := make([]string, 0)
	var reviewerShouldStay *string
	if len(storagePr.AssignedReviewers) > 1 {
		// Случай, что есть чел, которого надо оставить.
		if storagePr.AssignedReviewers[0] == oldUserID {
			reviewerShouldStay = &storagePr.AssignedReviewers[1]
		} else {
			reviewerShouldStay = &storagePr.AssignedReviewers[0]
		}
		newReviewers = append(newReviewers, *reviewerShouldStay)
	}

	// 4. Выделяем активных тиммейтов, которых можно назначить на ревью.
	activeMembers, err := u.teamStorage.GetUserActiveTeammates(oldUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active teammates: %v", err)
	}

	// 5. Из полученных пользователей вычитаем ревьюера, который должен остаться и автора.
	var mustRemove map[string]struct{}
	if reviewerShouldStay != nil {
		mustRemove = map[string]struct{}{
			*reviewerShouldStay: {},
			storagePr.AuthorID:  {},
		}
	} else {
		mustRemove = map[string]struct{}{
			storagePr.AuthorID: {},
		}
	}

	// Фильтруем activeMembers
	filtered := make([]string, 0, len(activeMembers))
	for _, v := range activeMembers {
		if _, toRemove := mustRemove[v]; !toRemove {
			filtered = append(filtered, v)
		}
	}
	activeMembers = filtered
	// Нет кондидатов.
	if len(activeMembers) == 0 {
		return nil, fmt.Errorf("failed to assign new condidate: %w", ErrNoCandidate)
	}

	// 6. Берем рандома.
	newReviewer := GetRandomReviewer(activeMembers)
	newReviewers = append(newReviewers, newReviewer)
	// 7. Меняем запись в бд.
	filter := repository.ResetReviewerFilter{
		PrID:      prID,
		OldUserID: oldUserID,
		NewUserID: newReviewer,
	}

	err = u.prStorage.ResetPrMember(filter)
	if err != nil {
		return nil, fmt.Errorf("failed tu reset pr member: %v", err)
	}

	updatedPR := fromStoragePr(storagePr)
	updatedPR.AssignedReviewers = newReviewers

	return &ReassignedRewiew{
		Pr:          *updatedPR,
		NewReviewer: newReviewer,
	}, nil
}

func getRandomReviewer(activeMembers []string) string {
	i := rand.Intn(len(activeMembers))
	return activeMembers[i]
}
