package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

const (
	separator = ","
	maxLength = 4086
)

type StringPredicate = func(v string) bool

type StringArray struct {
	values []string
}

func NewStringArray(arr []string) *StringArray {
	return &StringArray{
		values: arr,
	}
}

// Valid checks if an StringArray object meets the character limit criteria
func (arr StringArray) Valid() bool {
	count := 0
	for _, v := range arr.values {
		count += len(v)
	}

	count -= 1

	return count < maxLength
}

// Any checks if at least one value successfully passes a predicate
func (arr StringArray) Any(predicate StringPredicate) bool {
	for _, v := range arr.values {
		if r := predicate(v); r {
			return true
		}
	}

	return false
}

// All checks if all values successfully passes a predicate
func (arr StringArray) All(predicate StringPredicate) bool {
	for _, v := range arr.values {
		if r := predicate(v); !r {
			return false
		}
	}

	return true
}

// Contains checks if the array contains a given value
func (arr StringArray) Contains(v string) bool {
	return arr.Any(func(vv string) bool {
		return vv == v
	})
}

// GormDataType represents the column data type to which the value should be
// mapped in a database
func (arr StringArray) GormDataType() string {
	return fmt.Sprintf("string,size:%d", maxLength)
}

// Scan tells GORM how to read the value from the database
func (arr *StringArray) Scan(value interface{}) error {
	val, ok := value.(string)
	if !ok {
		return errors.Wrap(ErrInvalidType, "expected string")
	}

	arr.values = strings.Split(val, separator)

	return nil
}

// Value tells GORM how to write the value to the database
func (arr StringArray) Value() (driver.Value, error) {
	return strings.Join(arr.values, separator), nil
}

func (arr *StringArray) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var data []string
	if err := unmarshal(&data); err != nil {
		return fmt.Errorf("%w: %v", ErrUnparsable, err)
	}

	arr.values = data

	return nil
}

func (arr *StringArray) UnmarshalJSON(in []byte) error {
	var data []string
	if err := json.Unmarshal(in, &data); err != nil {
		return fmt.Errorf("%w: %v", ErrUnparsable, err)
	}

	arr.values = data

	return nil
}
