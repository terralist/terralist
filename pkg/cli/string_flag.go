package cli

import (
	"fmt"
	"strings"

	"github.com/ssoroka/slice"
)

// StringFlag holds data for the flags with string values
type StringFlag struct {
	Description  string
	Choices      []string
	DefaultValue string
	Hidden       bool
	Required     bool

	Value string

	isSet bool
}

func (t *StringFlag) IsHidden() bool {
	return t.Hidden
}

func (t *StringFlag) IsSet() bool {
	return t.isSet
}

func (t *StringFlag) Set(value any) error {
	if value == nil {
		t.Value = t.DefaultValue
		t.isSet = false
	} else {
		v, ok := value.(string)
		if !ok {
			return fmt.Errorf("type %T is not string", value)
		}

		if v == "" {
			if v != t.DefaultValue {
				t.Value = v
			} else {
				t.Value = t.DefaultValue
			}

			return nil
		}

		if len(t.Choices) > 0 {
			lv := strings.ToLower(v)

			if !slice.Contains(t.Choices, lv) {
				options := strings.Join(t.Choices, ", ")
				return fmt.Errorf("value (%v) must be one of the values: %s", value, options)
			}
		}

		t.Value = v
		t.isSet = true
	}

	return nil
}

func (t *StringFlag) Format() string {
	description := t.Description

	if len(t.Choices) > 0 {
		options := strings.Join(t.Choices, ", ")
		description += fmt.Sprintf(" Options: [%s]", options)
	}

	if t.DefaultValue != "" {
		description += fmt.Sprintf(" (default %q)", t.DefaultValue)
	}

	return description
}

func (t *StringFlag) Validate() error {
	if t.Required && !t.isSet {
		return fmt.Errorf("required but not set")
	}

	lv := strings.ToLower(t.Value)

	if len(t.Choices) > 0 {
		if !slice.Contains(t.Choices, lv) {
			options := strings.Join(t.Choices, ", ")
			return fmt.Errorf("invalid value %v; must be one of: %s", options)
		}
	}

	return nil
}
