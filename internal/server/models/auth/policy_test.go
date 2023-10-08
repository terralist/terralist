package auth

import (
	"encoding/json"
	"fmt"
	"terralist/internal/server/models/auth/action"
	"terralist/pkg/types"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/yaml.v3"
)

func TestParseJSON(t *testing.T) {
	Convey("Subject: Parse a JSON string and extract a policy", t, func() {
		Convey("Given a valid policy in JSON format", func() {
			text := `{
				"label": "Test",
				"effect": "Allow",
				"actions": [
  				"modules:*",
					"providers:get",
					"providers:create",
					"authorities:*"
				],
				"resources": [
					"module:*",
					"authority:test",
					"provider:aws"
				]
			}`

			Convey("When the policy is unmarshalled", func() {
				policy := &Policy{}
				err := json.Unmarshal([]byte(text), policy)

				Convey("The policy should be parsed correctly", func() {
					So(err, ShouldBeNil)
					So(policy.Label, ShouldEqual, "Test")
					So(policy.Effect, ShouldEqual, "Allow")
					So(policy.Actions.Contains("modules:*"), ShouldBeTrue)
					So(policy.Actions.Contains("providers:get"), ShouldBeTrue)
					So(policy.Actions.Contains("providers:create"), ShouldBeTrue)
					So(policy.Actions.Contains("authorities:*"), ShouldBeTrue)
					So(policy.Resources.Contains("module:*"), ShouldBeTrue)
					So(policy.Resources.Contains("authority:test"), ShouldBeTrue)
					So(policy.Resources.Contains("provider:aws"), ShouldBeTrue)
					So(policy.Resources.Contains("provider:random"), ShouldBeFalse)
				})
			})
		})

		Convey("Given an invalid policy in JSON format", func() {
			text := `{
				"label": "Test",
				"effect": "Allow",
				"actions": "modules:*",
				"resources": "module:*"
			}`
			Convey("When the policy is unmarshalled", func() {
				policy := &Policy{}
				err := yaml.Unmarshal([]byte(text), policy)

				Convey("The parser should return an error", func() {
					So(err, ShouldNotBeNil)
				})
			})
		})
	})
}

func TestParseYAML(t *testing.T) {
	Convey("Subject: Parse a YAML string and extract a policy", t, func() {
		Convey("Given a valid policy in YAML format", func() {
			text := `label: Test
effect: Allow
actions:
  - modules:*
  - providers:get
  - providers:create
  - authorities:*
resources:
  - module:*
  - authority:test
  - provider:aws`

			Convey("When the policy is unmarshalled", func() {
				policy := &Policy{}
				err := yaml.Unmarshal([]byte(text), policy)

				Convey("The policy should be parsed correctly", func() {
					So(err, ShouldBeNil)
					So(policy.Label, ShouldEqual, "Test")
					So(policy.Effect, ShouldEqual, "Allow")
					So(policy.Actions.Contains("modules:*"), ShouldBeTrue)
					So(policy.Actions.Contains("providers:get"), ShouldBeTrue)
					So(policy.Actions.Contains("providers:create"), ShouldBeTrue)
					So(policy.Actions.Contains("authorities:*"), ShouldBeTrue)
					So(policy.Resources.Contains("module:*"), ShouldBeTrue)
					So(policy.Resources.Contains("authority:test"), ShouldBeTrue)
					So(policy.Resources.Contains("provider:aws"), ShouldBeTrue)
					So(policy.Resources.Contains("provider:random"), ShouldBeFalse)
				})
			})
		})

		Convey("Given an invalid policy in YAML format", func() {
			text := `label: Test
effect: Allow
actions: modules:*
resources: module:*
`
			Convey("When the policy is unmarshalled", func() {
				policy := &Policy{}
				err := yaml.Unmarshal([]byte(text), policy)

				Convey("The parser should return an error", func() {
					So(err, ShouldNotBeNil)
				})
			})
		})
	})
}

/*
Case 1: Allow, Action, Resource => Granted
Case 2: Allow, Action, !Resource => Uncertain
Case 3: Allow, Action, Resource* => Granted

Case 4: Allow, !Action, Resource => Uncertain
Case 5: Allow, !Action, !Resource => Uncertain
Case 6: Allow, !Action, Resource* => Uncertain

Case 7: Allow, Action*, Resource => Granted
Case 8: Allow, Action*, !Resource => Uncertain
Case 9: Allow, Action*, Resource* => Granted

Case 1: Deny, Action, Resource => Denied
Case 2: Deny, Action, !Resource => Uncertain
Case 3: Deny, Action, Resource* => Denied

Case 4: Deny, !Action, Resource => Uncertain
Case 5: Deny, !Action, !Resource => Uncertain
Case 6: Deny, !Action, Resource* => Uncertain

Case 7: Deny, Action*, Resource => Denied
Case 8: Deny, Action*, !Resource => Uncertain
Case 9: Deny, Action*, Resource* => Denied
*/
func TestPolicyEvaluate(t *testing.T) {
	Convey("Subject: Validating actions for resources using policies", t, func() {
		Convey("Given a policy, an action and a resource", func() {
			testData := []struct {
				policy   Policy
				action   action.Action
				resource Resource
				expected Permission
			}{
				{
					policy: Policy{
						Effect: EffectAllow,
						Actions: *types.NewStringArray([]string{
							action.MustNew(action.Modules, action.All).String(),
						}),
						Resources: *types.NewStringArray([]string{
							string(ComposeResource(ResourceModule, "*")),
						}),
					},
					action:   action.MustNew(action.Modules.String(), action.View.String()),
					resource: ComposeResource(ResourceModule, "some-module"),
					expected: PermissionGranted,
				},
				{
					policy: Policy{
						Effect: EffectAllow,
						Actions: *types.NewStringArray([]string{
							string(ComposeAction(ActionModules, CategoryAny)),
						}),
						Resources: *types.NewStringArray([]string{
							string(ComposeResource(ResourceModule, "some-other-module")),
						}),
					},
					action:   ComposeAction(ActionModules, CategoryRead),
					resource: ComposeResource(ResourceModule, "some-module"),
					expected: PermissionUncertain,
				},
				{
					policy: Policy{
						Effect: EffectDeny,
						Actions: *types.NewStringArray([]string{
							string(ComposeAction(ActionModules, CategoryAny)),
						}),
						Resources: *types.NewStringArray([]string{
							string(ComposeResource(ResourceModule, "some-module")),
						}),
					},
					action:   ComposeAction(ActionModules, CategoryRead),
					resource: ComposeResource(ResourceModule, "some-module"),
					expected: PermissionDenied,
				},
				{
					policy: Policy{
						Effect: EffectDeny,
						Actions: *types.NewStringArray([]string{
							string(ComposeAction(ActionModules, CategoryAny)),
						}),
						Resources: *types.NewStringArray([]string{
							string(ComposeResource(ResourceModule, "*")),
						}),
					},
					action:   ComposeAction(ActionModules, CategoryRead),
					resource: ComposeResource(ResourceModule, "some-module"),
					expected: PermissionDenied,
				},
				{
					policy: Policy{
						Effect: EffectDeny,
						Actions: *types.NewStringArray([]string{
							string(ComposeAction(ActionModules, CategoryAny)),
						}),
						Resources: *types.NewStringArray([]string{
							string(ComposeResource(ResourceModule, "some-other-module")),
						}),
					},
					action:   ComposeAction(ActionModules, CategoryRead),
					resource: ComposeResource(ResourceModule, "some-other-module"),
					expected: PermissionUncertain,
				},
			}

			for i, t := range testData {
				Convey(fmt.Sprintf("When the policy is queried (#%d)", i), func() {
					result := t.policy.Evaluate(t.action, t.resource)

					Convey("The policy should respond accordingly", func() {
						So(result, ShouldEqual, t.expected)
					})
				})
			}
		})
	})
}
