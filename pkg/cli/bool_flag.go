package cli

import (
	"fmt"
	"strconv"
)

// BoolFlag holds data for the flags with boolean values.
type BoolFlag struct {
	Description  string
	DefaultValue bool
	Hidden       bool
	Required     bool

	Value bool

	isSet bool
}

func (t *BoolFlag) IsHidden() bool {
	return t.Hidden
}

func (t *BoolFlag) IsSet() bool {
	return t.isSet
}

func (t *BoolFlag) Set(value any) error {
	if value == nil {
		t.Value = t.DefaultValue
		t.isSet = false
	} else {
		v, ok := value.(bool)
		if !ok {
			s, ok := value.(string)
			if !ok {
				return fmt.Errorf("unsupported type %T for boolean flag", value)
			}

			if env, ok := environmentLookup(s); ok {
				s = env
			}

			b, err := strconv.ParseBool(s)
			if err != nil {
				return fmt.Errorf("could not convert string %v into boolean: %v", s, err)
			}

			v = b
		}

		if !v {
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

func (t *BoolFlag) Format() string {
	return fmt.Sprintf("%s (default %v)", t.Description, t.DefaultValue)
}

func (t *BoolFlag) Validate() error {
	if t.Required && t.isSet {
		return fmt.Errorf("required but not set")
	}

	return nil
}
