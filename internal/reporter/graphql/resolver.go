// Package graphql provides GraphQL API for the fern-reporter service
package graphql

import (
	"github.com/guidewire-oss/fern-platform/internal/reporter/graphql/dataloader"
	testingApp "github.com/guidewire-oss/fern-platform/internal/domains/testing/application"
	projectsApp "github.com/guidewire-oss/fern-platform/internal/domains/projects/application"
	tagsApp "github.com/guidewire-oss/fern-platform/internal/domains/tags/application"
	analyticsApp "github.com/guidewire-oss/fern-platform/internal/domains/analytics/application"
	"github.com/guidewire-oss/fern-platform/pkg/logging"
	"gorm.io/gorm"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

// Resolver is the root GraphQL resolver
type Resolver struct {
	testingService        *testingApp.TestRunService
	projectService        *projectsApp.ProjectService
	tagService           *tagsApp.TagService
	flakyDetectionService *analyticsApp.FlakyDetectionService
	loaders              *dataloader.Loaders
	db                   *gorm.DB
	logger               *logging.Logger
}

// NewResolver creates a new GraphQL resolver
func NewResolver(
	testingService *testingApp.TestRunService,
	projectService *projectsApp.ProjectService,
	tagService *tagsApp.TagService,
	flakyDetectionService *analyticsApp.FlakyDetectionService,
	db *gorm.DB,
	logger *logging.Logger,
) *Resolver {
	return &Resolver{
		testingService:        testingService,
		projectService:        projectService,
		tagService:           tagService,
		flakyDetectionService: flakyDetectionService,
		loaders:              dataloader.NewLoaders(db),
		db:                   db,
		logger:               logger,
	}
}