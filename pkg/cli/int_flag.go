package cli

import (
	"fmt"
	"strconv"
)

// IntFlag holds data for the flags with integer values
type IntFlag struct {
	Description  string
	DefaultValue int
	Hidden       bool
	Required     bool

	Value int

	isSet bool
}

func (t *IntFlag) IsHidden() bool {
	return t.Hidden
}

func (t *IntFlag) IsSet() bool {
	return t.isSet
}

func (t *IntFlag) Set(value any) error {
	if value == nil {
		t.Value = t.DefaultValue
		t.isSet = false
	} else {
		v, ok := value.(int)
		if !ok {
			s, ok := value.(string)
			if !ok {
				return fmt.Errorf("unsupported type %T for integer flag", value)
			}

			i, err := strconv.Atoi(s)
			if err != nil {
				return fmt.Errorf("could not convert string %v into integer: %v", s, err)
			}

			v = i
		}

		if v == 0 {
			if v != t.DefaultValue {
				t.Value = v
			} else {
				t.Value = t.DefaultValue
			}

			return nil
		}

		t.Value = v
		t.isSet = true
	}

	return nil
}

func (t *IntFlag) Format() string {
	description := t.Description

	if t.DefaultValue != 0 {
		description += fmt.Sprintf(" (default %d)", t.DefaultValue)
	}

	return description
}

func (t *IntFlag) Validate() error {
	if t.Required && !t.isSet {
		return fmt.Errorf("required but not set")
	}

	return nil
}
