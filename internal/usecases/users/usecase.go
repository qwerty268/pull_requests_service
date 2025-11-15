//go:generate mockgen --source=usecase.go --destination=mocks/usecase.go -package=mocks

package users

import (
	"context"
	"errors"
	"fmt"

	repository "github.com/qwerty268/pull_request_service/internal/usecases/users/storage"
)

var ErrNotFound = errors.New("not found")

type storage interface {
	SetUserActive(userID string, isActive bool) (*repository.User, error)
	GetUserReviewRequests(userID string) ([]repository.PullRequestShort, error)
}

type Usecase struct {
	storage storage
}

func NewUsecase(storage storage) Usecase {
	return Usecase{
		storage: storage,
	}
}

func (u Usecase) SetUserActive(_ context.Context, userID string, isActive bool) (*User, error) {
	storageUser, err := u.storage.SetUserActive(userID, isActive)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, fmt.Errorf("failed to find user: %w", ErrNotFound)
		}
		return nil, fmt.Errorf("failed to set user active: %v", err)
	}
	user := User(*storageUser)
	return &user, nil
}

func (u Usecase) GetUserReviewRequests(_ context.Context, userID string) ([]PullRequestShort, error) {
	storagePrs, err := u.storage.GetUserReviewRequests(userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, fmt.Errorf("failed to get user PRs: %w", ErrNotFound)
		}
		return nil, fmt.Errorf("failed to get user PRs: %v", err)
	}
	return fromStoragePrs(storagePrs), nil
}

func fromStoragePrs(storagePrs []repository.PullRequestShort) []PullRequestShort {
	prs := make([]PullRequestShort, len(storagePrs))
	for i, v := range storagePrs {
		prs[i] = PullRequestShort(v)
	}
	return prs
}
