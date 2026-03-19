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
	ResourceApiKeys     = "api-keys"

	ActionGet    = "get"
	ActionUpdate = "update"
	ActionCreate = "create"
	ActionDelete = "delete"

	EffectAllow = "allow"
	EffectDeny  = "deny"

	SubjectAnonymous = "role:anonymous"
	SubjectReadonly  = "role:readonly"
	SubjectAdmin     = "role:admin"
)

var (
	Resources []string = []string{
		ResourceModules,
		ResourceProviders,
		ResourceAuthorities,
		ResourceApiKeys,
	}

	Actions []string = []string{
		ActionGet,
		ActionUpdate,
		ActionCreate,
		ActionDelete,
	}

	Effects []string = []string{
		EffectAllow,
		EffectDeny,
	}
)

//go:embed model.conf
var defaultModel string

// defaultPolicies are baked-in policies that provide sensible defaults for built-in roles.
// User-provided policies are loaded on top of these.
var defaultPolicies = [][]string{
	{SubjectAdmin, "*", "*", "*", EffectAllow},
	{SubjectReadonly, ResourceModules, ActionGet, "*", EffectAllow},
	{SubjectReadonly, ResourceProviders, ActionGet, "*", EffectAllow},
	{SubjectReadonly, ResourceAuthorities, ActionGet, "*", EffectAllow},
}

// Make sure that CasbinEnforcer interface properly wraps the casbin.Enforcer struct.
var _ CasbinEnforcer = &casbin.Enforcer{}

// CasbinEnforcer defines the methods we use from casbin.Enforcer, allowing for easier testing/mocking.
type CasbinEnforcer interface {
	AddFunction(name string, function govaluate.ExpressionFunction)
	AddPolicy(params ...any) (bool, error)
	GetRolesForUser(name string, domain ...string) ([]string, error)
	BatchEnforce(rvals [][]any) ([]bool, error)
}

// Enforcer is a wrapper around casbin.Enforcer that supports default roles and glob matching.
type Enforcer struct {
	enforcer    CasbinEnforcer
	defaultRole string
}

// NewEnforcer creates a new authorization manager with a file-based policy.
func NewEnforcer(policyPath string, defaultRoleName string) (*Enforcer, error) {
	var a persist.Adapter
	if policyPath == "" {
		a = stringadapter.NewAdapter("# empty policy")
	} else {
		a = fileadapter.NewAdapter(policyPath)
	}

	return newEnforcer(a, defaultRoleName)
}

// NewEnforcerFromString creates a new authorization manager with a string-based policy.
func NewEnforcerFromString(policyCSV string, defaultRoleName string) (*Enforcer, error) {
	return newEnforcer(stringadapter.NewAdapter(policyCSV), defaultRoleName)
}

func newEnforcer(adapter persist.Adapter, defaultRoleName string) (*Enforcer, error) {
	m, err := model.NewModelFromString(defaultModel)
	if err != nil {
		return nil, err
	}

	enforcer, err := casbin.NewEnforcer(m, adapter)
	if err != nil {
		return nil, err
	}

	enforcer.AddFunction("glob_match", globMatch)

	for _, policy := range defaultPolicies {
		if _, err := enforcer.AddPolicy(lo.Map(policy, func(s string, _ int) any { return s })...); err != nil {
			return nil, fmt.Errorf("failed to add default policy %v: %w", policy, err)
		}
	}

	defaultRole := SubjectReadonly
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
		if subject == "" {
			continue
		}

		userRoles, err := e.enforcer.GetRolesForUser(subject)

		if err != nil {
			logger.Warn().Str("subject", subject).Err(err).Msg("Failed to get roles for user")
			continue
		}

		roles = append(roles, userRoles...)
	}

	// If the user is authenticated and has no roles, assign the default role.
	if !slices.Contains(subjects, SubjectAnonymous) && len(roles) == 0 {
		roles = []string{e.defaultRole}
	}

	// Evaluate all subjects and their roles against the policy.
	allSubjects := lo.Uniq(append(subjects, roles...))
	requests := lo.Map(allSubjects, func(subject string, _ int) []any {
		return []any{subject, resource, action, object}
	})

	results, err := e.enforcer.BatchEnforce(requests)
	if err != nil {
		return false
	}

	return lo.SomeBy(results, func(r bool) bool { return r })
}

// EvaluateInline evaluates a set of inline policies against a request.
// It follows the same semantics as the casbin model:
// allowed if some policy allows AND no policy denies.
func EvaluateInline(policies []auth.Policy, resource, action, object string) bool {
	hasAllow := false
	for _, p := range policies {
		resMatchVal, _ := globMatch(resource, p.Resource)
		actMatchVal, _ := globMatch(action, p.Action)
		objMatchVal, _ := globMatch(object, p.Object)

		resMatch, _ := resMatchVal.(bool)
		actMatch, _ := actMatchVal.(bool)
		objMatch, _ := objMatchVal.(bool)

		if resMatch && actMatch && objMatch {
			if p.Effect == EffectDeny {
				return false
			}
			if p.Effect == EffectAllow {
				hasAllow = true
			}
		}
	}
	return hasAllow
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

	// If the user has inline policies (standalone API key), evaluate them directly.
	if len(subject.InlinePolicies) > 0 {
		if !EvaluateInline(subject.InlinePolicies, resource, action, object) {
			log.Debug().
				Str("user", subject.String()).
				Str("resource", resource).
				Str("action", action).
				Str("object", object).
				Msg("User not authorized by inline policies")

			return ErrUnauthorizedSubject
		}

		return nil
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
