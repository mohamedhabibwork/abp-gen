package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mohamedhabibwork/abp-gen/internal/detector"
	"github.com/mohamedhabibwork/abp-gen/internal/generator"
	"github.com/mohamedhabibwork/abp-gen/internal/prompts"
	"github.com/mohamedhabibwork/abp-gen/internal/schema"
	"github.com/mohamedhabibwork/abp-gen/internal/templates"
	"github.com/mohamedhabibwork/abp-gen/internal/writer"
	"github.com/spf13/cobra"
)

var (
	// Version information (set via build flags)
	Version   = "dev"
	BuildDate = "unknown"
	GitCommit = "unknown"
)

var (
	// Global flags
	verbose bool

	// Generate command flags
	inputFile       string
	solutionPath    string
	moduleName      string
	templatesPath   string
	targetFramework string
	autoScaffold    bool
	dryRun          bool
	force           bool
	mergeMode       bool
	noMerge         bool
	mergeAll        bool
	mergeStrategy   string

	// Schema override flags (can override values from schema file)
	schemaSolutionName        string
	schemaNamespaceRoot       string
	schemaABPVersion          string
	schemaPrimaryKeyType      string
	schemaDBProvider          string
	schemaGenerateControllers bool
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "abp-gen",
	Short: "ABP Framework code generator",
	Long: `abp-gen is a cross-platform CLI tool that generates ABP Framework C# code
from JSON schemas or through interactive prompts.

It supports:
  - Entity, DTO, Service, Repository, and Controller generation
  - One-to-Many and Many-to-Many relationships
  - Entity Framework Core and MongoDB
  - Configurable primary key types (Guid, long)
  - Interactive schema building mode
  - Smart file merging with conflict resolution`,
	Version: fmt.Sprintf("%s (commit: %s, built: %s)", Version, GitCommit, BuildDate),
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Long:  "Print the version number and build information for abp-gen",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("abp-gen version %s\n", Version)
		fmt.Printf("Commit: %s\n", GitCommit)
		fmt.Printf("Built: %s\n", BuildDate)
	},
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Extract embedded templates to local directory",
	Long: `Extracts all embedded templates to ./abp-gen-templates/ directory.
This allows you to customize the templates for your specific needs.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runInit()
	},
}

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate ABP code from schema",
	Long: `Generates ABP Framework C# code from a JSON schema file or through interactive prompts.

If --input is not provided, enters interactive mode to build the schema.

Schema values can be overridden via CLI flags. CLI flags take precedence over schema file values.

Examples:
  # Generate from schema file
  abp-gen generate --input schema.json

  # Override module name from command line
  abp-gen generate --input schema.json --moduleName=ProductService

  # Override multiple values
  abp-gen generate --input schema.json --moduleName=ProductService --namespaceRoot=MyCompany.MyApp --primaryKeyType=Guid

  # Interactive mode
  abp-gen generate

  # Dry run to preview changes
  abp-gen generate --input schema.json --dry-run

  # Force overwrite existing files
  abp-gen generate --input schema.json --force`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runGenerate()
	},
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	// Generate command flags
	generateCmd.Flags().StringVarP(&inputFile, "input", "i", "", "input schema JSON file (optional, triggers interactive mode if not provided)")
	generateCmd.Flags().StringVarP(&solutionPath, "solution", "s", "", "path to solution file (auto-detected if not provided)")
	generateCmd.Flags().StringVarP(&moduleName, "module", "m", "", "module name (read from schema if not provided)")
	generateCmd.Flags().StringVar(&moduleName, "moduleName", "", "module name (same as --module, -m)")
	generateCmd.Flags().StringVarP(&templatesPath, "templates", "t", "", "custom templates directory")
	generateCmd.Flags().StringVar(&targetFramework, "target", "auto", "target framework: aspnetcore9, aspnetcore10, abp8-monolith, abp8-microservice, abp9-*, abp10-*, or auto")
	generateCmd.Flags().BoolVar(&autoScaffold, "auto-scaffold", false, "automatically create missing solutions/projects without prompting")
	generateCmd.Flags().BoolVar(&dryRun, "dry-run", false, "preview changes without writing files")
	generateCmd.Flags().BoolVar(&force, "force", false, "overwrite existing files")
	generateCmd.Flags().BoolVar(&mergeMode, "merge", false, "enable smart merge mode for existing files")
	generateCmd.Flags().BoolVar(&noMerge, "no-merge", false, "disable merge mode (skip existing files)")
	generateCmd.Flags().BoolVar(&mergeAll, "merge-all", false, "automatically merge all files without prompting")
	generateCmd.Flags().StringVar(&mergeStrategy, "merge-strategy", "", "merge strategy: pattern, ast, or json (auto-detected if not specified)")

	// Schema override flags - can override values from schema file
	generateCmd.Flags().StringVar(&schemaSolutionName, "solutionName", "", "solution name (overrides schema)")
	generateCmd.Flags().StringVar(&schemaNamespaceRoot, "namespaceRoot", "", "namespace root (overrides schema, e.g., MyCompany.MyApp)")
	generateCmd.Flags().StringVar(&schemaABPVersion, "abpVersion", "", "ABP version (overrides schema, e.g., 10.0)")
	generateCmd.Flags().StringVar(&schemaPrimaryKeyType, "primaryKeyType", "", "primary key type: Guid or long (overrides schema)")
	generateCmd.Flags().StringVar(&schemaDBProvider, "dbProvider", "", "database provider: efcore, mongodb, or both (overrides schema)")
	generateCmd.Flags().BoolVar(&schemaGenerateControllers, "generateControllers", false, "generate controllers (overrides schema)")

	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(versionCmd)
}

func runInit() error {
	fmt.Println("Extracting embedded templates to ./abp-gen-templates/...")

	if err := templates.ExtractTemplates("./abp-gen-templates"); err != nil {
		return fmt.Errorf("failed to extract templates: %w", err)
	}

	fmt.Println("\n✓ Templates extracted successfully!")
	fmt.Println("\nYou can now customize the templates in ./abp-gen-templates/")
	fmt.Println("Use --templates ./abp-gen-templates when generating code to use customized templates.")

	return nil
}

// detectAndPromptMissingFields detects missing required fields from solution structure
// and prompts user if detection fails
func detectAndPromptMissingFields(sch *schema.Schema, solutionInfo *detector.SolutionInfo, solutionDetectErr error) error {
	// Detect solution name
	if sch.Solution.Name == "" {
		if solutionDetectErr == nil && solutionInfo != nil && solutionInfo.Name != "" {
			sch.Solution.Name = solutionInfo.Name
			if verbose {
				fmt.Printf("✓ Auto-detected solution name from solution file: %s\n", solutionInfo.Name)
			}
		} else {
			// Try to get from current directory name
			wd, wdErr := os.Getwd()
			if wdErr == nil {
				dirName := filepath.Base(wd)
				if dirName != "" && dirName != "." && dirName != "/" {
					sch.Solution.Name = dirName
					if verbose {
						fmt.Printf("✓ Auto-detected solution name from current directory: %s\n", dirName)
					}
				}
			}

			// If still empty, prompt user
			if sch.Solution.Name == "" {
				fmt.Print("Solution name not found. Please enter solution name: ")
				var solutionName string
				fmt.Scanln(&solutionName)
				if solutionName == "" {
					return fmt.Errorf("solution name is required")
				}
				sch.Solution.Name = solutionName
			}
		}
	}

	// Detect module name from project names
	if sch.Solution.ModuleName == "" {
		detectedModuleName := ""
		if solutionInfo != nil && len(solutionInfo.Projects) > 0 {
			// Try to extract module name from project names
			// Pattern: SolutionName.ModuleName.Domain, SolutionName.ModuleName.Application, etc.
			for _, project := range solutionInfo.Projects {
				projectName := project.Name
				// Remove solution name prefix if present
				if strings.HasPrefix(projectName, sch.Solution.Name+".") {
					remaining := strings.TrimPrefix(projectName, sch.Solution.Name+".")
					// Extract module name (part before first dot)
					parts := strings.Split(remaining, ".")
					if len(parts) > 0 {
						candidate := parts[0]
						// Remove "Module" suffix if present (e.g., "UserModule" -> "User")
						candidate = strings.TrimSuffix(candidate, "Module")
						// Check if it's a valid module name (not a layer name)
						layerNames := map[string]bool{
							"Domain": true, "Application": true, "HttpApi": true,
							"EntityFrameworkCore": true, "MongoDB": true,
						}
						if !layerNames[candidate] && candidate != "" {
							detectedModuleName = candidate
							break
						}
					}
				}
			}
		}

		if detectedModuleName != "" {
			sch.Solution.ModuleName = detectedModuleName
			if verbose {
				fmt.Printf("✓ Auto-detected module name from project structure: %s\n", detectedModuleName)
			}
		} else {
			// Prompt user for module name
			fmt.Print("Module name not found. Please enter module name: ")
			var moduleNameInput string
			fmt.Scanln(&moduleNameInput)
			if moduleNameInput == "" {
				return fmt.Errorf("module name is required")
			}
			sch.Solution.ModuleName = moduleNameInput
		}
	}

	// Detect namespace root (defaults to solution name, but can be detected from projects)
	if sch.Solution.NamespaceRoot == "" {
		// Will be set to Solution.Name in validator, but we can try to detect from projects
		if solutionInfo != nil && len(solutionInfo.Projects) > 0 {
			// Try to extract namespace from first project
			for _, project := range solutionInfo.Projects {
				projectName := project.Name
				// If project name contains dots, extract namespace root
				if strings.Contains(projectName, ".") {
					parts := strings.Split(projectName, ".")
					if len(parts) >= 2 {
						// Take first part as namespace root
						candidate := parts[0]
						if candidate != "" && candidate != sch.Solution.Name {
							sch.Solution.NamespaceRoot = candidate
							if verbose {
								fmt.Printf("✓ Auto-detected namespace root from project: %s\n", candidate)
							}
							break
						}
					}
				}
			}
		}
		// If still empty, will default to Solution.Name in validator
	}

	// Detect ABP version from solution projects
	if sch.Solution.ABPVersion == "" {
		if solutionInfo != nil && len(solutionInfo.Projects) > 0 {
			abpVer, _ := detector.ScanProjectsForVersions(solutionInfo)
			if abpVer != "" {
				sch.Solution.ABPVersion = abpVer + ".0" // Convert "8" to "8.0"
				if verbose {
					fmt.Printf("✓ Auto-detected ABP version: %s\n", sch.Solution.ABPVersion)
				}
			}
		}
		// If still empty, will default to "9.0" in validator
	}

	// Detect primary key type (hard to detect, will default to "Guid" in validator)
	// Can check project files for entity base classes, but that's complex
	// For now, we'll let it default

	// Detect DB provider (can check for MongoDB projects)
	if sch.Solution.DBProvider == "" {
		if solutionInfo != nil {
			hasMongoDB := false
			hasEFCore := false
			for _, project := range solutionInfo.Projects {
				if project.Type == detector.ProjectTypeMongoDB {
					hasMongoDB = true
				}
				if project.Type == detector.ProjectTypeEntityFrameworkCore {
					hasEFCore = true
				}
			}
			if hasMongoDB && hasEFCore {
				sch.Solution.DBProvider = "both"
				if verbose {
					fmt.Printf("✓ Auto-detected database provider: both (EF Core and MongoDB)\n")
				}
			} else if hasMongoDB {
				sch.Solution.DBProvider = "mongodb"
				if verbose {
					fmt.Printf("✓ Auto-detected database provider: mongodb\n")
				}
			} else if hasEFCore {
				sch.Solution.DBProvider = "efcore"
				if verbose {
					fmt.Printf("✓ Auto-detected database provider: efcore\n")
				}
			}
		}
		// If still empty, will default to "efcore" in validator
	}

	return nil
}

// applySchemaOverrides applies CLI flag values to schema, overriding schema file values.
// CLI flags take precedence over schema file values when provided.
func applySchemaOverrides(sch *schema.Schema) {
	// Override solution name
	if schemaSolutionName != "" {
		sch.Solution.Name = schemaSolutionName
		if verbose {
			fmt.Printf("✓ Overriding solution name from CLI: %s\n", schemaSolutionName)
		}
	}

	// Override namespace root
	if schemaNamespaceRoot != "" {
		sch.Solution.NamespaceRoot = schemaNamespaceRoot
		if verbose {
			fmt.Printf("✓ Overriding namespace root from CLI: %s\n", schemaNamespaceRoot)
		}
	}

	// Override module name (supports both --module/-m and --moduleName)
	if moduleName != "" {
		sch.Solution.ModuleName = moduleName
		if verbose {
			fmt.Printf("✓ Overriding module name from CLI: %s\n", moduleName)
		}
	}

	// Override ABP version
	if schemaABPVersion != "" {
		sch.Solution.ABPVersion = schemaABPVersion
		if verbose {
			fmt.Printf("✓ Overriding ABP version from CLI: %s\n", schemaABPVersion)
		}
	}

	// Override primary key type
	if schemaPrimaryKeyType != "" {
		sch.Solution.PrimaryKeyType = schemaPrimaryKeyType
		if verbose {
			fmt.Printf("✓ Overriding primary key type from CLI: %s\n", schemaPrimaryKeyType)
		}
	}

	// Override database provider
	if schemaDBProvider != "" {
		sch.Solution.DBProvider = schemaDBProvider
		if verbose {
			fmt.Printf("✓ Overriding database provider from CLI: %s\n", schemaDBProvider)
		}
	}

	// Override generate controllers
	// Note: Bool flags default to false, so we check if it was explicitly set
	// by checking the command's flag set. For now, we'll override if true.
	// Users can set it to false in schema file if they don't want controllers.
	if schemaGenerateControllers {
		sch.Solution.GenerateControllers = true
		if verbose {
			fmt.Println("✓ Overriding generate controllers from CLI: true")
		}
	}
}

func runGenerate() error {
	// Load or build schema
	var sch *schema.Schema
	var err error

	if inputFile != "" {
		// Load from file
		fmt.Printf("Loading schema from %s...\n", inputFile)
		sch, err = schema.LoadFromFile(inputFile)
		if err != nil {
			return fmt.Errorf("failed to load schema: %w", err)
		}
	} else {
		// Interactive mode
		sch, err = prompts.BuildSchemaInteractively()
		if err != nil {
			return fmt.Errorf("failed to build schema interactively: %w", err)
		}
	}

	// Apply CLI flag overrides to schema (CLI flags take precedence)
	applySchemaOverrides(sch)

	// Try to detect solution and all missing information before validation
	var solutionInfo *detector.SolutionInfo
	var solutionDetectErr error

	// Try to detect solution first
	fmt.Println("\nDetecting solution structure...")
	if solutionPath != "" {
		solutionInfo, solutionDetectErr = detector.ParseSolution(solutionPath)
	} else {
		solutionInfo, solutionDetectErr = detector.FindSolution(".")
	}

	// Detect and prompt for all missing required fields
	if err := detectAndPromptMissingFields(sch, solutionInfo, solutionDetectErr); err != nil {
		return err
	}

	// Validate schema
	if err := sch.Validate(); err != nil {
		return fmt.Errorf("schema validation failed: %w", err)
	}

	// Detect solution if not already detected (for use in rest of function)
	if solutionInfo == nil {
		if solutionPath != "" {
			solutionInfo, err = detector.ParseSolution(solutionPath)
		} else {
			solutionInfo, err = detector.FindSolution(".")
		}
	} else {
		// Reuse the error from earlier detection attempt
		err = solutionDetectErr
	}

	// If no solution found, offer to create one
	if err != nil {
		scaffolder := prompts.NewScaffolder()
		created, newSolutionPath, scaffoldErr := scaffolder.PromptCreateSolution(".", autoScaffold)

		if !created {
			if scaffoldErr != nil {
				return fmt.Errorf("failed to detect or create solution: %w", scaffoldErr)
			}
			return fmt.Errorf("failed to detect solution: %w", err)
		}

		// Try to detect the newly created solution
		solutionInfo, err = detector.FindSolution(newSolutionPath)
		if err != nil {
			return fmt.Errorf("failed to detect newly created solution: %w", err)
		}
	}

	fmt.Printf("✓ Found solution: %s\n", solutionInfo.Name)

	// Determine target framework
	effectiveTarget := targetFramework
	if effectiveTarget == "auto" || effectiveTarget == "" {
		effectiveTarget = solutionInfo.TargetFramework
		fmt.Printf("✓ Auto-detected target framework: %s", effectiveTarget)

		// Show detected versions for transparency
		if verbose {
			abpVer, dotnetVer := detector.ScanProjectsForVersions(solutionInfo)
			if abpVer != "" {
				fmt.Printf(" (ABP %s", abpVer)
				if dotnetVer != "" {
					fmt.Printf(", .NET %s", dotnetVer)
				}
				fmt.Printf(")")
			} else if dotnetVer != "" {
				fmt.Printf(" (.NET %s)", dotnetVer)
			}
		}
		fmt.Println()
	} else {
		fmt.Printf("✓ Using specified target framework: %s\n", effectiveTarget)
	}

	// Update schema with target framework if not already set
	if sch.Solution.TargetFramework == "" || sch.Solution.TargetFramework == "auto" {
		sch.Solution.TargetFramework = schema.TargetFramework(effectiveTarget)
	}

	// Auto-detect and show configuration summary
	configScanner := detector.NewConfigScanner()
	tenancyEnabled, tenancyStrategy, _ := configScanner.DetectMultiTenancy(solutionInfo)

	// Update schema with detected settings if not already set
	if sch.Solution.MultiTenancy == nil && tenancyEnabled {
		sch.Solution.MultiTenancy = &schema.MultiTenancy{
			Enabled:             true,
			Strategy:            tenancyStrategy,
			EnableDataIsolation: true,
			TenantIdProperty:    "TenantId",
		}
		if verbose {
			fmt.Printf("✓ Auto-detected multi-tenancy: %s strategy\n", tenancyStrategy)
		}
	}

	// Show detected projects in verbose mode
	if verbose {
		fmt.Printf("\nDetected projects:\n")
		for _, project := range solutionInfo.Projects {
			projectType := string(project.Type)
			if projectType == "Unknown" {
				projectType = "Unknown (not recognized as ABP layer)"
			}
			fmt.Printf("  - %s (%s)\n", project.Name, projectType)
		}

		// Show configuration summary
		fmt.Println("\nConfiguration Summary:")
		fmt.Print(configScanner.SummarizeConfiguration(solutionInfo))
	}

	// Detect layer paths
	module := sch.Solution.ModuleName
	if moduleName != "" {
		module = moduleName
	}

	paths, err := detector.DetectLayerPaths(solutionInfo, module)
	if err != nil {
		return fmt.Errorf("failed to detect layer paths: %w", err)
	}

	// Ensure directories exist
	if !dryRun {
		if err := paths.EnsureDirectories(); err != nil {
			return fmt.Errorf("failed to create directories: %w", err)
		}
		// Ensure module-specific directories exist
		if err := paths.EnsureModuleDirectories(module); err != nil {
			return fmt.Errorf("failed to create module directories: %w", err)
		}
	}

	// Handle merge flags
	enableMerge := mergeMode && !noMerge && !force

	// Initialize generators with target framework
	tmplLoader := templates.NewLoaderWithTarget(templatesPath, effectiveTarget)
	w := writer.NewWriterWithMerge(dryRun, force, verbose, enableMerge)

	// Configure merge engine with flags if merge is enabled
	if enableMerge && mergeAll {
		// mergeStrategy is reserved for future use to specify merge strategy
		// Currently, the strategy is auto-detected based on file type
		_ = mergeStrategy
		w.SetMergeAll(true)
	}

	entityGen := generator.NewEntityGenerator(tmplLoader, w)
	managerGen := generator.NewManagerGenerator(tmplLoader, w)
	dtoGen := generator.NewDTOGenerator(tmplLoader, w)
	validatorGen := generator.NewValidatorGenerator(tmplLoader, w)
	eventHandlerGen := generator.NewEventHandlerGenerator(tmplLoader, w)
	serviceGen := generator.NewServiceGenerator(tmplLoader, w)
	permissionsGen := generator.NewPermissionsGenerator(tmplLoader, w)
	relationHandler := generator.NewRelationshipHandler()
	customRepoGen := generator.NewCustomRepositoryGenerator(tmplLoader, w)
	domainEventsGen := generator.NewDomainEventsGenerator(tmplLoader, w)
	enumGen := generator.NewEnumGenerator(tmplLoader, w)
	valueObjectGen := generator.NewValueObjectGenerator(tmplLoader, w)
	localizationGen := generator.NewLocalizationGenerator(w)
	integrationTestGen := generator.NewIntegrationTestGenerator(tmplLoader, w)

	var efcoreGen *generator.EFCoreGenerator
	var mongoGen *generator.MongoDBGenerator

	if sch.Solution.DBProvider == "efcore" || sch.Solution.DBProvider == "both" {
		efcoreGen = generator.NewEFCoreGenerator(tmplLoader, w)
	}

	if sch.Solution.DBProvider == "mongodb" || sch.Solution.DBProvider == "both" {
		mongoGen = generator.NewMongoDBGenerator(tmplLoader, w)
	}

	// Print merge mode status
	if enableMerge {
		fmt.Println("\n✓ Smart merge mode enabled - existing files will be merged intelligently")
	} else if force {
		fmt.Println("\n⚠️  Force mode enabled - existing files will be overwritten")
	} else {
		fmt.Println("\n✓ Safe mode - existing files will be skipped")
	}

	// Generate test project if integration tests are enabled
	if sch.Options.GenerateIntegrationTests {
		fmt.Println("\n✓ Integration tests enabled - generating test infrastructure")
		if err := integrationTestGen.GenerateTestProject(sch, paths); err != nil {
			fmt.Printf("⚠️  Failed to generate test project: %v\n", err)
		}
	}

	// Generate code for each entity
	fmt.Printf("\nGenerating code for %d entity(s)...\n\n", len(sch.Entities))

	for i, entity := range sch.Entities {
		fmt.Printf("[%d/%d] Generating %s...\n", i+1, len(sch.Entities), entity.Name)

		// Process relationships
		if err := relationHandler.ProcessRelationships(sch, &entity); err != nil {
			return fmt.Errorf("failed to process relationships for %s: %w", entity.Name, err)
		}

		// Generate enums if defined
		if err := enumGen.Generate(sch, &entity, paths); err != nil {
			return fmt.Errorf("failed to generate enums for %s: %w", entity.Name, err)
		}

		// Generate value object or entity
		if entity.EntityType == "ValueObject" {
			if err := valueObjectGen.Generate(sch, &entity, paths); err != nil {
				return fmt.Errorf("failed to generate value object %s: %w", entity.Name, err)
			}
			if err := valueObjectGen.GenerateFactory(sch, &entity, paths); err != nil {
				return fmt.Errorf("failed to generate value object factory for %s: %w", entity.Name, err)
			}
		} else {
			// Generate entity and related files
			if err := entityGen.Generate(sch, &entity, paths); err != nil {
				return fmt.Errorf("failed to generate entity %s: %w", entity.Name, err)
			}
		}

		if err := entityGen.GenerateRepository(sch, &entity, paths); err != nil {
			return fmt.Errorf("failed to generate repository for %s: %w", entity.Name, err)
		}

		// Generate custom repository if defined
		if err := customRepoGen.Generate(sch, &entity, paths); err != nil {
			return fmt.Errorf("failed to generate custom repository for %s: %w", entity.Name, err)
		}

		// Generate domain events if defined
		if err := domainEventsGen.Generate(sch, &entity, paths); err != nil {
			return fmt.Errorf("failed to generate domain events for %s: %w", entity.Name, err)
		}

		if err := managerGen.Generate(sch, &entity, paths); err != nil {
			return fmt.Errorf("failed to generate manager for %s: %w", entity.Name, err)
		}

		if err := entityGen.GenerateConstants(sch, &entity, paths); err != nil {
			return fmt.Errorf("failed to generate constants for %s: %w", entity.Name, err)
		}

		if err := entityGen.GenerateEvents(sch, &entity, paths); err != nil {
			return fmt.Errorf("failed to generate events for %s: %w", entity.Name, err)
		}

		if err := entityGen.GenerateDataSeeder(sch, &entity, paths); err != nil {
			return fmt.Errorf("failed to generate data seeder for %s: %w", entity.Name, err)
		}

		// Generate DTOs
		if err := dtoGen.Generate(sch, &entity, paths); err != nil {
			return fmt.Errorf("failed to generate DTOs for %s: %w", entity.Name, err)
		}

		if err := dtoGen.GenerateAppServiceInterface(sch, &entity, paths); err != nil {
			return fmt.Errorf("failed to generate app service interface for %s: %w", entity.Name, err)
		}

		// Generate validators
		if err := validatorGen.Generate(sch, &entity, paths); err != nil {
			return fmt.Errorf("failed to generate validators for %s: %w", entity.Name, err)
		}

		// Generate service
		if err := serviceGen.Generate(sch, &entity, paths); err != nil {
			return fmt.Errorf("failed to generate service for %s: %w", entity.Name, err)
		}

		if err := serviceGen.GenerateAutoMapperProfile(sch, &entity, paths); err != nil {
			return fmt.Errorf("failed to generate AutoMapper profile for %s: %w", entity.Name, err)
		}

		if err := serviceGen.GenerateController(sch, &entity, paths); err != nil {
			return fmt.Errorf("failed to generate controller for %s: %w", entity.Name, err)
		}

		// Generate permissions
		if err := permissionsGen.Generate(sch, &entity, paths); err != nil {
			return fmt.Errorf("failed to generate permissions for %s: %w", entity.Name, err)
		}

		if err := permissionsGen.GenerateLocalization(sch, &entity, paths); err != nil {
			return fmt.Errorf("failed to generate localization for %s: %w", entity.Name, err)
		}

		// Generate and merge localization files
		if err := localizationGen.GenerateEntityLocalization(sch, &entity, paths); err != nil {
			return fmt.Errorf("failed to generate entity localization for %s: %w", entity.Name, err)
		}

		// Generate event handlers
		if err := eventHandlerGen.Generate(sch, &entity, paths); err != nil {
			return fmt.Errorf("failed to generate event handlers for %s: %w", entity.Name, err)
		}

		// Generate EF Core files
		if efcoreGen != nil {
			if err := efcoreGen.Generate(sch, &entity, paths); err != nil {
				return fmt.Errorf("failed to generate EF Core files for %s: %w", entity.Name, err)
			}
		}

		// Generate MongoDB files
		if mongoGen != nil {
			if err := mongoGen.Generate(sch, &entity, paths); err != nil {
				return fmt.Errorf("failed to generate MongoDB files for %s: %w", entity.Name, err)
			}
		}

		// Generate integration tests
		if err := integrationTestGen.Generate(sch, &entity, paths); err != nil {
			return fmt.Errorf("failed to generate integration tests for %s: %w", entity.Name, err)
		}

		fmt.Printf("✓ Generated %s\n\n", entity.Name)
	}

	// Print summary
	w.PrintSummary()

	if dryRun {
		fmt.Println("\nTo apply these changes, run the command without --dry-run")
	} else {
		fmt.Println("\n✓ Code generation completed successfully!")
		fmt.Println("\nNext steps:")
		fmt.Println("  1. Add database migration: dotnet ef migrations add Add<EntityName>")
		fmt.Println("  2. Update database: dotnet ef database update")
		fmt.Println("  3. Build solution: dotnet build")
	}

	return nil
}
