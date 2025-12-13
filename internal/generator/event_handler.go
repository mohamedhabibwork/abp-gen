package generator

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mohamedhabibwork/abp-gen/internal/detector"
	"github.com/mohamedhabibwork/abp-gen/internal/schema"
	"github.com/mohamedhabibwork/abp-gen/internal/templates"
	"github.com/mohamedhabibwork/abp-gen/internal/writer"
)

// EventHandlerGenerator generates distributed event handlers
type EventHandlerGenerator struct {
	tmplLoader *templates.Loader
	writer     *writer.Writer
}

// NewEventHandlerGenerator creates a new event handler generator
func NewEventHandlerGenerator(tmplLoader *templates.Loader, w *writer.Writer) *EventHandlerGenerator {
	return &EventHandlerGenerator{
		tmplLoader: tmplLoader,
		writer:     w,
	}
}

// Generate generates event handlers for an entity
func (g *EventHandlerGenerator) Generate(sch *schema.Schema, entity *schema.Entity, paths *detector.LayerPaths) error {
	if !sch.Options.GenerateEventHandlers {
		return nil
	}

	if entity.EntityType == "ValueObject" || entity.EntityType == "Entity" {
		return nil // These types don't generate distributed events
	}

	// Ensure event handlers directory exists
	moduleFolder := sch.Solution.ModuleName + "Module"
	handlersPath := filepath.Join(paths.Application, "EventHandlers", moduleFolder)
	if err := os.MkdirAll(handlersPath, 0755); err != nil {
		return fmt.Errorf("failed to create event handlers directory: %w", err)
	}

	// Generate Created event handler
	if err := g.GenerateCreatedHandler(sch, entity, paths); err != nil {
		return err
	}

	// Generate Updated event handler
	if err := g.GenerateUpdatedHandler(sch, entity, paths); err != nil {
		return err
	}

	// Generate Deleted event handler
	return g.GenerateDeletedHandler(sch, entity, paths)
}

// GenerateCreatedHandler generates the Created event handler
func (g *EventHandlerGenerator) GenerateCreatedHandler(sch *schema.Schema, entity *schema.Entity, paths *detector.LayerPaths) error {
	tmpl, err := g.tmplLoader.Load("event_handler_created.tmpl")
	if err != nil {
		return fmt.Errorf("failed to load created event handler template: %w", err)
	}

	data := g.prepareEventHandlerData(sch, entity, "Created")

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute created event handler template: %w", err)
	}

	moduleFolder := sch.Solution.ModuleName + "Module"
	handlersPath := filepath.Join(paths.Application, "EventHandlers", moduleFolder)
	filePath := filepath.Join(handlersPath, entity.Name+"CreatedEventHandler.cs")
	return g.writer.WriteFile(filePath, buf.String())
}

// GenerateUpdatedHandler generates the Updated event handler
func (g *EventHandlerGenerator) GenerateUpdatedHandler(sch *schema.Schema, entity *schema.Entity, paths *detector.LayerPaths) error {
	tmpl, err := g.tmplLoader.Load("event_handler_updated.tmpl")
	if err != nil {
		return fmt.Errorf("failed to load updated event handler template: %w", err)
	}

	data := g.prepareEventHandlerData(sch, entity, "Updated")

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute updated event handler template: %w", err)
	}

	moduleFolder := sch.Solution.ModuleName + "Module"
	handlersPath := filepath.Join(paths.Application, "EventHandlers", moduleFolder)
	filePath := filepath.Join(handlersPath, entity.Name+"UpdatedEventHandler.cs")
	return g.writer.WriteFile(filePath, buf.String())
}

// GenerateDeletedHandler generates the Deleted event handler
func (g *EventHandlerGenerator) GenerateDeletedHandler(sch *schema.Schema, entity *schema.Entity, paths *detector.LayerPaths) error {
	tmpl, err := g.tmplLoader.Load("event_handler_deleted.tmpl")
	if err != nil {
		return fmt.Errorf("failed to load deleted event handler template: %w", err)
	}

	data := g.prepareEventHandlerData(sch, entity, "Deleted")

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute deleted event handler template: %w", err)
	}

	moduleFolder := sch.Solution.ModuleName + "Module"
	handlersPath := filepath.Join(paths.Application, "EventHandlers", moduleFolder)
	filePath := filepath.Join(handlersPath, entity.Name+"DeletedEventHandler.cs")
	return g.writer.WriteFile(filePath, buf.String())
}

// prepareEventHandlerData prepares data for event handler templates
func (g *EventHandlerGenerator) prepareEventHandlerData(sch *schema.Schema, entity *schema.Entity, eventType string) map[string]interface{} {
	return map[string]interface{}{
		"SolutionName":  sch.Solution.Name,
		"ModuleName":    sch.Solution.ModuleName,
		"NamespaceRoot": sch.Solution.NamespaceRoot,
		"EntityName":    entity.Name,
		"EventType":     eventType,
	}
}
