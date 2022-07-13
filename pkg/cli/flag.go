package cli

type Flag interface {
	IsSet() bool
	IsHidden() bool

	Set(value any) error

	Format() string

	Validate() error
}
