package users

import (
	"errors"
	"testing"

	"context"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/qwerty268/pull_request_service/internal/usecases/users/mocks"
	repository "github.com/qwerty268/pull_request_service/internal/usecases/users/storage"
)

func TestUsecase_SetUserActive(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockstorage(ctrl)
	usecase := NewUsecase(mockStorage)
	ctx := context.Background()

	expectedRepoUser := &repository.User{
		UserID:   "u10",
		Username: "Bruce",
		TeamName: "avengers",
		IsActive: true,
	}

	t.Run("success", func(t *testing.T) {
		mockStorage.EXPECT().
			SetUserActive("u10", true).
			Return(expectedRepoUser, nil)
		user, err := usecase.SetUserActive(ctx, "u10", true)
		require.NoError(t, err)
		require.NotNil(t, user)
		require.Equal(t, User(*expectedRepoUser), *user)
	})

	t.Run("not found", func(t *testing.T) {
		mockStorage.EXPECT().
			SetUserActive("u99", true).
			Return(nil, repository.ErrNotFound)
		user, err := usecase.SetUserActive(ctx, "u99", true)
		require.ErrorIs(t, err, ErrNotFound)
		require.Contains(t, err.Error(), "failed to find user")
		require.Nil(t, user)
	})

	t.Run("other storage error", func(t *testing.T) {
		mockStorage.EXPECT().
			SetUserActive("u10", false).
			Return(nil, errors.New("db down"))
		user, err := usecase.SetUserActive(ctx, "u10", false)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to set user active")
		require.Contains(t, err.Error(), "db down")
		require.Nil(t, user)
	})
}

func TestUsecase_GetUserReviewRequests(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockstorage(ctrl)
	usecase := NewUsecase(mockStorage)
	ctx := context.Background()

	prs := []repository.PullRequestShort{
		{
			PullRequestID:   "pr1",
			PullRequestName: "Fix bug",
			AuthorID:        "alice",
			Status:          "NEW",
		},
		{
			PullRequestID:   "pr2",
			PullRequestName: "Add feature",
			AuthorID:        "bob",
			Status:          "REVIEW",
		},
	}

	t.Run("success", func(t *testing.T) {
		mockStorage.EXPECT().
			GetUserReviewRequests("uid1").
			Return(prs, nil)

		result, err := usecase.GetUserReviewRequests(ctx, "uid1")
		require.NoError(t, err)
		require.Len(t, result, 2)
		require.Equal(t, PullRequestShort(prs[0]), result[0])
		require.Equal(t, PullRequestShort(prs[1]), result[1])
	})

	t.Run("not found", func(t *testing.T) {
		mockStorage.EXPECT().
			GetUserReviewRequests("nouser").
			Return(nil, repository.ErrNotFound)

		result, err := usecase.GetUserReviewRequests(ctx, "nouser")
		require.ErrorIs(t, err, ErrNotFound)
		require.Contains(t, err.Error(), "failed to get user PRs")
		require.Nil(t, result)
	})

	t.Run("other error", func(t *testing.T) {
		mockStorage.EXPECT().
			GetUserReviewRequests("failuser").
			Return(nil, errors.New("db is down"))

		result, err := usecase.GetUserReviewRequests(ctx, "failuser")
		require.Error(t, err)
		require.Contains(t, err.Error(), "db is down")
		require.Contains(t, err.Error(), "failed to get user PRs")
		require.Nil(t, result)
	})
}
