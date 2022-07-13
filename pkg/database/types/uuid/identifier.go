package uuid

import (
	"bytes"
	"database/sql/driver"
	"fmt"

	_uuid "github.com/google/uuid"
)

// ID is a wrapper over the google/uuid package, to enable GORM support
// for the respective column type
type ID _uuid.UUID

var (
	// emptyID holds the zero-value of the ID type
	emptyID = MustParse("00000000-0000-0000-0000-000000000000")
)

// Parse returns an ID parsed from a string value
// It returns an error in case the string value cannot be parsed
func Parse(s string) (ID, error) {
	uid, err := _uuid.Parse(s)
	if err != nil {
		return emptyID, fmt.Errorf("could not parse identifier: %v", err)
	}

	return ID(uid), nil
}

// MustParse returns an ID parsed from a string value
// It panics in case the string value cannot be parsed
// Recommended only in cases where it's known that the string value is
// well-formatted
func MustParse(s string) ID {
	return ID(_uuid.MustParse(s))
}

// FromBytes creates a new ID from a byte slice
func FromBytes(b []byte) (ID, error) {
	uid, err := _uuid.FromBytes(b)
	return ID(uid), err
}

// NewRandom returns a new random ID
func NewRandom() (ID, error) {
	uid, err := _uuid.NewRandom()
	return ID(uid), err
}

// MarshalJSON writes the ID in a JSON format
func (t ID) MarshalJSON() ([]byte, error) {
	s := _uuid.UUID(t)
	j := fmt.Sprintf("\"%s\"", s.String())

	return []byte(j), nil
}

// UnmarshalJSON unpacks an ID from a JSON value
func (t *ID) UnmarshalJSON(b []byte) error {
	s, err := _uuid.ParseBytes(b)
	if err != nil {
		return err
	}

	*t = ID(s)

	return nil
}

// GormDataType represents the column data type to which the value should be
// mapped in a database
func (ID) GormDataType() string {
	return "binary(16)"
}

// Scan tells GORM how to read the value from the database
func (t *ID) Scan(value interface{}) error {
	b, _ := value.([]byte)

	uid, err := _uuid.FromBytes(b)
	if err != nil {
		return err
	}

	*t = ID(uid)

	return nil
}

// Value tells GORM how to write the value to the database
func (t ID) Value() (driver.Value, error) {
	return t.MarshalBinary()
}

// MarshalBinary converts the ID to a slice of bytes
func (t ID) MarshalBinary() ([]byte, error) {
	return _uuid.UUID(t).MarshalBinary()
}

// String converts the ID to a string
func (t ID) String() string {
	return _uuid.UUID(t).String()
}

// Empty checks if an ID is empty (is zero-value)
func (t ID) Empty() bool {
	this, _ := t.MarshalBinary()
	other, _ := emptyID.MarshalBinary()

	return bytes.Compare(this, other) == 0
}
