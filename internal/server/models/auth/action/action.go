package action

import (
	"fmt"

	"errors"
)

var (
	ErrUnknownTarget    = errors.New("unknown target")
	ErrUnknownOperation = errors.New("unknown operation")
	ErrInvalidAction    = errors.New("invalid action")
)

type Target int

const (
	Modules Target = iota
	Providers
	Authorities
	ApiKeys
	Any
)

func NewTarget(t string) (Target, error) {
	switch t {
	case "modules":
		return Modules, nil
	case "providers":
		return Providers, nil
	case "authorities":
		return Authorities, nil
	case "apiKeys":
		return ApiKeys, nil
	case "*":
		return Any, nil
	}

	return Target(0), fmt.Errorf("%w: %s", ErrUnknownTarget, t)
}

func (t Target) String() string {
	return [...]string{"", "modules", "providers", "authorities", "apiKeys"}[t]
}

func (t Target) Valid() bool {
	return t.String() != ""
}

type Operation int

const (
	View Operation = iota
	Create
	Update
	Delete
	All
)

func NewOperation(o string) (Operation, error) {
	switch o {
	case "get":
		return View, nil
	case "post":
		return Create, nil
	case "patch":
		return Update, nil
	case "delete":
		return Delete, nil
	case "*":
		return All, nil
	}

	return Operation(0), fmt.Errorf("%w: %s", ErrUnknownTarget, o)
}

func (op Operation) String() string {
	return [...]string{"", "get", "post", "patch", "delete"}[op]
}

func (op Operation) Valid() bool {
	return op.String() != ""
}

type Action struct {
	t *Target
	o *Operation
}

func New(opts ...any) (Action, error) {
	if len(opts) == 0 {
		return Action{}, nil
	}

	if len(opts) >= 2 {
		return Action{}, fmt.Errorf("%w: too many arguments in constructor call", ErrInvalidAction)
	}

	action := Action{}
	for _, arg := range opts {
		if opt, ok := arg.(string); ok {
			target, e1 := NewTarget(opt)
			operation, e2 := NewOperation(opt)

			if e1 != nil && e2 != nil {
				return Action{}, fmt.Errorf("%w: unknown argument: %v", ErrInvalidAction, errors.Join(e1, e2))
			}

			if e1 == nil && action.t == nil {
				action.t = &target
			}

			if e2 == nil && action.o == nil {
				action.o = &operation
			}

			continue
		}

		if opt, ok := arg.(Target); ok {
			action.t = &opt
			continue
		}

		if opt, ok := arg.(Operation); ok {
			action.o = &opt
			continue
		}

		return Action{}, fmt.Errorf("%w: argument of type %T not accepted", ErrInvalidAction, arg)
	}

	return action, nil
}

func MustNew(opts ...any) Action {
	act, _ := New(opts...)
	return act
}

func (a Action) String() string {
	if a.t == nil && a.o == nil {
		return "*"
	}

	if a.o == nil {
		return fmt.Sprintf("%s:*", a.t)
	}

	if a.t == nil {
		return fmt.Sprintf("*:%s", a.o)
	}

	return fmt.Sprintf("%s:%s", a.t, a.o)
}
