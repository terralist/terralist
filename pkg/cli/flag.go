package cli

import (
	"github.com/mitchellh/mapstructure"
	"reflect"
)

type Flag interface {
	IsSet() bool
	IsHidden() bool

	Set(value any) error

	Format() string

	Validate() error
}

func FlagDecoder(result any) *mapstructure.Decoder {
	dec, _ := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result: result,
		DecodeHook: func(f reflect.Type, t reflect.Type, data any) (any, error) {
			if v, ok := data.(*StringFlag); ok {
				return v.Value, nil
			}

			if v, ok := data.(*BoolFlag); ok {
				return v.Value, nil
			}

			if v, ok := data.(*IntFlag); ok {
				return v.Value, nil
			}

			return nil, nil
		},
	})

	return dec
}
