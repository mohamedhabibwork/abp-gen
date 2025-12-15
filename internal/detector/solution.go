package detector

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SolutionInfo contains information about the detected ABP solution
type SolutionInfo struct {
	Path            string
	Name            string
	RootDirectory   string
	Projects        []ProjectInfo
	TargetFramework string // Detected target framework: "aspnetcore9", "abp8-microservice", "abp8-monolith"
	IsMicroservice  bool   // Whether this is a microservice architecture
}

// ProjectInfo contains information about a project in the solution
type ProjectInfo struct {
	Name      string
	Path      string
	Directory string
	Type      ProjectType
}

// ProjectType represents the ABP layer type
type ProjectType string

const (
	ProjectTypeDomain               ProjectType = "Domain"
	ProjectTypeDomainShared         ProjectType = "Domain.Shared"
	ProjectTypeApplicationContracts ProjectType = "Application.Contracts"
	ProjectTypeApplication          ProjectType = "Application"
	ProjectTypeHttpApi              ProjectType = "HttpApi"
	ProjectTypeEntityFrameworkCore  ProjectType = "EntityFrameworkCore"
	ProjectTypeMongoDB              ProjectType = "MongoDB"
	ProjectTypeUnknown              ProjectType = "Unknown"
)

// FindSolution searches for solution files (.sln, .slnx, .abpsln, .abpslnx)
// starting from the current directory and moving upward through parent directories.
// If no solution file is found, attempts to discover projects from .csproj files.
func FindSolution(startPath string) (*SolutionInfo, error) {
	currentPath, err := filepath.Abs(startPath)
	if err != nil {
		return nil, err
	}

	// Solution file extensions in priority order
	solutionExtensions := []string{".sln", ".slnx", ".abpsln", ".abpslnx"}

	for {
		// Check for solution files in current directory
		for _, ext := range solutionExtensions {
			files, err := filepath.Glob(filepath.Join(currentPath, "*"+ext))
			if err != nil {
				continue
			}

			if len(files) > 0 {
				// Found a solution file
				return ParseSolution(files[0])
			}
		}

		// Move up one directory
		parentPath := filepath.Dir(currentPath)
		if parentPath == currentPath {
			// Reached root directory without finding solution
			// Try to discover from csproj files
			return discoverFromProjects(startPath)
		}
		currentPath = parentPath
	}
}

// discoverFromProjects attempts to discover solution info from .csproj files
// when no traditional solution file exists
func discoverFromProjects(startPath string) (*SolutionInfo, error) {
	csprojFiles, err := filepath.Glob(filepath.Join(startPath, "**/*.csproj"))
	if err != nil || len(csprojFiles) == 0 {
		// Try current directory
		csprojFiles, err = filepath.Glob(filepath.Join(startPath, "*.csproj"))
		if err != nil || len(csprojFiles) == 0 {
			return nil, fmt.Errorf("no solution or project files found")
		}
	}

	// Create a synthetic solution from discovered projects
	info := &SolutionInfo{
		Path:            startPath,
		Name:            filepath.Base(startPath),
		RootDirectory:   startPath,
		Projects:        []ProjectInfo{},
		TargetFramework: "abp8-monolith", // Default, will be refined
		IsMicroservice:  false,
	}

	// Parse each csproj to extract project info
	for _, csprojPath := range csprojFiles {
		project := parseCsprojFile(csprojPath)
		if project != nil {
			info.Projects = append(info.Projects, *project)
		}
	}

	// Detect target framework and architecture from projects
	info.TargetFramework = DetectTargetFramework(info)
	info.IsMicroservice = IsMicroserviceArchitecture(info)

	return info, nil
}

// ParseSolution parses a solution file and extracts project information
func ParseSolution(solutionPath string) (*SolutionInfo, error) {
	file, err := os.Open(solutionPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	solutionDir := filepath.Dir(solutionPath)
	solutionName := strings.TrimSuffix(filepath.Base(solutionPath), ".sln")

	info := &SolutionInfo{
		Path:            solutionPath,
		Name:            solutionName,
		RootDirectory:   solutionDir,
		Projects:        []ProjectInfo{},
		TargetFramework: "abp8-monolith", // Default
		IsMicroservice:  false,
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Look for project lines: Project("{...}") = "ProjectName", "Path\To\Project.csproj", "{...}"
		if strings.HasPrefix(line, "Project(") {
			project := parseProjectLine(line, solutionDir)
			if project != nil {
				info.Projects = append(info.Projects, *project)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Detect target framework based on projects and structure
	info.TargetFramework = DetectTargetFramework(info)
	info.IsMicroservice = IsMicroserviceArchitecture(info)

	return info, nil
}

// DetectTargetFramework detects the target framework based on solution structure and csproj files
func DetectTargetFramework(info *SolutionInfo) string {
	// Scan projects for ABP and .NET versions
	abpVersion, dotnetVersion := ScanProjectsForVersions(info)

	// Check for microservice architecture
	isMicro := IsMicroserviceArchitecture(info)

	// Map versions to target framework
	targetFramework := MapToTargetFramework(abpVersion, dotnetVersion, isMicro)

	return targetFramework
}

// IsMicroserviceArchitecture checks if the solution follows microservice architecture patterns
func IsMicroserviceArchitecture(info *SolutionInfo) bool {
	// Check for common microservice indicators:
	// 1. Multiple service projects
	// 2. Shared projects across services
	// 3. Gateway or API gateway projects
	// 4. Message bus or event bus projects

	serviceCount := 0
	hasGateway := false
	hasShared := false

	for _, proj := range info.Projects {
		nameLower := strings.ToLower(proj.Name)

		// Count service projects
		if strings.Contains(nameLower, "service") && !strings.Contains(nameLower, "shared") {
			serviceCount++
		}

		// Check for gateway
		if strings.Contains(nameLower, "gateway") || strings.Contains(nameLower, "apigateway") {
			hasGateway = true
		}

		// Check for shared projects
		if strings.Contains(nameLower, "shared") {
			hasShared = true
		}
	}

	// Microservice if we have multiple services or gateway + services
	return serviceCount > 1 || (hasGateway && serviceCount > 0 && hasShared)
}

// parseProjectLine parses a project line from the solution file
// Format: Project("{GUID}") = "ProjectName", "path\Project.csproj", "{GUID}"
func parseProjectLine(line string, solutionDir string) *ProjectInfo {
	// Extract project name and path
	parts := strings.Split(line, "\"")
	// We need at least 8 parts: Project({, GUID, ) = , ProjectName, , path, , GUID, )
	if len(parts) < 8 {
		return nil
	}

	// Skip solution folders and other non-project entries
	// Solution folders have GUID {2150E333-8FDC-42A3-9474-1A3956D46DE8}
	// C# projects have GUID {9A19103F-16F7-4668-BE54-9A1E7A4F7556}
	projectTypeGUID := parts[1]

	// Skip solution folders
	if projectTypeGUID == "2150E333-8FDC-42A3-9474-1A3956D46DE8" {
		return nil
	}

	// Extract project name (parts[3]) and path (parts[5])
	projectName := parts[3]
	projectPath := parts[5]

	// Skip if path doesn't end with .csproj (not a C# project)
	if !strings.HasSuffix(projectPath, ".csproj") {
		return nil
	}

	// Convert relative path to absolute
	// Handle both Windows (\\) and Unix (/) path separators
	projectPath = strings.ReplaceAll(projectPath, "\\", string(filepath.Separator))
	absPath := filepath.Join(solutionDir, projectPath)
	projectDir := filepath.Dir(absPath)

	// Determine project type based on name
	projectType := DetermineProjectType(projectName)

	return &ProjectInfo{
		Name:      projectName,
		Path:      absPath,
		Directory: projectDir,
		Type:      projectType,
	}
}

// DetermineProjectType determines the ABP layer type from project name
// It checks both suffix patterns (e.g., "Module.Domain") and exact/contains patterns (e.g., "Domain")
func DetermineProjectType(projectName string) ProjectType {
	// Normalize project name for comparison
	nameLower := strings.ToLower(projectName)

	// Check for Domain.Shared first (more specific)
	if strings.HasSuffix(projectName, ".Domain.Shared") ||
		strings.Contains(nameLower, ".domain.shared") ||
		nameLower == "domain.shared" {
		return ProjectTypeDomainShared
	}

	// Check for Application.Contracts (more specific than Application)
	if strings.HasSuffix(projectName, ".Application.Contracts") ||
		strings.Contains(nameLower, ".application.contracts") ||
		nameLower == "application.contracts" {
		return ProjectTypeApplicationContracts
	}

	// Check for EntityFrameworkCore
	if strings.HasSuffix(projectName, ".EntityFrameworkCore") ||
		strings.Contains(nameLower, ".entityframeworkcore") ||
		nameLower == "entityframeworkcore" ||
		strings.Contains(nameLower, ".efcore") ||
		nameLower == "efcore" {
		return ProjectTypeEntityFrameworkCore
	}

	// Check for MongoDB
	if strings.HasSuffix(projectName, ".MongoDB") ||
		strings.Contains(nameLower, ".mongodb") ||
		nameLower == "mongodb" {
		return ProjectTypeMongoDB
	}

	// Check for Domain (after Domain.Shared to avoid false positives)
	if strings.HasSuffix(projectName, ".Domain") ||
		strings.Contains(nameLower, ".domain") ||
		nameLower == "domain" {
		return ProjectTypeDomain
	}

	// Check for Application (after Application.Contracts to avoid false positives)
	if strings.HasSuffix(projectName, ".Application") ||
		strings.Contains(nameLower, ".application") ||
		nameLower == "application" {
		return ProjectTypeApplication
	}

	// Check for HttpApi (be careful not to match Application)
	if strings.HasSuffix(projectName, ".HttpApi") ||
		strings.Contains(nameLower, ".httpapi") ||
		nameLower == "httpapi" ||
		(strings.HasSuffix(nameLower, ".api") && !strings.Contains(nameLower, "application")) {
		return ProjectTypeHttpApi
	}

	return ProjectTypeUnknown
}

// GetProject returns the project of a specific type
func (s *SolutionInfo) GetProject(projectType ProjectType) *ProjectInfo {
	for _, project := range s.Projects {
		if project.Type == projectType {
			return &project
		}
	}
	return nil
}

// HasProject checks if the solution has a project of the specified type
func (s *SolutionInfo) HasProject(projectType ProjectType) bool {
	return s.GetProject(projectType) != nil
}

// GetProjectDirectory returns the directory path for a specific project type
func (s *SolutionInfo) GetProjectDirectory(projectType ProjectType) string {
	project := s.GetProject(projectType)
	if project == nil {
		return ""
	}
	return project.Directory
}
