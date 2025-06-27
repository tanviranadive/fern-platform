// Package graphql provides GraphQL API for the fern-reporter service
package graphql

import (
	"github.com/guidewire-oss/fern-platform/internal/reporter/graphql/dataloader"
	svc "github.com/guidewire-oss/fern-platform/internal/service"
	// "github.com/guidewire-oss/fern-platform/pkg/database"
	"github.com/guidewire-oss/fern-platform/pkg/logging"
	"gorm.io/gorm"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

// Resolver is the root GraphQL resolver
type Resolver struct {
	testRunService svc.TestRunService
	projectService svc.ProjectService
	tagService     svc.TagService
	loaders        *dataloader.Loaders
	db             *gorm.DB
	logger         *logging.Logger
}

// NewResolver creates a new GraphQL resolver
func NewResolver(
	testRunService svc.TestRunService,
	projectService svc.ProjectService,
	tagService svc.TagService,
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