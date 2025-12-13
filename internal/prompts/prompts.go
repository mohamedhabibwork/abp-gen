package prompts

import (
	"fmt"
	"strconv"
	"github.com/AlecAivazis/survey/v2"
)

// PromptText prompts for a text input
func PromptText(message string, defaultValue string) (string, error) {
	var result string
	prompt := &survey.Input{
		Message: message,
		Default: defaultValue,
	}
	if err := survey.AskOne(prompt, &result); err != nil {
		return "", err
	}
	return result, nil
}

// PromptConfirm prompts for a yes/no confirmation
func PromptConfirm(message string, defaultValue bool) (bool, error) {
	var result bool
	prompt := &survey.Confirm{
		Message: message,
		Default: defaultValue,
	}
	if err := survey.AskOne(prompt, &result); err != nil {
		return false, err
	}
	return result, nil
}

// PromptSelect prompts for a single selection from options
func PromptSelect(message string, options []string, defaultValue string) (string, error) {
	var result string
	prompt := &survey.Select{
		Message: message,
		Options: options,
		Default: defaultValue,
	}
	if err := survey.AskOne(prompt, &result); err != nil {
		return "", err
	}
	return result, nil
}

// PromptMultiSelect prompts for multiple selections from options
func PromptMultiSelect(message string, options []string, defaults []string) ([]string, error) {
	var result []string
	prompt := &survey.MultiSelect{
		Message: message,
		Options: options,
		Default: defaults,
	}
	if err := survey.AskOne(prompt, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// PromptInt prompts for an integer input
func PromptInt(message string, defaultValue int) (int, error) {
	var result string
	prompt := &survey.Input{
		Message: message,
		Default: fmt.Sprintf("%d", defaultValue),
	}
	if err := survey.AskOne(prompt, &result); err != nil {
		return 0, err
	}
	
	// Parse integer
	value, err := strconv.Atoi(result)
	if err != nil {
		return 0, fmt.Errorf("invalid integer: %w", err)
	}
	return value, nil
}

