package rbac

import (
	_ "embed"
	"errors"
	"fmt"
	"slices"
	"terralist/pkg/auth"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
	fileadapter "github.com/casbin/casbin/v2/persist/file-adapter"
	"github.com/casbin/govaluate"
	stringadapter "github.com/qiangmzsx/string-adapter/v2"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

var (
	ErrUnsupported         = errors.New("element not supported")
	ErrUnauthorizedSubject = errors.New("subject not authorized")
)

const (
	ResourceModules     = "modules"
	ResourceProviders   = "providers"
	ResourceAuthorities = "authorities"

	ActionGet    = "get"
	ActionUpdate = "update"
	ActionCreate = "create"
	ActionDelete = "delete"
)

var (
	Resources []string = []string{
		ResourceModules,
		ResourceProviders,
		ResourceAuthorities,
	}

	Actions []string = []string{
		ActionGet,
		ActionUpdate,
		ActionCreate,
		ActionDelete,
	}
)

//go:embed model.conf
var defaultModel string

// Make sure that CasbinEnforcer interface properly wraps the casbin.Enforcer struct.
var _ CasbinEnforcer = &casbin.Enforcer{}

// CasbinEnforcer defines the methods we use from casbin.Enforcer, allowing for easier testing/mocking.
type CasbinEnforcer interface {
	AddFunction(name string, function govaluate.ExpressionFunction)
	GetRolesForUser(name string, domain ...string) ([]string, error)
	BatchEnforce(rvals [][]any) ([]bool, error)
}

// Enforcer is a wrapper around casbin.Enforcer that supports default roles and glob matching.
type Enforcer struct {
	enforcer    CasbinEnforcer
	defaultRole string
}

// NewEnforcer creates a new authorization manager.
func NewEnforcer(policyPath string, defaultRoleName string) (*Enforcer, error) {
	m, err := model.NewModelFromString(defaultModel)
	if err != nil {
		return nil, err
	}

	var a persist.Adapter
	if policyPath == "" {
		a = stringadapter.NewAdapter("# empty policy")
	} else {
		a = fileadapter.NewAdapter(policyPath)
	}

	enforcer, err := casbin.NewEnforcer(m, a)
	if err != nil {
		return nil, err
	}

	enforcer.AddFunction("glob_match", globMatch)

	defaultRole := "role:readonly"
	if defaultRoleName != "" {
		defaultRole = fmt.Sprintf("role:%v", defaultRoleName)
	}

	return &Enforcer{
		enforcer:    enforcer,
		defaultRole: defaultRole,
	}, nil
}

// enforce checks if the subject is allowed to perform the action on the resource and object.
func (e *Enforcer) enforce(subjects []string, resource, object, action string) bool {
	logger := log.With().
		Strs("subjects", subjects).
		Str("resource", resource).
		Str("action", action).
		Logger()

	var roles []string

	for _, subject := range subjects {
		userRoles, err := e.enforcer.GetRolesForUser(subject)

		if err != nil {
			logger.Warn().Str("subject", subject).Err(err).Msg("Failed to get roles for user")
			continue
		}

		roles = append(roles, userRoles...)
	}

	// If the user has no roles, assign the default role.
	if len(roles) == 0 {
		roles = []string{e.defaultRole}
	}

	if slices.Contains(roles, "role:admin") {
		logger.Debug().Msg("Administrator role detected, allowing any action.")
		return true
	}

	if action == ActionGet && slices.Contains(roles, "role:readonly") {
		logger.Debug().Msg("Read-only role detected, allowing 'get' action.")
		return true
	}

	requests := lo.Map(subjects, func(subject string, _ int) []any {
		return []any{subject, resource, action, object}
	})

	results, err := e.enforcer.BatchEnforce(requests)
	if err != nil {
		return false
	}

	return lo.SomeBy(results, func(r bool) bool { return r })
}

// Protect checks if the user is authorized to perform the action on the resource and object,
// considering their origin. It returns an error if the user is not authorized.
func (e *Enforcer) Protect(subject auth.User, resource, action, object string) error {
	if !slices.Contains(Resources, resource) {
		return fmt.Errorf("%w: resource %v", ErrUnsupported, resource)
	}

	if !slices.Contains(Actions, action) {
		return fmt.Errorf("%w: action %v", ErrUnsupported, action)
	}

	subjects := lo.Uniq(
		append(
			[]string{subject.Name, subject.Email},
			lo.Map(subject.Groups, func(group string, _ int) string {
				return fmt.Sprintf("role:%s", group)
			})...,
		),
	)

	if ok := e.enforce(subjects, resource, object, action); !ok {
		log.Debug().
			Str("user", subject.String()).
			Str("resource", resource).
			Str("action", action).
			Str("object", object).
			Msg("User not authorized")

		return ErrUnauthorizedSubject
	}

	return nil
}
