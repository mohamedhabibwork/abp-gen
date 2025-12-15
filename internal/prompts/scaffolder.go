package prompts

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// scaffolder.go handles interactive creation of new ABP and ASP.NET solutions
// using `abp` or `dotnet` CLI tools when no existing solution is found.

// Scaffolder handles creation of new solutions and projects via CLI tools
type Scaffolder struct {
	reader *bufio.Reader
}

// NewScaffolder creates a new scaffolder
func NewScaffolder() *Scaffolder {
	return &Scaffolder{
		reader: bufio.NewReader(os.Stdin),
	}
}

// PromptCreateSolution prompts user to create a new ABP or ASP.NET solution
// Returns: (created bool, solutionPath string, error)
func (s *Scaffolder) PromptCreateSolution(workingDir string, autoScaffold bool) (bool, string, error) {
	if !autoScaffold {
		fmt.Println("\n❌ No solution found in the current directory or parent directories.")
		fmt.Print("Would you like to create a new solution? (y/N): ")

		response, _ := s.reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))

		if response != "y" && response != "yes" {
			return false, "", fmt.Errorf("solution creation declined by user")
		}
	}

	// Determine if abp CLI is available
	hasAbpCLI := s.checkABPCLI()
	hasDotnetCLI := s.checkDotNetCLI()

	if !hasAbpCLI && !hasDotnetCLI {
		return false, "", fmt.Errorf("neither 'abp' nor 'dotnet' CLI tools are available. Please install one to create a new solution")
	}

	// Get solution details from user
	solutionName, template, err := s.promptSolutionDetails(hasAbpCLI, autoScaffold)
	if err != nil {
		return false, "", err
	}

	// Execute the appropriate CLI command
	var solutionPath string
	if hasAbpCLI && (template == "app" || template == "microservice" || template == "module") {
		solutionPath, err = s.createABPSolution(workingDir, solutionName, template)
	} else if hasDotnetCLI {
		solutionPath, err = s.createDotNetSolution(workingDir, solutionName, template)
	} else {
		return false, "", fmt.Errorf("no suitable CLI tool available for selected template")
	}

	if err != nil {
		return false, "", err
	}

	fmt.Printf("\n✓ Solution created successfully at: %s\n", solutionPath)
	return true, solutionPath, nil
}

// checkABPCLI checks if ABP CLI is installed
func (s *Scaffolder) checkABPCLI() bool {
	cmd := exec.Command("abp", "--version")
	err := cmd.Run()
	return err == nil
}

// checkDotNetCLI checks if .NET CLI is installed
func (s *Scaffolder) checkDotNetCLI() bool {
	cmd := exec.Command("dotnet", "--version")
	err := cmd.Run()
	return err == nil
}

// promptSolutionDetails prompts for solution name and template type
func (s *Scaffolder) promptSolutionDetails(hasAbpCLI bool, autoScaffold bool) (name, template string, err error) {
	// Get solution name
	fmt.Print("\nEnter solution name (e.g., MyCompany.MyProject): ")
	name, _ = s.reader.ReadString('\n')
	name = strings.TrimSpace(name)

	if name == "" {
		return "", "", fmt.Errorf("solution name is required")
	}

	if autoScaffold {
		// Default to app template
		template = "app"
		return name, template, nil
	}

	// Get template type
	fmt.Println("\nSelect template type:")
	if hasAbpCLI {
		fmt.Println("  1. ABP Application (app) - Monolithic web application")
		fmt.Println("  2. ABP Microservice (microservice) - Microservice solution")
		fmt.Println("  3. ABP Module (module) - Reusable module")
		fmt.Println("  4. ASP.NET Core Web API (webapi) - Simple Web API")
	} else {
		fmt.Println("  1. ASP.NET Core Web API (webapi)")
		fmt.Println("  2. ASP.NET Core MVC (mvc)")
	}

	fmt.Print("Enter choice (1-4 or template name): ")
	choice, _ := s.reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	// Map choice to template
	if hasAbpCLI {
		switch choice {
		case "1", "app":
			template = "app"
		case "2", "microservice":
			template = "microservice"
		case "3", "module":
			template = "module"
		case "4", "webapi":
			template = "webapi"
		default:
			template = choice // Allow custom template names
		}
	} else {
		switch choice {
		case "1", "webapi":
			template = "webapi"
		case "2", "mvc":
			template = "mvc"
		default:
			template = "webapi" // Default
		}
	}

	return name, template, nil
}

// createABPSolution creates a new ABP solution using the ABP CLI
func (s *Scaffolder) createABPSolution(workingDir, solutionName, template string) (string, error) {
	fmt.Printf("\nCreating ABP solution with command: abp new %s -t %s\n", solutionName, template)
	fmt.Println("This may take a few minutes...")

	cmd := exec.Command("abp", "new", solutionName, "-t", template)
	cmd.Dir = workingDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to create ABP solution: %w", err)
	}

	// ABP CLI creates a folder with the solution name
	solutionPath := filepath.Join(workingDir, solutionName)
	return solutionPath, nil
}

// createDotNetSolution creates a new .NET solution using the dotnet CLI
func (s *Scaffolder) createDotNetSolution(workingDir, solutionName, template string) (string, error) {
	fmt.Printf("\nCreating .NET solution with command: dotnet new %s -n %s\n", template, solutionName)

	cmd := exec.Command("dotnet", "new", template, "-n", solutionName)
	cmd.Dir = workingDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to create .NET solution: %w", err)
	}

	// .NET CLI creates a folder with the solution name
	solutionPath := filepath.Join(workingDir, solutionName)
	return solutionPath, nil
}

// PromptForMissingInfo prompts user for information that couldn't be auto-detected
func (s *Scaffolder) PromptForMissingInfo(question string, defaultValue string) string {
	if defaultValue != "" {
		fmt.Printf("%s [%s]: ", question, defaultValue)
	} else {
		fmt.Printf("%s: ", question)
	}

	response, _ := s.reader.ReadString('\n')
	response = strings.TrimSpace(response)

	if response == "" {
		return defaultValue
	}
	return response
}

// ConfirmAutoDetected asks user to confirm auto-detected settings
func (s *Scaffolder) ConfirmAutoDetected(setting, value string) bool {
	fmt.Printf("Auto-detected %s: %s. Use this? (Y/n): ", setting, value)

	response, _ := s.reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))

	// Default to yes
	return response == "" || response == "y" || response == "yes"
}
