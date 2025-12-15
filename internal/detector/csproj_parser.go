package detector

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// csproj_parser.go contains utilities for parsing .csproj files
// to extract target framework versions, package references, and project metadata.

// CsprojProject represents a parsed .csproj file
type CsprojProject struct {
	XMLName       xml.Name        `xml:"Project"`
	SDK           string          `xml:"Sdk,attr"`
	PropertyGroup []PropertyGroup `xml:"PropertyGroup"`
	ItemGroup     []ItemGroup     `xml:"ItemGroup"`
}

// PropertyGroup contains project properties
type PropertyGroup struct {
	TargetFramework  string `xml:"TargetFramework"`
	TargetFrameworks string `xml:"TargetFrameworks"`
	RootNamespace    string `xml:"RootNamespace"`
	IsPackable       string `xml:"IsPackable"`
}

// ItemGroup contains project references and packages
type ItemGroup struct {
	PackageReference []PackageReference `xml:"PackageReference"`
	ProjectReference []ProjectReference `xml:"ProjectReference"`
}

// PackageReference represents a NuGet package reference
type PackageReference struct {
	Include string `xml:"Include,attr"`
	Version string `xml:"Version,attr"`
}

// ProjectReference represents a project reference
type ProjectReference struct {
	Include string `xml:"Include,attr"`
}

// parseCsprojFile parses a .csproj file and extracts project information
func parseCsprojFile(csprojPath string) *ProjectInfo {
	data, err := os.ReadFile(csprojPath)
	if err != nil {
		return nil
	}

	var project CsprojProject
	if err := xml.Unmarshal(data, &project); err != nil {
		return nil
	}

	projectDir := filepath.Dir(csprojPath)
	projectName := strings.TrimSuffix(filepath.Base(csprojPath), ".csproj")

	// Determine project type from name and structure
	projectType := DetermineProjectType(projectName)

	return &ProjectInfo{
		Name:      projectName,
		Path:      csprojPath,
		Directory: projectDir,
		Type:      projectType,
	}
}

// DetectABPVersion detects ABP framework version from csproj package references
func DetectABPVersion(csprojPath string) string {
	data, err := os.ReadFile(csprojPath)
	if err != nil {
		return ""
	}

	var project CsprojProject
	if err := xml.Unmarshal(data, &project); err != nil {
		return ""
	}

	// Look for Volo.Abp package references
	abpVersionRegex := regexp.MustCompile(`Volo\.Abp`)
	for _, itemGroup := range project.ItemGroup {
		for _, pkg := range itemGroup.PackageReference {
			if abpVersionRegex.MatchString(pkg.Include) && pkg.Version != "" {
				return normalizeABPVersion(pkg.Version)
			}
		}
	}

	return ""
}

// DetectDotNetVersion detects .NET target framework from csproj
func DetectDotNetVersion(csprojPath string) string {
	data, err := os.ReadFile(csprojPath)
	if err != nil {
		return ""
	}

	var project CsprojProject
	if err := xml.Unmarshal(data, &project); err != nil {
		return ""
	}

	// Check for TargetFramework or TargetFrameworks
	for _, propGroup := range project.PropertyGroup {
		if propGroup.TargetFramework != "" {
			return normalizeDotNetVersion(propGroup.TargetFramework)
		}
		if propGroup.TargetFrameworks != "" {
			// Take the first one if multiple targets
			frameworks := strings.Split(propGroup.TargetFrameworks, ";")
			if len(frameworks) > 0 {
				return normalizeDotNetVersion(frameworks[0])
			}
		}
	}

	return ""
}

// normalizeABPVersion extracts major version from ABP version string
// e.g., "8.3.0" -> "8", "10.0.0-rc.1" -> "10"
func normalizeABPVersion(version string) string {
	// Extract major version
	parts := strings.Split(version, ".")
	if len(parts) > 0 {
		majorVersion := strings.Split(parts[0], "-")[0]
		return majorVersion
	}
	return ""
}

// normalizeDotNetVersion converts .NET framework string to version number
// e.g., "net9.0" -> "9", "net8.0" -> "8", "net10.0" -> "10"
func normalizeDotNetVersion(framework string) string {
	framework = strings.ToLower(strings.TrimSpace(framework))

	// Match patterns like net9.0, net8.0, net10.0
	re := regexp.MustCompile(`net(\d+)\.?\d*`)
	matches := re.FindStringSubmatch(framework)
	if len(matches) > 1 {
		return matches[1]
	}

	return ""
}

// MapToTargetFramework maps ABP and .NET versions to TargetFramework
func MapToTargetFramework(abpVersion, dotnetVersion string, isMicroservice bool) string {
	// If no ABP detected, it's pure ASP.NET Core
	if abpVersion == "" {
		if dotnetVersion == "10" {
			return "aspnetcore10"
		}
		if dotnetVersion == "9" {
			return "aspnetcore9"
		}
		return "aspnetcore9" // Default
	}

	// Map ABP version to target framework
	suffix := "-monolith"
	if isMicroservice {
		suffix = "-microservice"
	}

	switch abpVersion {
	case "10":
		return fmt.Sprintf("abp10%s", suffix)
	case "9":
		return fmt.Sprintf("abp9%s", suffix)
	case "8":
		return fmt.Sprintf("abp8%s", suffix)
	default:
		return fmt.Sprintf("abp8%s", suffix) // Default to ABP 8
	}
}

// ScanProjectsForVersions scans all projects to determine versions
func ScanProjectsForVersions(info *SolutionInfo) (abpVersion, dotnetVersion string) {
	for _, project := range info.Projects {
		if project.Path == "" {
			continue
		}

		// Try to detect ABP version
		if abpVersion == "" {
			abpVersion = DetectABPVersion(project.Path)
		}

		// Try to detect .NET version
		if dotnetVersion == "" {
			dotnetVersion = DetectDotNetVersion(project.Path)
		}

		// Stop if both found
		if abpVersion != "" && dotnetVersion != "" {
			break
		}
	}

	return
}
