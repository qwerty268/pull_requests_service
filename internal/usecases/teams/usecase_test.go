package teams

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/qwerty268/pull_request_service/internal/usecases/teams/mocks"
	repository "github.com/qwerty268/pull_request_service/internal/usecases/teams/storage"
)

func TestUsecase_AddTeam(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockstorage(ctrl)
	u := Usecase{storage: mockStorage}

	team := Team{TeamName: "myteam"}

	t.Run("success", func(t *testing.T) {
		mockStorage.EXPECT().
			AddTeam(toStorageTeam(team)).
			Return(nil)

		err := u.AddTeam(context.Background(), team)
		require.NoError(t, err)
	})

	t.Run("team already exists", func(t *testing.T) {
		mockStorage.EXPECT().
			AddTeam(toStorageTeam(team)).
			Return(repository.ErrAlreadyExists)

		err := u.AddTeam(context.Background(), team)
		require.ErrorIs(t, err, ErrAlreadyExists)
	})

	t.Run("other error", func(t *testing.T) {
		someErr := errors.New("db fail")
		mockStorage.EXPECT().
			AddTeam(toStorageTeam(team)).
			Return(someErr)

		err := u.AddTeam(context.Background(), team)
		require.Error(t, err)
		require.Contains(t, err.Error(), "db fail")
	})
}

func TestUsecase_GetTeam(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockstorage(ctrl)
	usecase := NewUsecase(mockStorage)
	ctx := context.Background()

	wantRepoTeam := &repository.Team{
		TeamName: "dream",
		Members: []repository.TeamMember{
			{UserID: "u1", Username: "Alice", IsActive: true},
			{UserID: "u2", Username: "Bob", IsActive: false},
		},
	}

	t.Run("success", func(t *testing.T) {
		mockStorage.EXPECT().
			GetTeam("dream").
			Return(wantRepoTeam, nil)

		got, err := usecase.GetTeam(ctx, "dream")
		require.NoError(t, err)
		require.NotNil(t, got)

		require.Equal(t, wantRepoTeam.TeamName, got.TeamName)
		require.Len(t, got.Members, 2)
		require.Equal(t, TeamMember(wantRepoTeam.Members[0]), got.Members[0])
		require.Equal(t, TeamMember(wantRepoTeam.Members[1]), got.Members[1])
	})

	t.Run("not found", func(t *testing.T) {
		mockStorage.EXPECT().
			GetTeam("dream").
			Return(nil, fmt.Errorf("store not found: %w", repository.ErrNotFound))

		got, err := usecase.GetTeam(ctx, "dream")
		require.ErrorIs(t, err, ErrNotFound)
		require.Nil(t, got)
	})

	t.Run("storage error", func(t *testing.T) {
		someErr := errors.New("db fail")
		mockStorage.EXPECT().
			GetTeam("dream").
			Return(nil, someErr)

		got, err := usecase.GetTeam(ctx, "dream")
		require.Error(t, err)
		require.Contains(t, err.Error(), "db fail")
		require.Nil(t, got)
	})
}
