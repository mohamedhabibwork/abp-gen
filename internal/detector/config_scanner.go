package detector

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// config_scanner.go provides configuration scanning capabilities
// to auto-detect multi-tenancy settings, microservice architecture,
// and other configuration from appsettings.json and module files.

// ConfigScanner scans configuration files to detect multi-tenancy and other settings
type ConfigScanner struct{}

// NewConfigScanner creates a new config scanner
func NewConfigScanner() *ConfigScanner {
	return &ConfigScanner{}
}

// AppSettings represents a simplified appsettings.json structure
type AppSettings struct {
	Abp struct {
		MultiTenancy struct {
			IsEnabled bool `json:"IsEnabled"`
		} `json:"MultiTenancy"`
	} `json:"Abp"`
	ConnectionStrings map[string]string `json:"ConnectionStrings"`
}

// DetectMultiTenancy scans for multi-tenancy configuration in the solution
// Returns: (enabled, strategy, error)
func (s *ConfigScanner) DetectMultiTenancy(solutionInfo *SolutionInfo) (enabled bool, strategy string, err error) {
	// Strategy 1: Check appsettings.json files
	appsettingsFiles := s.findAppSettingsFiles(solutionInfo.RootDirectory)
	for _, file := range appsettingsFiles {
		if isEnabled, strat := s.parseAppSettings(file); isEnabled {
			return true, strat, nil
		}
	}

	// Strategy 2: Check module files for [MultiTenant] attribute or IsMultiTenant property
	for _, project := range solutionInfo.Projects {
		if project.Type == ProjectTypeDomain || project.Type == ProjectTypeDomainShared {
			if isEnabled := s.scanModuleFiles(project.Directory); isEnabled {
				return true, "host", nil // Default to host strategy
			}
		}
	}

	// Strategy 3: Check for ConnectionString__<TenantName> pattern
	for _, file := range appsettingsFiles {
		if s.hasPerTenantConnectionStrings(file) {
			return true, "tenant-per-db", nil
		}
	}

	return false, "none", nil
}

// findAppSettingsFiles recursively finds appsettings*.json files
func (s *ConfigScanner) findAppSettingsFiles(rootDir string) []string {
	var files []string

	// Common locations for appsettings
	searchPaths := []string{
		filepath.Join(rootDir, "appsettings.json"),
		filepath.Join(rootDir, "src", "**", "appsettings*.json"),
		filepath.Join(rootDir, "**", "appsettings*.json"),
	}

	for _, pattern := range searchPaths {
		matches, err := filepath.Glob(pattern)
		if err == nil {
			files = append(files, matches...)
		}
	}

	// Also check each project directory
	filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && strings.HasPrefix(info.Name(), "appsettings") && strings.HasSuffix(info.Name(), ".json") {
			files = append(files, path)
		}
		return nil
	})

	return files
}

// parseAppSettings parses an appsettings.json file for multi-tenancy config
func (s *ConfigScanner) parseAppSettings(filePath string) (enabled bool, strategy string) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return false, ""
	}

	var settings AppSettings
	if err := json.Unmarshal(data, &settings); err != nil {
		return false, ""
	}

	if settings.Abp.MultiTenancy.IsEnabled {
		// Try to infer strategy from connection strings
		strategy = "host" // Default
		if s.hasMultipleConnectionStrings(settings.ConnectionStrings) {
			strategy = "tenant-per-db"
		}
		return true, strategy
	}

	return false, ""
}

// scanModuleFiles scans C# module files for multi-tenancy indicators
func (s *ConfigScanner) scanModuleFiles(directory string) bool {
	// Look for *Module.cs files
	moduleFiles, err := filepath.Glob(filepath.Join(directory, "*Module.cs"))
	if err != nil {
		return false
	}

	// Patterns to search for
	multiTenantPatterns := []*regexp.Regexp{
		regexp.MustCompile(`\[MultiTenant\]`),
		regexp.MustCompile(`IsMultiTenant\s*=\s*true`),
		regexp.MustCompile(`ConfigureMultiTenancy`),
		regexp.MustCompile(`\.AddAbpMultiTenancy`),
	}

	for _, file := range moduleFiles {
		content, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		contentStr := string(content)
		for _, pattern := range multiTenantPatterns {
			if pattern.MatchString(contentStr) {
				return true
			}
		}
	}

	return false
}

// hasPerTenantConnectionStrings checks for per-tenant connection string pattern
func (s *ConfigScanner) hasPerTenantConnectionStrings(filePath string) bool {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return false
	}

	var settings AppSettings
	if err := json.Unmarshal(data, &settings); err != nil {
		return false
	}

	return s.hasMultipleConnectionStrings(settings.ConnectionStrings)
}

// hasMultipleConnectionStrings checks if there are multiple connection strings
func (s *ConfigScanner) hasMultipleConnectionStrings(connStrings map[string]string) bool {
	// If more than one connection string, likely tenant-per-db
	return len(connStrings) > 1
}

// DetectMicroserviceMode determines if the solution follows microservice architecture
// by analyzing project structure and configuration
func (s *ConfigScanner) DetectMicroserviceMode(solutionInfo *SolutionInfo) (isMicroservice bool, services []string) {
	// Already detected by IsMicroserviceArchitecture
	if solutionInfo.IsMicroservice {
		return true, s.extractServiceNames(solutionInfo)
	}

	// Additional checks: look for microservice indicators in config
	for _, project := range solutionInfo.Projects {
		// Check for gateway indicators in project files
		if strings.Contains(strings.ToLower(project.Name), "gateway") ||
			strings.Contains(strings.ToLower(project.Name), "apigateway") {
			return true, s.extractServiceNames(solutionInfo)
		}

		// Check for Ocelot or YARP configuration (API Gateway libraries)
		if s.hasGatewayConfig(project.Directory) {
			return true, s.extractServiceNames(solutionInfo)
		}
	}

	return false, nil
}

// hasGatewayConfig checks for API gateway configuration files
func (s *ConfigScanner) hasGatewayConfig(directory string) bool {
	// Check for ocelot.json or yarp configuration
	gatewayConfigs := []string{
		filepath.Join(directory, "ocelot.json"),
		filepath.Join(directory, "appsettings.json"), // Contains YARP config
	}

	for _, configPath := range gatewayConfigs {
		if _, err := os.Stat(configPath); err == nil {
			data, err := os.ReadFile(configPath)
			if err == nil {
				content := string(data)
				// Check for gateway-specific configuration
				if strings.Contains(content, "\"Routes\"") ||
					strings.Contains(content, "\"ReverseProxy\"") ||
					strings.Contains(content, "\"GlobalConfiguration\"") {
					return true
				}
			}
		}
	}

	return false
}

// extractServiceNames extracts service names from microservice projects
func (s *ConfigScanner) extractServiceNames(solutionInfo *SolutionInfo) []string {
	var services []string
	seen := make(map[string]bool)

	for _, project := range solutionInfo.Projects {
		name := project.Name
		nameLower := strings.ToLower(name)

		// Extract service name from patterns like "MyApp.ServiceName.HttpApi"
		if strings.Contains(nameLower, "service") && !strings.Contains(nameLower, "shared") {
			// Extract the service part
			parts := strings.Split(name, ".")
			for _, part := range parts {
				partLower := strings.ToLower(part)
				if strings.Contains(partLower, "service") &&
					!strings.Contains(partLower, "shared") &&
					!seen[part] {
					services = append(services, part)
					seen[part] = true
				}
			}
		}
	}

	return services
}

// SummarizeConfiguration creates a human-readable summary of detected configuration
func (s *ConfigScanner) SummarizeConfiguration(solutionInfo *SolutionInfo) string {
	var summary strings.Builder

	// Multi-tenancy info
	enabled, strategy, _ := s.DetectMultiTenancy(solutionInfo)
	if enabled {
		summary.WriteString(fmt.Sprintf("  Multi-tenancy: Enabled (strategy: %s)\n", strategy))
	} else {
		summary.WriteString("  Multi-tenancy: Disabled\n")
	}

	// Microservice info
	isMicro, services := s.DetectMicroserviceMode(solutionInfo)
	if isMicro {
		summary.WriteString("  Architecture: Microservice")
		if len(services) > 0 {
			summary.WriteString(fmt.Sprintf(" (%d services detected)", len(services)))
		}
		summary.WriteString("\n")
	} else {
		summary.WriteString("  Architecture: Monolith\n")
	}

	return summary.String()
}
