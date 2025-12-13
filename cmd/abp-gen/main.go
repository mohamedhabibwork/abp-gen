package main

import (
	"fmt"
	"os"

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
	inputFile      string
	solutionPath   string
	moduleName     string
	templatesPath  string
	dryRun         bool
	force          bool
	mergeMode      bool
	noMerge        bool
	mergeAll       bool
	mergeStrategy  string
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

Examples:
  # Generate from schema file
  abp-gen generate --input schema.json

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
	generateCmd.Flags().StringVarP(&templatesPath, "templates", "t", "", "custom templates directory")
	generateCmd.Flags().BoolVar(&dryRun, "dry-run", false, "preview changes without writing files")
	generateCmd.Flags().BoolVar(&force, "force", false, "overwrite existing files")

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

	// Validate schema
	if err := sch.Validate(); err != nil {
		return fmt.Errorf("schema validation failed: %w", err)
	}

	// Detect solution
	fmt.Println("\nDetecting ABP solution structure...")
	var solutionInfo *detector.SolutionInfo

	if solutionPath != "" {
		solutionInfo, err = detector.ParseSolution(solutionPath)
	} else {
		solutionInfo, err = detector.FindSolution(".")
	}

	if err != nil {
		return fmt.Errorf("failed to detect solution: %w", err)
	}

	fmt.Printf("✓ Found solution: %s\n", solutionInfo.Name)

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
	}

	// Handle merge flags
	enableMerge := mergeMode && !noMerge && !force
	
	// Initialize generators
	tmplLoader := templates.NewLoader(templatesPath)
	w := writer.NewWriterWithMerge(dryRun, force, verbose, enableMerge)

	entityGen := generator.NewEntityGenerator(tmplLoader, w)
	managerGen := generator.NewManagerGenerator(tmplLoader, w)
	dtoGen := generator.NewDTOGenerator(tmplLoader, w)
	validatorGen := generator.NewValidatorGenerator(tmplLoader, w)
	eventHandlerGen := generator.NewEventHandlerGenerator(tmplLoader, w)
	serviceGen := generator.NewServiceGenerator(tmplLoader, w)
	permissionsGen := generator.NewPermissionsGenerator(tmplLoader, w)
	relationHandler := generator.NewRelationshipHandler()

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
	
	// Generate code for each entity
	fmt.Printf("\nGenerating code for %d entity(s)...\n\n", len(sch.Entities))

	for i, entity := range sch.Entities {
		fmt.Printf("[%d/%d] Generating %s...\n", i+1, len(sch.Entities), entity.Name)

		// Process relationships
		if err := relationHandler.ProcessRelationships(sch, &entity); err != nil {
			return fmt.Errorf("failed to process relationships for %s: %w", entity.Name, err)
		}

		// Generate entity and related files
		if err := entityGen.Generate(sch, &entity, paths); err != nil {
			return fmt.Errorf("failed to generate entity %s: %w", entity.Name, err)
		}

		if err := entityGen.GenerateRepository(sch, &entity, paths); err != nil {
			return fmt.Errorf("failed to generate repository for %s: %w", entity.Name, err)
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

