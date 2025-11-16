//go:generate mockgen --source=usecase.go --destination=mocks/usecase.go -package=mocks

package users

import (
	"context"
	"errors"
	"fmt"

	prRepository "github.com/qwerty268/pull_request_service/internal/usecases/pullrequests/storage"
	userRepository "github.com/qwerty268/pull_request_service/internal/usecases/users/storage"
)

var ErrNotFound = errors.New("not found")

const (
	statusOpen   = "OPEN"
	statusMerged = "MERGED"
)

type userStorage interface {
	SetUserActive(userID string, isActive bool) (*userRepository.User, error)
}

type prStorage interface {
	GetUserReviewRequests(userID string) ([]prRepository.PullRequestShort, error)
}

type Usecase struct {
	userStorage userStorage
	prStorage   prStorage
}

func NewUsecase(storage userStorage, prStorage prStorage) Usecase {
	return Usecase{
		userStorage: storage,
		prStorage:   prStorage,
	}
}

func (u Usecase) SetUserActive(_ context.Context, userID string, isActive bool) (*User, error) {
	storageUser, err := u.userStorage.SetUserActive(userID, isActive)
	if err != nil {
		if errors.Is(err, userRepository.ErrNotFound) {
			return nil, fmt.Errorf("failed to find user: %w", ErrNotFound)
		}
		return nil, fmt.Errorf("failed to set user active: %v", err)
	}
	user := User(*storageUser)
	return &user, nil
}

func (u Usecase) GetUserReviewRequests(_ context.Context, userID string) ([]PullRequestShort, error) {
	storagePrs, err := u.prStorage.GetUserReviewRequests(userID)
	if err != nil {
		if errors.Is(err, prRepository.ErrNotFound) {
			return nil, fmt.Errorf("failed to get user PRs: %w", ErrNotFound)
		}
		return nil, fmt.Errorf("failed to get user PRs: %v", err)
	}
	return fromStoragePrs(storagePrs), nil
}

func fromStoragePrs(storagePrs []prRepository.PullRequestShort) []PullRequestShort {
	prs := make([]PullRequestShort, len(storagePrs))
	for i, v := range storagePrs {
		prs[i] = PullRequestShort{
			PullRequestID:   v.PullRequestID,
			PullRequestName: v.PullRequestName,
			AuthorID:        v.AuthorID,
		}
		if v.IsMerged {
			prs[i].Status = statusMerged
		} else {
			prs[i].Status = statusOpen
		}
	}
	return prs
}
