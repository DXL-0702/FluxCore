package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/jaxson/FluxCore/server/model"
	"github.com/jaxson/FluxCore/server/service"
)

const maxJSONBodyBytes = 1 << 20

type projectService interface {
	CreateProject(service.CreateProjectInput) (model.Project, error)
	ListProjects() ([]model.Project, error)
	CreateRepository(uint, service.CreateRepositoryInput) (model.Repository, error)
	ListRepositories(uint) ([]model.Repository, error)
}

type createProjectRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

type createRepositoryRequest struct {
	Name          string `json:"name"`
	LocalPath     string `json:"local_path"`
	RemoteURL     string `json:"remote_url"`
	DefaultBranch string `json:"default_branch"`
}

func registerProjectRoutes(group *gin.RouterGroup, projects projectService) {
	group.POST("/projects", createProjectHandler(projects))
	group.GET("/projects", listProjectsHandler(projects))
	group.POST("/projects/:project_id/repositories", createRepositoryHandler(projects))
	group.GET("/projects/:project_id/repositories", listRepositoriesHandler(projects))
}

func createProjectHandler(projects projectService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var request createProjectRequest
		if !decodeJSONBody(ctx, &request) {
			return
		}

		input, ok := validateCreateProjectRequest(ctx, request)
		if !ok {
			return
		}

		project, err := projects.CreateProject(input)
		if err != nil {
			writeServiceError(ctx, err)
			return
		}

		ctx.JSON(http.StatusCreated, gin.H{"project": project})
	}
}

func listProjectsHandler(projects projectService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		items, err := projects.ListProjects()
		if err != nil {
			writeServiceError(ctx, err)
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"projects": items})
	}
}

func createRepositoryHandler(projects projectService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		projectID, ok := projectIDFromPath(ctx)
		if !ok {
			return
		}

		var request createRepositoryRequest
		if !decodeJSONBody(ctx, &request) {
			return
		}

		input, ok := validateCreateRepositoryRequest(ctx, request)
		if !ok {
			return
		}

		repository, err := projects.CreateRepository(projectID, input)
		if err != nil {
			writeServiceError(ctx, err)
			return
		}

		ctx.JSON(http.StatusCreated, gin.H{"repository": repository})
	}
}

func listRepositoriesHandler(projects projectService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		projectID, ok := projectIDFromPath(ctx)
		if !ok {
			return
		}

		items, err := projects.ListRepositories(projectID)
		if err != nil {
			writeServiceError(ctx, err)
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"repositories": items})
	}
}

func decodeJSONBody(ctx *gin.Context, destination interface{}) bool {
	ctx.Request.Body = http.MaxBytesReader(ctx.Writer, ctx.Request.Body, maxJSONBodyBytes)

	decoder := json.NewDecoder(ctx.Request.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(destination); err != nil {
		writeJSONDecodeError(ctx, err)
		return false
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		writeAPIError(ctx, http.StatusBadRequest, "invalid_json", "request body must contain a single JSON object")
		return false
	}

	return true
}

func writeJSONDecodeError(ctx *gin.Context, err error) {
	var syntaxError *json.SyntaxError
	var unmarshalTypeError *json.UnmarshalTypeError
	var maxBytesError *http.MaxBytesError

	switch {
	case errors.Is(err, io.EOF):
		writeAPIError(ctx, http.StatusBadRequest, "invalid_json", "request body is required")
	case errors.As(err, &maxBytesError):
		writeAPIError(ctx, http.StatusRequestEntityTooLarge, "request_too_large", "request body exceeds the maximum allowed size")
	case errors.As(err, &syntaxError):
		writeAPIError(ctx, http.StatusBadRequest, "invalid_json", "request body contains malformed JSON")
	case errors.As(err, &unmarshalTypeError):
		writeAPIError(ctx, http.StatusBadRequest, "invalid_json", fmt.Sprintf("field %q has an invalid type", unmarshalTypeError.Field))
	case strings.HasPrefix(err.Error(), "json: unknown field "):
		writeAPIError(ctx, http.StatusBadRequest, "invalid_json", "request body contains an unknown field")
	default:
		writeAPIError(ctx, http.StatusBadRequest, "invalid_json", "request body is invalid")
	}
}

func validateCreateProjectRequest(ctx *gin.Context, request createProjectRequest) (service.CreateProjectInput, bool) {
	name, ok := requiredString(ctx, request.Name, "name", 120)
	if !ok {
		return service.CreateProjectInput{}, false
	}
	description, ok := optionalString(ctx, request.Description, "description", 500)
	if !ok {
		return service.CreateProjectInput{}, false
	}
	status := strings.TrimSpace(request.Status)
	if status == "" {
		status = string(model.ProjectStatusActive)
	}
	if status != string(model.ProjectStatusActive) {
		writeAPIError(ctx, http.StatusBadRequest, "invalid_request", "status must be active")
		return service.CreateProjectInput{}, false
	}

	return service.CreateProjectInput{
		Name:        name,
		Description: description,
		Status:      model.ProjectStatus(status),
	}, true
}

func validateCreateRepositoryRequest(ctx *gin.Context, request createRepositoryRequest) (service.CreateRepositoryInput, bool) {
	name, ok := requiredString(ctx, request.Name, "name", 120)
	if !ok {
		return service.CreateRepositoryInput{}, false
	}
	localPath, ok := requiredString(ctx, request.LocalPath, "local_path", 500)
	if !ok {
		return service.CreateRepositoryInput{}, false
	}
	remoteURL, ok := requiredString(ctx, request.RemoteURL, "remote_url", 500)
	if !ok {
		return service.CreateRepositoryInput{}, false
	}
	defaultBranch, ok := optionalString(ctx, request.DefaultBranch, "default_branch", 120)
	if !ok {
		return service.CreateRepositoryInput{}, false
	}
	if defaultBranch == "" {
		defaultBranch = "main"
	}

	return service.CreateRepositoryInput{
		Name:          name,
		LocalPath:     localPath,
		RemoteURL:     remoteURL,
		DefaultBranch: defaultBranch,
	}, true
}

func requiredString(ctx *gin.Context, value string, field string, maxLength int) (string, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		writeAPIError(ctx, http.StatusBadRequest, "invalid_request", field+" is required")
		return "", false
	}
	return validateStringLength(ctx, value, field, maxLength)
}

func optionalString(ctx *gin.Context, value string, field string, maxLength int) (string, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", true
	}
	return validateStringLength(ctx, value, field, maxLength)
}

func validateStringLength(ctx *gin.Context, value string, field string, maxLength int) (string, bool) {
	if utf8.RuneCountInString(value) > maxLength {
		writeAPIError(ctx, http.StatusBadRequest, "invalid_request", fmt.Sprintf("%s must be at most %d characters", field, maxLength))
		return "", false
	}
	return value, true
}

func projectIDFromPath(ctx *gin.Context) (uint, bool) {
	rawID := ctx.Param("project_id")
	id, err := strconv.ParseUint(rawID, 10, 64)
	if err != nil || id == 0 || id > uint64(math.MaxUint) {
		writeAPIError(ctx, http.StatusBadRequest, "invalid_request", "project_id must be a positive integer")
		return 0, false
	}

	return uint(id), true
}

func writeServiceError(ctx *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrProjectNotFound):
		writeAPIError(ctx, http.StatusNotFound, "not_found", "project not found")
	case errors.Is(err, service.ErrProjectNameExists):
		writeAPIError(ctx, http.StatusConflict, "conflict", "project name already exists")
	case errors.Is(err, service.ErrRepositoryNameExists):
		writeAPIError(ctx, http.StatusConflict, "conflict", "repository name already exists in project")
	case errors.Is(err, service.ErrRepositoryRemoteExists):
		writeAPIError(ctx, http.StatusConflict, "conflict", "repository remote_url already exists in project")
	case errors.Is(err, service.ErrRepositoryLocalPathExists):
		writeAPIError(ctx, http.StatusConflict, "conflict", "repository local_path already exists")
	case errors.Is(err, service.ErrRepositoryConflict):
		writeAPIError(ctx, http.StatusConflict, "conflict", "repository conflicts with an existing repository")
	default:
		writeAPIError(ctx, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}
