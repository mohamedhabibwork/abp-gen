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
	Path          string
	Name          string
	RootDirectory string
	Projects      []ProjectInfo
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

// FindSolution searches for a .sln file starting from the current directory
// and moving upward through parent directories
func FindSolution(startPath string) (*SolutionInfo, error) {
	currentPath, err := filepath.Abs(startPath)
	if err != nil {
		return nil, err
	}

	for {
		// Check for .sln files in current directory
		files, err := filepath.Glob(filepath.Join(currentPath, "*.sln"))
		if err != nil {
			return nil, err
		}

		if len(files) > 0 {
			// Found a solution file
			return ParseSolution(files[0])
		}

		// Move up one directory
		parentPath := filepath.Dir(currentPath)
		if parentPath == currentPath {
			// Reached root directory
			return nil, fmt.Errorf("no solution file found")
		}
		currentPath = parentPath
	}
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
		Path:          solutionPath,
		Name:          solutionName,
		RootDirectory: solutionDir,
		Projects:      []ProjectInfo{},
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

	return info, nil
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
