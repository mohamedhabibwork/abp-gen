package generator

import (
	"bytes"
	"fmt"
	"path/filepath"
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

	// Choose template based on event type
	if event.Type == "domain" {
		tmpl, err = g.tmplLoader.Load("domain_event.tmpl")
	} else {
		tmpl, err = g.tmplLoader.Load("distributed_event.tmpl")
	}

	if err != nil {
		return fmt.Errorf("failed to load event template: %w", err)
	}

	data := map[string]interface{}{
		"SolutionName":    sch.Solution.Name,
		"ModuleName":      sch.Solution.ModuleName,
		"NamespaceRoot":   sch.Solution.NamespaceRoot,
		"EntityName":      entity.Name,
		"EventName":       event.Name,
		"EventType":       event.Type,
		"Payload":         event.Payload,
		"Description":     event.Description,
		"TargetFramework": sch.Solution.TargetFramework,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute event template: %w", err)
	}

	moduleFolder := sch.Solution.ModuleName + "Module"

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

	// Choose template based on handler type
	switch handler.HandlerType {
	case "local":
		tmpl, err = g.tmplLoader.Load("event_handler_local.tmpl")
	case "distributed":
		tmpl, err = g.tmplLoader.Load("event_handler_distributed.tmpl")
	case "integration":
		tmpl, err = g.tmplLoader.Load("event_handler_integration.tmpl")
	default:
		tmpl, err = g.tmplLoader.Load("event_handler_local.tmpl")
	}

	if err != nil {
		return fmt.Errorf("failed to load event handler template: %w", err)
	}

	data := map[string]interface{}{
		"SolutionName":    sch.Solution.Name,
		"ModuleName":      sch.Solution.ModuleName,
		"NamespaceRoot":   sch.Solution.NamespaceRoot,
		"EntityName":      entity.Name,
		"EventName":       event.Name,
		"HandlerName":     handler.Name,
		"HandlerType":     handler.HandlerType,
		"Action":          handler.Action,
		"TargetFramework": sch.Solution.TargetFramework,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute event handler template: %w", err)
	}

	moduleFolder := sch.Solution.ModuleName + "Module"
	handlerPath := filepath.Join(paths.ApplicationEventHandlers, moduleFolder, handler.Name+".cs")
	return g.writer.WriteFile(handlerPath, buf.String())
}
