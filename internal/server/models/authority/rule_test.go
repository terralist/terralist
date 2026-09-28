package authority

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestAllowsUpstream(t *testing.T) {
	Convey("Subject: Deciding whether an upstream artifact version may be served", t, func() {
		Convey("Given an authority without an enabled upstream", func() {
			a := Authority{}

			Convey("Then nothing is allowed", func() {
				So(a.AllowsUpstream(RuleKindProvider, "aws", "5.0.0"), ShouldBeFalse)
			})
		})

		Convey("Given an enabled upstream with the allow policy", func() {
			a := Authority{UpstreamEnabled: true, UpstreamDefaultPolicy: PolicyAllow}

			Convey("Then everything is allowed without rules", func() {
				So(a.AllowsUpstream(RuleKindProvider, "aws", "5.0.0"), ShouldBeTrue)
			})

			Convey("When a deny rule matches the name and version", func() {
				a.Rules = []Rule{{Kind: RuleKindProvider, Name: "aws", Version: "5.*", Effect: EffectDeny}}

				Convey("Then the matching versions are denied and the others allowed", func() {
					So(a.AllowsUpstream(RuleKindProvider, "aws", "5.0.0"), ShouldBeFalse)
					So(a.AllowsUpstream(RuleKindProvider, "aws", "4.9.0"), ShouldBeTrue)
					So(a.AllowsUpstream(RuleKindProvider, "google", "5.0.0"), ShouldBeTrue)
				})
			})

			Convey("When a deny rule targets another kind", func() {
				a.Rules = []Rule{{Kind: RuleKindModule, Name: "*", Version: "*", Effect: EffectDeny}}

				Convey("Then providers are not affected", func() {
					So(a.AllowsUpstream(RuleKindProvider, "aws", "5.0.0"), ShouldBeTrue)
					So(a.AllowsUpstream(RuleKindModule, "vpc/aws", "1.0.0"), ShouldBeFalse)
				})
			})

			Convey("When allow and deny rules both match", func() {
				a.Rules = []Rule{
					{Kind: RuleKindProvider, Name: "aws", Version: "*", Effect: EffectAllow},
					{Kind: RuleKindProvider, Name: "*", Version: "5.*", Effect: EffectDeny},
				}

				Convey("Then deny wins", func() {
					So(a.AllowsUpstream(RuleKindProvider, "aws", "5.0.0"), ShouldBeFalse)
				})
			})
		})

		Convey("Given an enabled upstream with the deny policy", func() {
			a := Authority{UpstreamEnabled: true, UpstreamDefaultPolicy: PolicyDeny}

			Convey("Then nothing is allowed without rules", func() {
				So(a.AllowsUpstream(RuleKindProvider, "aws", "5.0.0"), ShouldBeFalse)
			})

			Convey("When an allow rule matches", func() {
				a.Rules = []Rule{{Kind: RuleKindProvider, Name: "aws", Version: "5.*", Effect: EffectAllow}}

				Convey("Then only the matching versions are allowed", func() {
					So(a.AllowsUpstream(RuleKindProvider, "aws", "5.0.0"), ShouldBeTrue)
					So(a.AllowsUpstream(RuleKindProvider, "aws", "4.9.0"), ShouldBeFalse)
					So(a.AllowsUpstream(RuleKindProvider, "google", "5.0.0"), ShouldBeFalse)
				})
			})

			Convey("When allow and deny rules both match", func() {
				a.Rules = []Rule{
					{Kind: RuleKindProvider, Name: "aws", Version: "*", Effect: EffectAllow},
					{Kind: RuleKindProvider, Name: "aws", Version: "5.0.0", Effect: EffectDeny},
				}

				Convey("Then deny wins", func() {
					So(a.AllowsUpstream(RuleKindProvider, "aws", "5.0.0"), ShouldBeFalse)
					So(a.AllowsUpstream(RuleKindProvider, "aws", "5.0.1"), ShouldBeTrue)
				})
			})
		})

		Convey("Given a module rule with a name and system glob", func() {
			a := Authority{UpstreamEnabled: true, UpstreamDefaultPolicy: PolicyDeny, Rules: []Rule{
				{Kind: RuleKindModule, Name: "*/aws", Version: "*", Effect: EffectAllow},
			}}

			Convey("Then modules for that system are allowed", func() {
				So(a.AllowsUpstream(RuleKindModule, "vpc/aws", "1.0.0"), ShouldBeTrue)
				So(a.AllowsUpstream(RuleKindModule, "vpc/google", "1.0.0"), ShouldBeFalse)
			})
		})

		Convey("Given a rule whose name differs in case from the artifact", func() {
			a := Authority{UpstreamEnabled: true, UpstreamDefaultPolicy: PolicyAllow, Rules: []Rule{
				{Kind: RuleKindProvider, Name: "AWS", Version: "5.0.0-RC1", Effect: EffectDeny},
			}}

			Convey("Then the name matches regardless of case, while the version stays exact", func() {
				So(a.AllowsUpstream(RuleKindProvider, "aws", "5.0.0-RC1"), ShouldBeFalse)
				So(a.AllowsUpstream(RuleKindProvider, "aws", "5.0.0-rc1"), ShouldBeTrue)
			})
		})
	})
}

func TestRuleValidate(t *testing.T) {
	Convey("Subject: Validating a rule", t, func() {
		valid := Rule{Kind: RuleKindProvider, Name: "*", Version: "*", Effect: EffectAllow}

		Convey("A complete rule is valid", func() {
			So(valid.Validate(), ShouldBeNil)
		})

		Convey("An unknown kind is rejected", func() {
			r := valid
			r.Kind = "bucket"
			So(r.Validate(), ShouldNotBeNil)
		})

		Convey("An unknown effect is rejected", func() {
			r := valid
			r.Effect = "maybe"
			So(r.Validate(), ShouldNotBeNil)
		})

		Convey("An empty name glob is rejected", func() {
			r := valid
			r.Name = ""
			So(r.Validate(), ShouldNotBeNil)
		})

		Convey("An empty version glob is rejected", func() {
			r := valid
			r.Version = ""
			So(r.Validate(), ShouldNotBeNil)
		})

		Convey("An invalid glob is rejected", func() {
			r := valid
			r.Name = "[unclosed"
			So(r.Validate(), ShouldNotBeNil)
		})
	})
}
