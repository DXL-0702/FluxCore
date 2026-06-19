package service

import (
	"errors"
	"fmt"

	"github.com/jaxson/FluxCore/server/model"
	"gorm.io/gorm"
)

var (
	ErrProjectNotFound           = errors.New("project not found")
	ErrProjectNameExists         = errors.New("project name already exists")
	ErrRepositoryNameExists      = errors.New("repository name already exists in project")
	ErrRepositoryRemoteExists    = errors.New("repository remote url already exists in project")
	ErrRepositoryLocalPathExists = errors.New("repository local path already exists")
	ErrRepositoryConflict        = errors.New("repository conflicts with existing repository")
)

type ProjectService struct {
	conn *gorm.DB
}

type CreateProjectInput struct {
	Name        string
	Description string
	Status      model.ProjectStatus
}

type CreateRepositoryInput struct {
	Name          string
	LocalPath     string
	RemoteURL     string
	DefaultBranch string
}

func NewProjectService(conn *gorm.DB) *ProjectService {
	return &ProjectService{conn: conn}
}

func (svc *ProjectService) CreateProject(input CreateProjectInput) (model.Project, error) {
	if err := svc.ensureConnection(); err != nil {
		return model.Project{}, err
	}

	project := model.Project{
		Name:        input.Name,
		Description: input.Description,
		Status:      input.Status,
	}
	if project.Status == "" {
		project.Status = model.ProjectStatusActive
	}

	if err := svc.conn.Create(&project).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return model.Project{}, ErrProjectNameExists
		}
		return model.Project{}, fmt.Errorf("create project: %w", err)
	}

	return project, nil
}

func (svc *ProjectService) ListProjects() ([]model.Project, error) {
	if err := svc.ensureConnection(); err != nil {
		return nil, err
	}

	var projects []model.Project
	if err := svc.conn.Order("id ASC").Find(&projects).Error; err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}

	return projects, nil
}

func (svc *ProjectService) CreateRepository(projectID uint, input CreateRepositoryInput) (model.Repository, error) {
	if err := svc.ensureConnection(); err != nil {
		return model.Repository{}, err
	}
	if err := svc.ensureProjectExists(projectID); err != nil {
		return model.Repository{}, err
	}

	repository := model.Repository{
		ProjectID:     projectID,
		Name:          input.Name,
		LocalPath:     input.LocalPath,
		RemoteURL:     input.RemoteURL,
		DefaultBranch: input.DefaultBranch,
	}
	if repository.DefaultBranch == "" {
		repository.DefaultBranch = "main"
	}

	if err := svc.conn.Create(&repository).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return model.Repository{}, svc.repositoryConflictError(projectID, input)
		}
		return model.Repository{}, fmt.Errorf("create repository: %w", err)
	}

	return repository, nil
}

func (svc *ProjectService) ListRepositories(projectID uint) ([]model.Repository, error) {
	if err := svc.ensureConnection(); err != nil {
		return nil, err
	}
	if err := svc.ensureProjectExists(projectID); err != nil {
		return nil, err
	}

	var repositories []model.Repository
	if err := svc.conn.Where("project_id = ?", projectID).Order("id ASC").Find(&repositories).Error; err != nil {
		return nil, fmt.Errorf("list repositories: %w", err)
	}

	return repositories, nil
}

func (svc *ProjectService) ensureConnection() error {
	if svc == nil || svc.conn == nil {
		return fmt.Errorf("database connection is nil")
	}
	return nil
}

func (svc *ProjectService) ensureProjectExists(projectID uint) error {
	var count int64
	if err := svc.conn.Model(&model.Project{}).Where("id = ?", projectID).Count(&count).Error; err != nil {
		return fmt.Errorf("check project exists: %w", err)
	}
	if count == 0 {
		return ErrProjectNotFound
	}
	return nil
}

func (svc *ProjectService) ensureRepositoryNameAvailable(projectID uint, name string) error {
	var count int64
	if err := svc.conn.Model(&model.Repository{}).Where("project_id = ? AND name = ?", projectID, name).Count(&count).Error; err != nil {
		return fmt.Errorf("check repository name: %w", err)
	}
	if count > 0 {
		return ErrRepositoryNameExists
	}
	return nil
}

func (svc *ProjectService) ensureRepositoryRemoteAvailable(projectID uint, remoteURL string) error {
	var count int64
	if err := svc.conn.Model(&model.Repository{}).Where("project_id = ? AND remote_url = ?", projectID, remoteURL).Count(&count).Error; err != nil {
		return fmt.Errorf("check repository remote url: %w", err)
	}
	if count > 0 {
		return ErrRepositoryRemoteExists
	}
	return nil
}

func (svc *ProjectService) ensureRepositoryLocalPathAvailable(localPath string) error {
	var count int64
	if err := svc.conn.Model(&model.Repository{}).Where("local_path = ?", localPath).Count(&count).Error; err != nil {
		return fmt.Errorf("check repository local path: %w", err)
	}
	if count > 0 {
		return ErrRepositoryLocalPathExists
	}
	return nil
}

func (svc *ProjectService) repositoryConflictError(projectID uint, input CreateRepositoryInput) error {
	for _, check := range []func() error{
		func() error { return svc.ensureRepositoryNameAvailable(projectID, input.Name) },
		func() error { return svc.ensureRepositoryRemoteAvailable(projectID, input.RemoteURL) },
		func() error { return svc.ensureRepositoryLocalPathAvailable(input.LocalPath) },
	} {
		if err := check(); err != nil {
			return err
		}
	}

	return ErrRepositoryConflict
}
