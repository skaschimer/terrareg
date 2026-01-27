package service

import (
	"github.com/rs/zerolog"

	configModel "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/config/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
)

// ModuleDetailsWithID wraps ModuleDetails with a database ID
type ModuleDetailsWithID struct {
	*model.ModuleDetails
	ID int
}

// ProcessedSubmoduleInfo represents a processed submodule with detailed information
type ProcessedSubmoduleInfo struct {
	Path          string         `json:"path"`
	Source        string         `json:"source"`
	Version       string         `json:"version"`
	Description   string         `json:"description"`
	ReadmeContent string         `json:"readme_content,omitempty"`
	Variables     []VariableInfo `json:"variables,omitempty"`
	Outputs       []OutputInfo   `json:"outputs,omitempty"`
}

// ProcessedExampleFile represents a file within an example directory
type ProcessedExampleFile struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

// ModuleProcessorServiceImpl implements the ModuleProcessorService interface
type ModuleProcessorServiceImpl struct {
	moduleParser      ModuleParser
	moduleDetailsRepo repository.ModuleDetailsRepository
	moduleVersionRepo repository.ModuleVersionRepository
	submoduleRepo     repository.SubmoduleRepository
	exampleFileRepo   repository.ExampleFileRepository
	config            *configModel.DomainConfig
	logger            zerolog.Logger
}
