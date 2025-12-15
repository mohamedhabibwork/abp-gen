package detector

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// LayerPaths contains the paths to various ABP layers
type LayerPaths struct {
	Domain               string
	DomainShared         string
	ApplicationContracts string
	Application          string
	HttpApi              string
	EntityFrameworkCore  string
	MongoDB              string

	// Subdirectories within layers
	DomainEntities           string
	DomainRepositories       string
	DomainManagers           string
	DomainData               string
	DomainSharedConstants    string
	DomainSharedEvents       string
	DomainSharedEnums        string
	DomainSharedLocalization string
	ContractsPermissions     string
	ContractsDTOs            string
	ContractsServices        string
	ApplicationServices      string
	ApplicationAutoMapper    string
	ApplicationValidators    string
	ApplicationEventHandlers string
	HttpApiControllers       string
	EFCoreConfigurations     string
	EFCoreRepositories       string
	MongoDBRepositories      string
}

// DetectLayerPaths detects and returns paths to all ABP layers
func DetectLayerPaths(solutionInfo *SolutionInfo, moduleName string) (*LayerPaths, error) {
	paths := &LayerPaths{}

	// Get base project paths
	if domain := solutionInfo.GetProject(ProjectTypeDomain); domain != nil {
		paths.Domain = domain.Directory
		paths.DomainEntities = filepath.Join(domain.Directory, "Entities")
		paths.DomainRepositories = filepath.Join(domain.Directory, "Repositories")
		paths.DomainManagers = filepath.Join(domain.Directory, "Managers")
		paths.DomainData = filepath.Join(domain.Directory, "Data")
	}

	if domainShared := solutionInfo.GetProject(ProjectTypeDomainShared); domainShared != nil {
		paths.DomainShared = domainShared.Directory
		paths.DomainSharedConstants = filepath.Join(domainShared.Directory, "Constants")
		paths.DomainSharedEvents = filepath.Join(domainShared.Directory, "Events")
		paths.DomainSharedEnums = filepath.Join(domainShared.Directory, "Enums")
		paths.DomainSharedLocalization = filepath.Join(domainShared.Directory, "Localization", moduleName)
	}

	if appContracts := solutionInfo.GetProject(ProjectTypeApplicationContracts); appContracts != nil {
		paths.ApplicationContracts = appContracts.Directory
		paths.ContractsPermissions = filepath.Join(appContracts.Directory, "Permissions")
		paths.ContractsDTOs = appContracts.Directory // DTOs are organized by entity
		paths.ContractsServices = filepath.Join(appContracts.Directory, "Services")
	}

	if app := solutionInfo.GetProject(ProjectTypeApplication); app != nil {
		paths.Application = app.Directory
		paths.ApplicationServices = filepath.Join(app.Directory, "Services")
		paths.ApplicationAutoMapper = filepath.Join(app.Directory, "AutoMapper")
		paths.ApplicationValidators = filepath.Join(app.Directory, "Validators")
		paths.ApplicationEventHandlers = filepath.Join(app.Directory, "EventHandlers")
	}

	if httpApi := solutionInfo.GetProject(ProjectTypeHttpApi); httpApi != nil {
		paths.HttpApi = httpApi.Directory
		paths.HttpApiControllers = filepath.Join(httpApi.Directory, "Controllers")
	}

	if efCore := solutionInfo.GetProject(ProjectTypeEntityFrameworkCore); efCore != nil {
		paths.EntityFrameworkCore = efCore.Directory
		paths.EFCoreConfigurations = filepath.Join(efCore.Directory, "EntityFrameworkCore", "Configurations")
		paths.EFCoreRepositories = filepath.Join(efCore.Directory, "EntityFrameworkCore", "Repositories")
	}

	if mongodb := solutionInfo.GetProject(ProjectTypeMongoDB); mongodb != nil {
		paths.MongoDB = mongodb.Directory
		paths.MongoDBRepositories = filepath.Join(mongodb.Directory, "MongoDB", "Repositories")
	}

	// Validate required paths exist
	if paths.Domain == "" {
		// Build a helpful error message listing detected projects
		var detectedProjects []string
		var unknownProjects []string
		for _, project := range solutionInfo.Projects {
			if project.Type != ProjectTypeUnknown {
				detectedProjects = append(detectedProjects, fmt.Sprintf("%s (%s)", project.Name, project.Type))
			} else {
				unknownProjects = append(unknownProjects, project.Name)
			}
		}

		errMsg := "Domain project not found in solution"
		if len(detectedProjects) > 0 {
			errMsg += fmt.Sprintf("\nDetected projects: %s", strings.Join(detectedProjects, ", "))
		}
		if len(unknownProjects) > 0 {
			errMsg += fmt.Sprintf("\nUnknown projects (not recognized as ABP layers): %s", strings.Join(unknownProjects, ", "))
		}
		errMsg += "\n\nExpected project naming patterns:"
		errMsg += "\n  - Domain: '*.Domain' or 'Domain'"
		errMsg += "\n  - Domain.Shared: '*.Domain.Shared' or 'Domain.Shared'"
		errMsg += "\n  - Application: '*.Application' or 'Application'"
		errMsg += "\n  - Application.Contracts: '*.Application.Contracts' or 'Application.Contracts'"
		errMsg += "\n  - HttpApi: '*.HttpApi' or 'HttpApi'"

		return nil, errors.New(errMsg)
	}

	return paths, nil
}

// EnsureDirectories creates all necessary directories
func (p *LayerPaths) EnsureDirectories() error {
	directories := []string{
		p.DomainEntities,
		p.DomainRepositories,
		p.DomainManagers,
		p.DomainData,
		p.DomainSharedConstants,
		p.DomainSharedEvents,
		p.DomainSharedEnums,
		p.DomainSharedLocalization,
		p.ContractsPermissions,
		p.ContractsServices,
		p.ApplicationServices,
		p.ApplicationAutoMapper,
		p.ApplicationValidators,
		p.ApplicationEventHandlers,
		p.HttpApiControllers,
		p.EFCoreConfigurations,
		p.EFCoreRepositories,
		p.MongoDBRepositories,
	}

	for _, dir := range directories {
		if dir == "" {
			continue
		}
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

// EnsureModuleDirectories creates all necessary module-specific directories
// moduleFolder should be the full module folder name (e.g., "ProductModule", "IntegrationServiceModule")
func (p *LayerPaths) EnsureModuleDirectories(moduleFolder string) error {
	directories := []string{
		filepath.Join(p.DomainEntities, moduleFolder),
		filepath.Join(p.DomainRepositories, moduleFolder),
		filepath.Join(p.DomainManagers, moduleFolder),
		filepath.Join(p.DomainData, moduleFolder),
		filepath.Join(p.DomainSharedConstants, moduleFolder),
		filepath.Join(p.DomainSharedEvents, moduleFolder),
		filepath.Join(p.ContractsPermissions, moduleFolder),
		filepath.Join(p.ContractsServices, moduleFolder),
		filepath.Join(p.ApplicationServices, moduleFolder),
		filepath.Join(p.ApplicationAutoMapper, moduleFolder),
		filepath.Join(p.ApplicationValidators, moduleFolder),
		filepath.Join(p.ApplicationEventHandlers, moduleFolder),
		filepath.Join(p.HttpApiControllers, moduleFolder),
		filepath.Join(p.EFCoreConfigurations, moduleFolder),
		filepath.Join(p.EFCoreRepositories, moduleFolder),
		filepath.Join(p.MongoDBRepositories, moduleFolder),
	}

	for _, dir := range directories {
		if dir == "" {
			continue
		}
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create module directory %s: %w", dir, err)
		}
	}

	return nil
}

// GetEntityDTOPath returns the path for DTOs of a specific entity
func (p *LayerPaths) GetEntityDTOPath(moduleName, entityName string) string {
	if p.ContractsDTOs == "" {
		return ""
	}
	return filepath.Join(p.ContractsDTOs, moduleName, entityName)
}

// GetDbContextPath returns the path to the DbContext file
func (p *LayerPaths) GetDbContextPath(serviceName string) string {
	if p.EntityFrameworkCore == "" {
		return ""
	}
	return filepath.Join(p.EntityFrameworkCore, "EntityFrameworkCore", serviceName+"DbContext.cs")
}

// GetIDbContextPath returns the path to the IDbContext file
func (p *LayerPaths) GetIDbContextPath(serviceName string) string {
	if p.EntityFrameworkCore == "" {
		return ""
	}
	return filepath.Join(p.EntityFrameworkCore, "EntityFrameworkCore", "I"+serviceName+"DbContext.cs")
}

// GetPermissionsFilePath returns the path to the permissions file
// moduleFolder should be the full module folder name (e.g., "ProductModule", "IntegrationServiceModule")
// moduleName should be the base module name (e.g., "Product", "IntegrationService")
func (p *LayerPaths) GetPermissionsFilePath(moduleFolder, moduleName string) string {
	if p.ContractsPermissions == "" {
		return ""
	}
	return filepath.Join(p.ContractsPermissions, moduleFolder, moduleName+"Permissions.cs")
}

// GetPermissionProviderPath returns the path to the permission provider file
// moduleFolder should be the full module folder name (e.g., "ProductModule", "IntegrationServiceModule")
// moduleName should be the base module name (e.g., "Product", "IntegrationService")
func (p *LayerPaths) GetPermissionProviderPath(moduleFolder, moduleName string) string {
	if p.ContractsPermissions == "" {
		return ""
	}
	return filepath.Join(p.ContractsPermissions, moduleFolder, moduleName+"PermissionDefinitionProvider.cs")
}
