package generator

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/mohamedhabibwork/abp-gen/internal/detector"
	"github.com/mohamedhabibwork/abp-gen/internal/schema"
	"github.com/mohamedhabibwork/abp-gen/internal/templates"
	"github.com/mohamedhabibwork/abp-gen/internal/writer"
)

// DomainEventsGenerator generates domain event classes and handlers
type DomainEventsGenerator struct {
	tmplLoader *templates.Loader
	writer     *writer.Writer
}

// NewDomainEventsGenerator creates a new domain events generator
func NewDomainEventsGenerator(tmplLoader *templates.Loader, w *writer.Writer) *DomainEventsGenerator {
	return &DomainEventsGenerator{
		tmplLoader: tmplLoader,
		writer:     w,
	}
}

// Generate generates domain event definitions
func (g *DomainEventsGenerator) Generate(sch *schema.Schema, entity *schema.Entity, paths *detector.LayerPaths) error {
	if len(entity.DomainEvents) == 0 {
		return nil
	}

	for _, event := range entity.DomainEvents {
		if err := g.generateEventClass(sch, entity, &event, paths); err != nil {
			return fmt.Errorf("failed to generate event %s: %w", event.Name, err)
		}

		// Generate handlers if defined
		for _, handler := range event.Handlers {
			if err := g.generateEventHandler(sch, entity, &event, &handler, paths); err != nil {
				return fmt.Errorf("failed to generate handler %s: %w", handler.Name, err)
			}
		}
	}

	return nil
}

func (g *DomainEventsGenerator) generateEventClass(sch *schema.Schema, entity *schema.Entity, event *schema.DomainEvent, paths *detector.LayerPaths) error {
	var tmpl *template.Template
	var err error
	templateName := ""

	// Choose template based on event type
	if event.Type == "domain" {
		templateName = "domain_event.tmpl"
	} else {
		templateName = "distributed_event.tmpl"
	}

	tmpl, err = g.tmplLoader.Load(templateName)
	if err != nil {
		// If template not found, skip event generation gracefully
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "file does not exist") {
			return nil // Skip event generation if template doesn't exist
		}
		return fmt.Errorf("failed to load event template '%s': %w", templateName, err)
	}

	data := map[string]interface{}{
		"SolutionName":         sch.Solution.Name,
		"ModuleName":           sch.Solution.ModuleName,
		"ModuleNameWithSuffix": sch.Solution.GetModuleNameWithSuffix(),
		"NamespaceRoot":        sch.Solution.NamespaceRoot,
		"EntityName":           entity.Name,
		"EventName":            event.Name,
		"EventType":            event.Type,
		"Payload":              event.Payload,
		"Description":          event.Description,
		"TargetFramework":      sch.Solution.TargetFramework,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute event template: %w", err)
	}

	moduleFolder := sch.Solution.GetModuleFolderName()

	var eventPath string
	if event.Type == "domain" {
		// Domain events go in Domain layer
		eventPath = filepath.Join(paths.Domain, "Events", moduleFolder, event.Name+".cs")
	} else {
		// Distributed events go in Domain.Shared layer
		eventPath = filepath.Join(paths.DomainSharedEvents, moduleFolder, event.Name+".cs")
	}

	return g.writer.WriteFile(eventPath, buf.String())
}

func (g *DomainEventsGenerator) generateEventHandler(sch *schema.Schema, entity *schema.Entity, event *schema.DomainEvent, handler *schema.EventHandler, paths *detector.LayerPaths) error {
	var tmpl *template.Template
	var err error
	templateName := ""

	// Choose template based on handler type
	switch handler.HandlerType {
	case "local":
		templateName = "event_handler_local.tmpl"
	case "distributed":
		templateName = "event_handler_distributed.tmpl"
	case "integration":
		templateName = "event_handler_integration.tmpl"
	default:
		templateName = "event_handler_local.tmpl"
	}

	tmpl, err = g.tmplLoader.Load(templateName)
	if err != nil {
		// If template not found, skip handler generation gracefully
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "file does not exist") {
			return nil // Skip handler generation if template doesn't exist
		}
		return fmt.Errorf("failed to load event handler template '%s': %w", templateName, err)
	}

	data := map[string]interface{}{
		"SolutionName":         sch.Solution.Name,
		"ModuleName":           sch.Solution.ModuleName,
		"ModuleNameWithSuffix": sch.Solution.GetModuleNameWithSuffix(),
		"NamespaceRoot":        sch.Solution.NamespaceRoot,
		"EntityName":           entity.Name,
		"EventName":            event.Name,
		"HandlerName":          handler.Name,
		"HandlerType":          handler.HandlerType,
		"Action":               handler.Action,
		"TargetFramework":      sch.Solution.TargetFramework,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute event handler template: %w", err)
	}

	moduleFolder := sch.Solution.GetModuleFolderName()
	handlerPath := filepath.Join(paths.ApplicationEventHandlers, moduleFolder, handler.Name+".cs")
	return g.writer.WriteFile(handlerPath, buf.String())
}
