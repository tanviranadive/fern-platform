package domains

import (
	"gorm.io/gorm"

	// Testing domain
	testingInterfaces "github.com/guidewire-oss/fern-platform/internal/domains/testing/interfaces"

	// Projects domain
	projectsInterfaces "github.com/guidewire-oss/fern-platform/internal/domains/projects/interfaces"

	// Tags domain
	tagsInterfaces "github.com/guidewire-oss/fern-platform/internal/domains/tags/interfaces"

	// Legacy repositories
	"github.com/guidewire-oss/fern-platform/internal/reporter/repository"
	"github.com/guidewire-oss/fern-platform/pkg/logging"
)

// DomainFactory creates and wires all domain components
type DomainFactory struct {
	db     *gorm.DB
	logger *logging.Logger
	
	// Testing domain
	testingAdapter         *testingInterfaces.TestRunServiceAdapter
	
	// Projects domain
	projectAdapter         *projectsInterfaces.ProjectServiceAdapter
	
	// Tags domain
	tagAdapter             *tagsInterfaces.TagServiceAdapter
	
	// Legacy repositories (for backward compatibility)
	legacyTestRunRepo    *repository.TestRunRepository
	legacySuiteRunRepo   *repository.SuiteRunRepository
	legacySpecRunRepo    *repository.SpecRunRepository
	legacyProjectRepo    *repository.ProjectRepository
	legacyTagRepo        *repository.TagRepository
}

// NewDomainFactory creates a new domain factory
func NewDomainFactory(db *gorm.DB, logger *logging.Logger) *DomainFactory {
	factory := &DomainFactory{
		db:     db,
		logger: logger,
	}
	
	// Initialize legacy repositories
	factory.initLegacyRepositories()
	
	// Initialize Testing domain
	factory.initTestingDomain()
	
	// Initialize Projects domain
	factory.initProjectsDomain()
	
	// Initialize Tags domain
	factory.initTagsDomain()
	
	return factory
}

// initLegacyRepositories initializes the legacy repositories for backward compatibility
func (f *DomainFactory) initLegacyRepositories() {
	f.legacyTestRunRepo = repository.NewTestRunRepository(f.db)
	f.legacySuiteRunRepo = repository.NewSuiteRunRepository(f.db)
	f.legacySpecRunRepo = repository.NewSpecRunRepository(f.db)
	f.legacyProjectRepo = repository.NewProjectRepository(f.db)
	f.legacyTagRepo = repository.NewTagRepository(f.db)
}

// initTestingDomain initializes the testing domain components
func (f *DomainFactory) initTestingDomain() {
	// TODO: Implement domain repositories and handlers once we're ready to fully migrate
	// For now, we're using legacy repositories directly in the adapters
	
	// Create repositories (commented out until fully implemented)
	// f.testRunRepo = testingInfra.NewGormTestRunRepository(f.db)
	// f.flakyTestRepo = testingInfra.NewGormFlakyTestRepository(f.db)
	
	// Create application handlers (commented out until repositories are ready)
	// f.recordTestRunHandler = testingApp.NewRecordTestRunHandler(f.testRunRepo)
	// f.completeTestRunHandler = testingApp.NewCompleteTestRunHandler(f.testRunRepo, f.flakyTestRepo)
	
	// Create service adapter using legacy repositories
	f.testingAdapter = testingInterfaces.NewTestRunServiceAdapter(
		nil, // recordTestRunHandler - not used yet
		nil, // completeTestRunHandler - not used yet
		nil, // testRunRepo - not used yet
		f.legacyTestRunRepo,
		f.legacySuiteRunRepo,
		f.legacySpecRunRepo,
		f.logger,
	)
}

// initProjectsDomain initializes the projects domain components
func (f *DomainFactory) initProjectsDomain() {
	// TODO: Implement domain repositories and handlers once we're ready to fully migrate
	// For now, we're using legacy repositories directly in the adapters
	
	// Create repositories (commented out until fully implemented)
	// f.projectRepo = projectsInfra.NewGormProjectRepository(f.db)
	// f.projectPermissionRepo = projectsInfra.NewGormProjectPermissionRepository(f.db)
	
	// Create application handlers (commented out until repositories are ready)
	// f.createProjectHandler = projectsApp.NewCreateProjectHandler(f.projectRepo, f.projectPermissionRepo)
	// f.updateProjectHandler = projectsApp.NewUpdateProjectHandler(f.projectRepo, f.projectPermissionRepo)
	
	// Create service adapter using legacy repositories
	f.projectAdapter = projectsInterfaces.NewProjectServiceAdapter(
		nil, // createProjectHandler - not used yet
		nil, // updateProjectHandler - not used yet
		nil, // projectRepo - not used yet
		f.legacyProjectRepo,
		f.logger,
	)
}

// GetTestRunService returns the test run service that maintains backward compatibility
func (f *DomainFactory) GetTestRunService() *testingInterfaces.TestRunServiceAdapter {
	return f.testingAdapter
}

// GetProjectService returns the project service that maintains backward compatibility
func (f *DomainFactory) GetProjectService() *projectsInterfaces.ProjectServiceAdapter {
	return f.projectAdapter
}


// initTagsDomain initializes the tags domain components
func (f *DomainFactory) initTagsDomain() {
	// TODO: Implement domain repositories and handlers once we're ready to fully migrate
	// For now, we're using legacy repositories directly in the adapters
	
	// Create repositories (commented out until fully implemented)
	// f.tagRepo = tagsInfra.NewGormTagRepository(f.db)
	
	// Create application handlers (commented out until repositories are ready)
	// f.createTagHandler = tagsApp.NewCreateTagHandler(f.tagRepo)
	// f.assignTagsHandler = tagsApp.NewAssignTagsHandler(f.tagRepo)
	
	// Create service adapter using legacy repositories
	f.tagAdapter = tagsInterfaces.NewTagServiceAdapter(
		nil, // createTagHandler - not used yet
		nil, // assignTagsHandler - not used yet
		nil, // tagRepo - not used yet
		f.legacyTagRepo,
		f.logger,
	)
}

// GetTagService returns the tag service that maintains backward compatibility
func (f *DomainFactory) GetTagService() *tagsInterfaces.TagServiceAdapter {
	return f.tagAdapter
}