// Package graphql provides GraphQL API for the fern-reporter service
package graphql

import (
	"github.com/guidewire-oss/fern-platform/internal/reporter/service"
)

// Resolver is the root GraphQL resolver
type Resolver struct {
	testRunService *service.TestRunService
	projectService *service.ProjectService
	tagService     *service.TagService
}

// NewResolver creates a new GraphQL resolver
func NewResolver(
	testRunService *service.TestRunService,
	projectService *service.ProjectService,
	tagService *service.TagService,
) *Resolver {
	return &Resolver{
		testRunService: testRunService,
		projectService: projectService,
		tagService:     tagService,
	}
}