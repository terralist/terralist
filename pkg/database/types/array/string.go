package array

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
)

// StringArray is a wrapper over the []string, to enable GORM support
// for the respective column type
type StringArray []string

// MarshalJSON writes the StringArray in a JSON format
func (t StringArray) MarshalJSON() ([]byte, error) {
	s := []string(t)

	b, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}

	return b, nil
}

// UnmarshalJSON unpacks an StringArray from a JSON value
func (t *StringArray) UnmarshalJSON(b []byte) error {
	var s []string

	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	*t = StringArray(s)

	return nil
}

// GormDataType represents the column data type to which the value should be
// mapped in a database
func (StringArray) GormDataType() string {
	return "binary(65535)"
}

// Scan tells GORM how to read the value from the database
func (t *StringArray) Scan(value interface{}) error {
	b, _ := value.([]byte)

	if err := t.UnmarshalJSON(b); err != nil {
		return err
	}

	return nil
}

// Value tells GORM how to write the value to the database
func (t StringArray) Value() (driver.Value, error) {
	return t.MarshalJSON()
}

// String converts the StringArray to a string
func (t StringArray) String() string {
	return fmt.Sprintf("[%s]", strings.Join(t, ", "))
}

// Empty checks if a StringArray is empty (is zero-value)
func (t StringArray) Empty() bool {
	return len(t) == 0
}
