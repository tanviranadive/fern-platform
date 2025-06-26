// Package graphql provides GraphQL API for the fern-reporter service
package graphql

import (
	"github.com/guidewire-oss/fern-platform/internal/reporter/graphql/dataloader"
	"github.com/guidewire-oss/fern-platform/internal/reporter/service"
	// "github.com/guidewire-oss/fern-platform/pkg/database"
	"github.com/guidewire-oss/fern-platform/pkg/logging"
	"gorm.io/gorm"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

// Resolver is the root GraphQL resolver
type Resolver struct {
	testRunService *service.TestRunService
	projectService *service.ProjectService
	tagService     *service.TagService
	loaders        *dataloader.Loaders
	db             *gorm.DB
	logger         *logging.Logger
}

// NewResolver creates a new GraphQL resolver
func NewResolver(
	testRunService *service.TestRunService,
	projectService *service.ProjectService,
	tagService *service.TagService,
	db *gorm.DB,
	logger *logging.Logger,
) *Resolver {
	return &Resolver{
		testRunService: testRunService,
		projectService: projectService,
		tagService:     tagService,
		loaders:        dataloader.NewLoaders(db),
		db:             db,
		logger:         logger,
	}
}