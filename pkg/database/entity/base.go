package entity

import (
	"time"

	"terralist/pkg/database"

	"github.com/google/uuid"
)

type Entity struct {
	ID        uuid.UUID `gorm:"primary_key;"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (e *Entity) BeforeCreate(*database.DB) error {
	if e.Empty() {
		var err error
		e.ID, err = uuid.NewRandom()
		return err
	}

	return nil
}

// Empty checks if an entity is empty.
// It assumes that an empty entity holds the zero-value ID.
func (e Entity) Empty() bool {
	return e.ID == uuid.MustParse("00000000-0000-0000-0000-000000000000")
}
