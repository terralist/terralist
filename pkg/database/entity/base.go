package entity

import (
	"terralist/pkg/database/types/uuid"
	"time"

	"terralist/pkg/database"
)

type Entity struct {
	ID        uuid.ID `gorm:"primary_key;"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (e *Entity) BeforeCreate(*database.DB) error {
	if e.ID.Empty() {
		var err error
		e.ID, err = uuid.NewRandom()
		return err
	}

	return nil
}

// Empty checks if an entity is empty
// It assumes that an empty entity holds the zero-value ID
func (e Entity) Empty() bool {
	return e.ID.Empty()
}
