package detector

import (
	"testing"
)

func TestNormalizeDotNetVersion(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"net9.0", "net9.0", "9"},
		{"net8.0", "net8.0", "8"},
		{"net10.0", "net10.0", "10"},
		{"NET9.0 uppercase", "NET9.0", "9"},
		{"with extra spaces", "  net9.0  ", "9"},
		{"netstandard2.0", "netstandard2.0", ""},
		{"empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeDotNetVersion(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeDotNetVersion(%q) = %q; want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNormalizeABPVersion(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"ABP 8.x", "8.3.0", "8"},
		{"ABP 9.x", "9.0.0", "9"},
		{"ABP 10.x", "10.0.0", "10"},
		{"ABP 10.x RC", "10.0.0-rc.1", "10"},
		{"ABP with preview", "9.0.0-preview.1", "9"},
		{"empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeABPVersion(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeABPVersion(%q) = %q; want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestMapToTargetFramework(t *testing.T) {
	tests := []struct {
		name           string
		abpVersion     string
		dotnetVersion  string
		isMicroservice bool
		expected       string
	}{
		{"ABP 10 Monolith", "10", "10", false, "abp10-monolith"},
		{"ABP 10 Microservice", "10", "10", true, "abp10-microservice"},
		{"ABP 9 Monolith", "9", "9", false, "abp9-monolith"},
		{"ABP 9 Microservice", "9", "9", true, "abp9-microservice"},
		{"ABP 8 Monolith", "8", "8", false, "abp8-monolith"},
		{"ABP 8 Microservice", "8", "8", true, "abp8-microservice"},
		{"ASP.NET Core 10", "", "10", false, "aspnetcore10"},
		{"ASP.NET Core 9", "", "9", false, "aspnetcore9"},
		{"No version info", "", "", false, "aspnetcore9"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MapToTargetFramework(tt.abpVersion, tt.dotnetVersion, tt.isMicroservice)
			if result != tt.expected {
				t.Errorf("MapToTargetFramework(%q, %q, %v) = %q; want %q",
					tt.abpVersion, tt.dotnetVersion, tt.isMicroservice, result, tt.expected)
			}
		})
	}
}

func TestDetermineProjectType(t *testing.T) {
	tests := []struct {
		name         string
		projectName  string
		expectedType ProjectType
	}{
		{"Domain project", "MyApp.Domain", ProjectTypeDomain},
		{"Domain.Shared project", "MyApp.Domain.Shared", ProjectTypeDomainShared},
		{"Application project", "MyApp.Application", ProjectTypeApplication},
		{"Application.Contracts", "MyApp.Application.Contracts", ProjectTypeApplicationContracts},
		{"HttpApi project", "MyApp.HttpApi", ProjectTypeHttpApi},
		{"EF Core project", "MyApp.EntityFrameworkCore", ProjectTypeEntityFrameworkCore},
		{"MongoDB project", "MyApp.MongoDB", ProjectTypeMongoDB},
		{"Unknown project", "MyApp.Web", ProjectTypeUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetermineProjectType(tt.projectName)
			if result != tt.expectedType {
				t.Errorf("DetermineProjectType(%q) = %q; want %q", tt.projectName, result, tt.expectedType)
			}
		})
	}
}

func TestIsMicroserviceArchitecture(t *testing.T) {
	tests := []struct {
		name     string
		solution *SolutionInfo
		expected bool
	}{
		{
			name: "Microservice with gateway",
			solution: &SolutionInfo{
				Projects: []ProjectInfo{
					{Name: "MyApp.Gateway"},
					{Name: "MyApp.ServiceA"},
					{Name: "MyApp.ServiceB"},
					{Name: "MyApp.Shared"},
				},
			},
			expected: true,
		},
		{
			name: "Microservice with multiple services",
			solution: &SolutionInfo{
				Projects: []ProjectInfo{
					{Name: "MyApp.ServiceA"},
					{Name: "MyApp.ServiceB"},
					{Name: "MyApp.ServiceC"},
				},
			},
			expected: true,
		},
		{
			name: "Monolith",
			solution: &SolutionInfo{
				Projects: []ProjectInfo{
					{Name: "MyApp.Domain"},
					{Name: "MyApp.Application"},
					{Name: "MyApp.HttpApi"},
				},
			},
			expected: false,
		},
		{
			name: "Single service not microservice",
			solution: &SolutionInfo{
				Projects: []ProjectInfo{
					{Name: "MyApp.Service"},
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsMicroserviceArchitecture(tt.solution)
			if result != tt.expected {
				t.Errorf("IsMicroserviceArchitecture() = %v; want %v", result, tt.expected)
			}
		})
	}
}
