package provider

import (
	"testing"
	"time"

	"terralist/pkg/database/entity"
)

func TestProviderToArtifactTimestamps(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name     string
		location *time.Location
	}{
		{name: "UTC", location: time.UTC},
		{name: "UTC+8", location: time.FixedZone("UTC+8", 8*60*60)},
		{name: "UTC-5", location: time.FixedZone("UTC-5", -5*60*60)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			createdAt := time.Date(2026, time.November, 11, 8, 0, 0, 123456789, time.UTC)
			updatedAt := time.Date(2026, time.November, 12, 10, 0, 0, 987654321, time.UTC)
			p := Provider{Entity: entity.Entity{
				CreatedAt: createdAt.In(tc.location),
				UpdatedAt: updatedAt.In(tc.location),
			}}

			got := p.ToArtifact()
			for _, timestamp := range []struct {
				name  string
				value string
				want  time.Time
			}{
				{name: "created_at", value: got.CreatedAt, want: createdAt},
				{name: "updated_at", value: got.UpdatedAt, want: updatedAt},
			} {
				parsed, err := time.Parse(time.RFC3339Nano, timestamp.value)
				if err != nil {
					t.Errorf("%s = %q: timestamp must include a timezone: %v", timestamp.name, timestamp.value, err)
					continue
				}
				if !parsed.Equal(timestamp.want) {
					t.Errorf("%s represents %v, want %v", timestamp.name, parsed, timestamp.want)
				}
			}
		})
	}
}
