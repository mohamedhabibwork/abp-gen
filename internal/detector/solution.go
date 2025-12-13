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
func parseProjectLine(line string, solutionDir string) *ProjectInfo {
	// Extract project name and path
	parts := strings.Split(line, "\"")
	if len(parts) < 5 {
		return nil
	}

	projectName := parts[1]
	projectPath := parts[3]

	// Convert relative path to absolute
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
func DetermineProjectType(projectName string) ProjectType {
	switch {
	case strings.HasSuffix(projectName, ".Domain.Shared"):
		return ProjectTypeDomainShared
	case strings.HasSuffix(projectName, ".Domain"):
		return ProjectTypeDomain
	case strings.HasSuffix(projectName, ".Application.Contracts"):
		return ProjectTypeApplicationContracts
	case strings.HasSuffix(projectName, ".Application"):
		return ProjectTypeApplication
	case strings.HasSuffix(projectName, ".HttpApi"):
		return ProjectTypeHttpApi
	case strings.HasSuffix(projectName, ".EntityFrameworkCore"):
		return ProjectTypeEntityFrameworkCore
	case strings.HasSuffix(projectName, ".MongoDB"):
		return ProjectTypeMongoDB
	default:
		return ProjectTypeUnknown
	}
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
