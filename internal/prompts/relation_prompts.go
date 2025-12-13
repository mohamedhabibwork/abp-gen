package prompts

import (
	"fmt"

	"github.com/mohamedhabibwork/abp-gen/internal/schema"
	"github.com/mohamedhabibwork/abp-gen/internal/templates"
)

// PromptRelations prompts for entity relations
func PromptRelations() (*schema.Relations, error) {
	hasRelations, err := PromptConfirm("Add relations?", false)
	if err != nil {
		return nil, err
	}

	if !hasRelations {
		return nil, nil
	}

	relations := &schema.Relations{}

	// One-to-Many relations
	addOneToMany, err := PromptConfirm("Add One-to-Many relation?", false)
	if err != nil {
		return nil, err
	}

	if addOneToMany {
		oneToManyRels, err := PromptOneToManyRelations()
		if err != nil {
			return nil, err
		}
		relations.OneToMany = oneToManyRels
	}

	// Many-to-Many relations
	addManyToMany, err := PromptConfirm("Add Many-to-Many relation?", false)
	if err != nil {
		return nil, err
	}

	if addManyToMany {
		manyToManyRels, err := PromptManyToManyRelations()
		if err != nil {
			return nil, err
		}
		relations.ManyToMany = manyToManyRels
	}

	return relations, nil
}

// PromptOneToManyRelations prompts for one-to-many relations
func PromptOneToManyRelations() ([]schema.OneToManyRelation, error) {
	var relations []schema.OneToManyRelation

	for {
		rel, err := PromptOneToManyRelation()
		if err != nil {
			return nil, err
		}

		relations = append(relations, *rel)

		addMore, err := PromptConfirm("Add another One-to-Many relation?", false)
		if err != nil {
			return nil, err
		}

		if !addMore {
			break
		}
	}

	return relations, nil
}

// PromptOneToManyRelation prompts for a single one-to-many relation
func PromptOneToManyRelation() (*schema.OneToManyRelation, error) {
	targetEntity, err := PromptText("Target entity name:", "")
	if err != nil {
		return nil, err
	}

	defaultFKName := targetEntity + "Id"
	foreignKeyName, err := PromptText("Foreign key name:", defaultFKName)
	if err != nil {
		return nil, err
	}

	defaultNavProp := templates.Pluralize(targetEntity)
	navigationProperty, err := PromptText("Navigation property name:", defaultNavProp)
	if err != nil {
		return nil, err
	}

	return &schema.OneToManyRelation{
		TargetEntity:       targetEntity,
		ForeignKeyName:     foreignKeyName,
		NavigationProperty: navigationProperty,
		IsCollection:       true,
	}, nil
}

// PromptManyToManyRelations prompts for many-to-many relations
func PromptManyToManyRelations() ([]schema.ManyToManyRelation, error) {
	var relations []schema.ManyToManyRelation

	for {
		rel, err := PromptManyToManyRelation()
		if err != nil {
			return nil, err
		}

		relations = append(relations, *rel)

		addMore, err := PromptConfirm("Add another Many-to-Many relation?", false)
		if err != nil {
			return nil, err
		}

		if !addMore {
			break
		}
	}

	return relations, nil
}

// PromptManyToManyRelation prompts for a single many-to-many relation
func PromptManyToManyRelation() (*schema.ManyToManyRelation, error) {
	targetEntity, err := PromptText("Target entity name:", "")
	if err != nil {
		return nil, err
	}

	// Default join entity name (will be sorted in validation)
	defaultJoinEntity := fmt.Sprintf("%sJoin", targetEntity)
	joinEntity, err := PromptText("Join entity name:", defaultJoinEntity)
	if err != nil {
		return nil, err
	}

	defaultNavProp := templates.Pluralize(targetEntity)
	navigationProperty, err := PromptText("Navigation property name:", defaultNavProp)
	if err != nil {
		return nil, err
	}

	return &schema.ManyToManyRelation{
		TargetEntity:       targetEntity,
		JoinEntity:         joinEntity,
		NavigationProperty: navigationProperty,
	}, nil
}
