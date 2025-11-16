//go:generate mockgen --source=usecase.go --destination=mocks/usecase.go -package=mocks

package teams

import (
	"context"
	"errors"
	"fmt"

	repository "github.com/qwerty268/pull_request_service/internal/usecases/teams/storage"
)

var (
	ErrAlreadyExists = errors.New("already exists")
	ErrNotFound      = errors.New("not found")
)

type storage interface {
	AddTeam(team repository.Team) error
	GetTeam(teamName string) (*repository.Team, error)
}

type Usecase struct {
	storage storage
}

func NewUsecase(storage storage) Usecase {
	return Usecase{
		storage: storage,
	}
}

func (u Usecase) AddTeam(_ context.Context, team Team) error {
	err := u.storage.AddTeam(toStorageTeam(team))
	if err != nil {
		if errors.Is(err, repository.ErrAlreadyExists) {
			return fmt.Errorf("failed to add new teram: %w", ErrAlreadyExists)
		}
		return fmt.Errorf("failed to add new teram: %v", err)
	}

	return nil
}

func toStorageTeam(team Team) repository.Team {
	storageTeam := repository.Team{
		TeamName: team.TeamName,
		Members:  make([]repository.TeamMember, len(team.Members)),
	}
	for i, v := range team.Members {
		storageTeam.Members[i] = repository.TeamMember(v)
	}
	return storageTeam
}

func (u Usecase) GetTeam(_ context.Context, teamName string) (*Team, error) {
	storageTeam, err := u.storage.GetTeam(teamName)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, fmt.Errorf("failed to get team: %w", ErrNotFound)
		}
		return nil, fmt.Errorf("failed to get team: %v", err)
	}

	return fromStorageTeam(storageTeam), nil
}

func fromStorageTeam(storageTeam *repository.Team) *Team {
	memers := make([]TeamMember, len(storageTeam.Members))
	for i, v := range storageTeam.Members {
		memers[i] = TeamMember(v)
	}

	return &Team{
		TeamName: storageTeam.TeamName,
		Members:  memers,
	}
}
