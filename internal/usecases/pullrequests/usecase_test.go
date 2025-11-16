package pullrequests

import (
	"errors"
	"testing"
	"time"

	"context"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/qwerty268/pull_request_service/internal/usecases/pullrequests/mocks"
	repo "github.com/qwerty268/pull_request_service/internal/usecases/pullrequests/storage"
)

func TestUsecase_CreatePR(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPRStorage := mocks.NewMockprStorage(ctrl)
	mockTeamStorage := mocks.NewMockteamStorage(ctrl)
	usecase := NewUsecase(mockPRStorage, mockTeamStorage, nil)
	ctx := context.Background()

	base := CreatePROpst{
		PullRequestID:   "pr42",
		PullRequestName: "Fix bug",
		AuthorID:        "authorA",
	}

	expectedPrToSave := repo.PullRequest{
		PullRequestID:   base.PullRequestID,
		PullRequestName: base.PullRequestName,
		AuthorID:        base.AuthorID,
		IsMerged:        false,
		CreatedAt:       time.Now(),
	}
	t.Run("user not exists in team", func(t *testing.T) {
		mockTeamStorage.EXPECT().
			CheckUserInCommand(base.AuthorID).
			Return(false, nil)

		pr, err := usecase.CreatePR(ctx, base)
		require.ErrorIs(t, err, ErrNotFound)
		require.Nil(t, pr)
	})

	t.Run("CheckUserInCommand error", func(t *testing.T) {
		mockTeamStorage.EXPECT().
			CheckUserInCommand(base.AuthorID).
			Return(false, errors.New("db check error"))

		pr, err := usecase.CreatePR(ctx, base)
		require.Error(t, err)
		require.Contains(t, err.Error(), "db check error")
		require.Nil(t, pr)
	})

	t.Run("GetUserActiveTeammates error", func(t *testing.T) {
		mockTeamStorage.EXPECT().
			CheckUserInCommand(base.AuthorID).
			Return(true, nil)

		mockTeamStorage.EXPECT().
			GetUserActiveTeammates(base.AuthorID).
			Return(nil, errors.New("teammates err"))
		pr, err := usecase.CreatePR(ctx, base)

		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to get user teammates")
		require.Contains(t, err.Error(), "teammates err")
		require.Nil(t, pr)
	})

	t.Run("teammates <= 2, AddPr success", func(t *testing.T) {
		activeTeammates := []string{"a1", "a2"}

		mockTeamStorage.EXPECT().
			CheckUserInCommand(base.AuthorID).
			Return(true, nil)

		mockTeamStorage.EXPECT().
			GetUserActiveTeammates(base.AuthorID).
			Return(activeTeammates, nil)

		mockPRStorage.EXPECT().
			AddPr(gomock.AssignableToTypeOf(repo.PullRequest{})).
			DoAndReturn(func(actual repo.PullRequest) error {
				require.Equal(t, expectedPrToSave.PullRequestID, actual.PullRequestID)
				require.Equal(t, expectedPrToSave.PullRequestName, actual.PullRequestName)
				require.Equal(t, expectedPrToSave.AuthorID, actual.AuthorID)
				require.Equal(t, expectedPrToSave.IsMerged, actual.IsMerged)
				require.WithinDuration(t, expectedPrToSave.CreatedAt, actual.CreatedAt, 2*time.Second)
				return nil
			})

		pr, err := usecase.CreatePR(ctx, base)
		require.NoError(t, err)
		require.NotNil(t, pr)

		require.ElementsMatch(t, []string{"a1", "a2"}, pr.AssignedReviewers)
		require.Equal(t, base.PullRequestID, pr.PullRequestID)
		require.Equal(t, base.PullRequestName, pr.PullRequestName)
		require.Equal(t, base.AuthorID, pr.AuthorID)
		require.Equal(t, "OPEN", pr.Status)
		require.WithinDuration(t, time.Now(), pr.CreatedAt, time.Second)
	})

	t.Run("teammates > 2, AddPr success, two random", func(t *testing.T) {
		team := []string{"a1", "a2", "a3", "a4"}

		mockTeamStorage.EXPECT().
			CheckUserInCommand(base.AuthorID).
			Return(true, nil)

		mockTeamStorage.EXPECT().
			GetUserActiveTeammates(base.AuthorID).
			Return(team, nil)

		mockPRStorage.EXPECT().
			AddPr(gomock.AssignableToTypeOf(repo.PullRequest{})).
			DoAndReturn(func(actual repo.PullRequest) error {
				require.Equal(t, expectedPrToSave.PullRequestID, actual.PullRequestID)
				require.Equal(t, expectedPrToSave.PullRequestName, actual.PullRequestName)
				require.Equal(t, expectedPrToSave.AuthorID, actual.AuthorID)
				require.Equal(t, expectedPrToSave.IsMerged, actual.IsMerged)
				require.Len(t, actual.AssignedReviewers, 2)
				require.WithinDuration(t, expectedPrToSave.CreatedAt, actual.CreatedAt, 2*time.Second)
				return nil
			})

		pr, err := usecase.CreatePR(ctx, base)
		require.NoError(t, err)
		require.NotNil(t, pr)

		require.Equal(t, base.PullRequestID, pr.PullRequestID)
		require.Equal(t, base.PullRequestName, pr.PullRequestName)
		require.Equal(t, base.AuthorID, pr.AuthorID)
		require.Equal(t, "OPEN", pr.Status)
		require.Len(t, pr.AssignedReviewers, 2)
		require.Subset(t, team, pr.AssignedReviewers)
		require.NotEqual(t, pr.AssignedReviewers[0], pr.AssignedReviewers[1])
		require.WithinDuration(t, time.Now(), pr.CreatedAt, time.Second)
	})

	t.Run("AddPr already exists", func(t *testing.T) {
		mockTeamStorage.EXPECT().
			CheckUserInCommand(base.AuthorID).
			Return(true, nil)
		mockTeamStorage.EXPECT().
			GetUserActiveTeammates(base.AuthorID).
			Return([]string{"a1", "a2"}, nil)

		mockPRStorage.EXPECT().
			AddPr(gomock.AssignableToTypeOf(repo.PullRequest{})).
			DoAndReturn(func(actual repo.PullRequest) error {
				require.Equal(t, expectedPrToSave.PullRequestID, actual.PullRequestID)
				require.Equal(t, expectedPrToSave.PullRequestName, actual.PullRequestName)
				require.Equal(t, expectedPrToSave.AuthorID, actual.AuthorID)
				require.Equal(t, expectedPrToSave.IsMerged, actual.IsMerged)
				require.Len(t, actual.AssignedReviewers, 2)
				require.WithinDuration(t, expectedPrToSave.CreatedAt, actual.CreatedAt, 2*time.Second)
				return repo.ErrAlreadyExists
			})

		pr, err := usecase.CreatePR(ctx, base)
		require.ErrorIs(t, err, ErrAlreadyExists)
		require.Nil(t, pr)
	})

	t.Run("AddPr other error", func(t *testing.T) {
		mockTeamStorage.EXPECT().
			CheckUserInCommand(base.AuthorID).
			Return(true, nil)
		mockTeamStorage.EXPECT().
			GetUserActiveTeammates(base.AuthorID).
			Return([]string{"a1", "a2"}, nil)
		mockPRStorage.EXPECT().
			AddPr(gomock.AssignableToTypeOf(repo.PullRequest{})).
			DoAndReturn(func(actual repo.PullRequest) error {
				require.Equal(t, expectedPrToSave.PullRequestID, actual.PullRequestID)
				require.Equal(t, expectedPrToSave.PullRequestName, actual.PullRequestName)
				require.Equal(t, expectedPrToSave.AuthorID, actual.AuthorID)
				require.Equal(t, expectedPrToSave.IsMerged, actual.IsMerged)
				require.Len(t, actual.AssignedReviewers, 2)
				require.WithinDuration(t, expectedPrToSave.CreatedAt, actual.CreatedAt, 2*time.Second)
				return errors.New("db save error")
			})

		pr, err := usecase.CreatePR(ctx, base)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to save pr")
		require.Contains(t, err.Error(), "db save error")
		require.Nil(t, pr)
	})
}

func TestUsecase_MergePR(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPRStorage := mocks.NewMockprStorage(ctrl)
	uc := NewUsecase(mockPRStorage, nil, nil)
	ctx := context.Background()

	basePR := &repo.PullRequest{
		PullRequestID:     "pr73",
		PullRequestName:   "Refactor",
		AuthorID:          "johnny",
		AssignedReviewers: []string{"alice", "bob"},
		IsMerged:          true,
		CreatedAt:         time.Now().Add(-time.Hour),
		MergedAt:          time.Now(),
	}

	t.Run("ok", func(t *testing.T) {
		mockPRStorage.EXPECT().
			SetPrMerged("pr73").
			Return(basePR, nil)

		pr, err := uc.MergePR(ctx, "pr73")
		require.NoError(t, err)
		require.NotNil(t, pr)
		require.Equal(t, basePR.PullRequestID, pr.PullRequestID)
		require.Equal(t, basePR.PullRequestName, pr.PullRequestName)
		require.Equal(t, basePR.AuthorID, pr.AuthorID)
		require.Equal(t, basePR.AssignedReviewers, pr.AssignedReviewers)
		require.Equal(t, statusMerged, pr.Status)
		require.WithinDuration(t, basePR.MergedAt, pr.MergedAt, time.Second)
		require.WithinDuration(t, basePR.CreatedAt, pr.CreatedAt, time.Second)
	})

	t.Run("not found", func(t *testing.T) {
		mockPRStorage.EXPECT().
			SetPrMerged("pr-notfound").
			Return(nil, ErrNotFound)

		pr, err := uc.MergePR(ctx, "pr-notfound")
		require.ErrorIs(t, err, ErrNotFound)
		require.Contains(t, err.Error(), "failed to set merged flag")
		require.Nil(t, pr)
	})

	t.Run("storage error", func(t *testing.T) {
		mockPRStorage.EXPECT().
			SetPrMerged("pr73").
			Return(nil, errors.New("unexpected error"))

		pr, err := uc.MergePR(ctx, "pr73")
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to set merged flag")
		require.Contains(t, err.Error(), "unexpected error")
		require.Nil(t, pr)
	})

	t.Run("status open", func(t *testing.T) {
		basePROpen := *basePR
		basePROpen.IsMerged = false

		mockPRStorage.EXPECT().
			SetPrMerged("pr-open").
			Return(&basePROpen, nil)

		pr, err := uc.MergePR(ctx, "pr-open")
		require.NoError(t, err)
		require.NotNil(t, pr)
		require.Equal(t, statusOpen, pr.Status)
	})
}

func TestUsecase_ReassignReviewer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPRStorage := mocks.NewMockprStorage(ctrl)
	mockUserStorage := mocks.NewMockuserStorage(ctrl)
	mockTeamStorage := mocks.NewMockteamStorage(ctrl)

	usecase := NewUsecase(mockPRStorage, mockTeamStorage, mockUserStorage)

	ctx := context.Background()
	prID := "pr1"
	oldUserID := "bob"
	now := time.Now()
	storagePr := &repo.PullRequest{
		PullRequestID:     prID,
		PullRequestName:   "Test PR",
		AuthorID:          "author",
		IsMerged:          false,
		AssignedReviewers: []string{"alice", "bob"}, // 2 ревьюера
		CreatedAt:         now,
		MergedAt:          time.Time{},
	}

	t.Run("PR not found", func(t *testing.T) {
		mockPRStorage.EXPECT().GetPrByID(prID).
			Return(nil, repo.ErrNotFound)
		res, err := usecase.ReassignReviewer(ctx, prID, oldUserID)
		require.ErrorIs(t, err, ErrNotFound)
		require.Nil(t, res)
	})

	t.Run("PR is merged", func(t *testing.T) {
		mergedPr := *storagePr
		mergedPr.IsMerged = true
		mockPRStorage.EXPECT().GetPrByID(prID).
			Return(&mergedPr, nil)
		res, err := usecase.ReassignReviewer(ctx, prID, oldUserID)
		require.ErrorIs(t, err, ErrPRMerged)
		require.Nil(t, res)
	})

	t.Run("user not found", func(t *testing.T) {
		mockPRStorage.EXPECT().GetPrByID(prID).
			Return(storagePr, nil)
		mockUserStorage.EXPECT().CheckUserExists(oldUserID).
			Return(false, nil)
		res, err := usecase.ReassignReviewer(ctx, prID, oldUserID)
		require.ErrorIs(t, err, ErrNotFound)
		require.Nil(t, res)
	})

	t.Run("CheckUserExists error", func(t *testing.T) {
		mockPRStorage.EXPECT().GetPrByID(prID).
			Return(storagePr, nil)
		mockUserStorage.EXPECT().CheckUserExists(oldUserID).
			Return(false, errors.New("db fail"))
		res, err := usecase.ReassignReviewer(ctx, prID, oldUserID)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to check user exists")
		require.Nil(t, res)
	})

	t.Run("not assigned to PR", func(t *testing.T) {
		mockPRStorage.EXPECT().GetPrByID(prID).
			Return(storagePr, nil)
		mockUserStorage.EXPECT().CheckUserExists(oldUserID).
			Return(true, nil)
		mockPRStorage.EXPECT().CheckUserInPr(prID, oldUserID).
			Return(false, nil)
		res, err := usecase.ReassignReviewer(ctx, prID, oldUserID)
		require.ErrorIs(t, err, ErrNotAssigned)
		require.Nil(t, res)
	})

	t.Run("CheckUserInPr error", func(t *testing.T) {
		mockPRStorage.EXPECT().GetPrByID(prID).
			Return(storagePr, nil)
		mockUserStorage.EXPECT().CheckUserExists(oldUserID).
			Return(true, nil)
		mockPRStorage.EXPECT().CheckUserInPr(prID, oldUserID).
			Return(false, errors.New("db error"))
		res, err := usecase.ReassignReviewer(ctx, prID, oldUserID)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to check user in pr")
		require.Nil(t, res)
	})

	t.Run("no candidates", func(t *testing.T) {
		mockPRStorage.EXPECT().GetPrByID(prID).
			Return(storagePr, nil)
		mockUserStorage.EXPECT().CheckUserExists(oldUserID).
			Return(true, nil)
		mockPRStorage.EXPECT().CheckUserInPr(prID, oldUserID).
			Return(true, nil)
		mockTeamStorage.EXPECT().GetUserActiveTeammates(oldUserID).
			Return([]string{}, nil)
		res, err := usecase.ReassignReviewer(ctx, prID, oldUserID)
		require.ErrorIs(t, err, ErrNoCandidate)
		require.Nil(t, res)
	})

	t.Run("team storage error", func(t *testing.T) {
		mockPRStorage.EXPECT().GetPrByID(prID).
			Return(storagePr, nil)
		mockUserStorage.EXPECT().CheckUserExists(oldUserID).
			Return(true, nil)
		mockPRStorage.EXPECT().CheckUserInPr(prID, oldUserID).
			Return(true, nil)
		mockTeamStorage.EXPECT().GetUserActiveTeammates(oldUserID).
			Return(nil, errors.New("team storage error"))
		res, err := usecase.ReassignReviewer(ctx, prID, oldUserID)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to get active teammates")
		require.Nil(t, res)
	})

	t.Run("success (two reviewers)", func(t *testing.T) {
		mockPRStorage.EXPECT().GetPrByID(prID).
			Return(storagePr, nil)
		mockUserStorage.EXPECT().CheckUserExists(oldUserID).
			Return(true, nil)
		mockPRStorage.EXPECT().CheckUserInPr(prID, oldUserID).
			Return(true, nil)

		mockTeamStorage.EXPECT().GetUserActiveTeammates(oldUserID).
			Return([]string{"alice", "carl"}, nil)
		mockPRStorage.EXPECT().ResetPrMember(repo.ResetReviewerFilter{
			PrID:      prID,
			OldUserID: oldUserID,
			NewUserID: "carl",
		}).Return(nil)

		orig := GetRandomReviewer
		GetRandomReviewer = func([]string) string { return "carl" }
		defer func() { GetRandomReviewer = orig }()

		res, err := usecase.ReassignReviewer(ctx, prID, oldUserID)
		require.NoError(t, err)
		require.NotNil(t, res)
		require.Equal(t, "carl", res.NewReviewer)
		require.ElementsMatch(t, []string{"alice", "carl"}, res.Pr.AssignedReviewers)
	})

	t.Run("success with one reviewers", func(t *testing.T) {
		prOneReviewer := *storagePr
		prOneReviewer.AssignedReviewers = []string{"bob"}
		mockPRStorage.EXPECT().GetPrByID(prID).
			Return(&prOneReviewer, nil)
		mockUserStorage.EXPECT().CheckUserExists(oldUserID).
			Return(true, nil)
		mockPRStorage.EXPECT().CheckUserInPr(prID, oldUserID).
			Return(true, nil)
		mockTeamStorage.EXPECT().GetUserActiveTeammates(oldUserID).
			Return([]string{"alice"}, nil)
		mockPRStorage.EXPECT().ResetPrMember(repo.ResetReviewerFilter{
			PrID:      prID,
			OldUserID: oldUserID,
			NewUserID: "alice",
		}).Return(nil)

		orig := GetRandomReviewer
		GetRandomReviewer = func([]string) string { return "alice" }
		defer func() { GetRandomReviewer = orig }()

		res, err := usecase.ReassignReviewer(ctx, prID, oldUserID)
		require.NoError(t, err)
		require.NotNil(t, res)
		require.Equal(t, "alice", res.NewReviewer)
		require.ElementsMatch(t, []string{"alice"}, res.Pr.AssignedReviewers)
	})

	t.Run("reset error", func(t *testing.T) {
		mockPRStorage.EXPECT().GetPrByID(prID).
			Return(storagePr, nil)
		mockUserStorage.EXPECT().CheckUserExists(oldUserID).
			Return(true, nil)
		mockPRStorage.EXPECT().CheckUserInPr(prID, oldUserID).
			Return(true, nil)
		mockTeamStorage.EXPECT().GetUserActiveTeammates(oldUserID).
			Return([]string{"alice", "carl"}, nil)
		mockPRStorage.EXPECT().ResetPrMember(repo.ResetReviewerFilter{
			PrID:      prID,
			OldUserID: oldUserID,
			NewUserID: "carl",
		}).Return(errors.New("reset failed"))

		orig := GetRandomReviewer
		GetRandomReviewer = func([]string) string { return "carl" }
		defer func() { GetRandomReviewer = orig }()

		res, err := usecase.ReassignReviewer(ctx, prID, oldUserID)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed tu reset pr member")
		require.Nil(t, res)
	})
}
