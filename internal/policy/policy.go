package policy

import (
	"strings"

	"github.com/Kim-Hyo-Bin/gostone/internal/auth"
)

// Policy is a minimal Keystone-style rule map (action → expression).
// Only a few rules are evaluated; unknown actions fall back to DefaultRule.
type Policy struct {
	Rules       map[string]string
	DefaultRule string
}

// Default builds the starter policy used by gostone (extend via config file later).
func Default() *Policy {
	return &Policy{
		Rules: map[string]string{
			"identity:get_user":    "role:admin or user_match",
			"identity:list_users":  "role:admin",
			"identity:create_user": "role:admin",

			"identity:list_domains":          "role:admin",
			"identity:get_domain":            "role:admin or domain_match",
			"identity:create_domain":         "role:admin",
			"identity:update_domain":         "role:admin",
			"identity:delete_domain":         "role:admin",
			"identity:list_projects":         "role:admin",
			"identity:get_project":           "role:admin or project_match",
			"identity:create_project":        "role:admin",
			"identity:update_project":        "role:admin",
			"identity:delete_project":        "role:admin",
			"identity:list_roles":            "role:admin",
			"identity:get_role":              "role:admin or authenticated",
			"identity:create_role":           "role:admin",
			"identity:update_role":           "role:admin",
			"identity:delete_role":           "role:admin",
			"identity:list_role_assignments": "role:admin",
		},
		DefaultRule: "authenticated",
	}
}

// Allow evaluates the rule for action using token context and optional targets (e.g. user_id).
func (p *Policy) Allow(action string, ctx auth.Context, target map[string]string) bool {
	rule, ok := p.Rules[action]
	if !ok {
		rule = p.DefaultRule
	}
	return p.evalOr(rule, ctx, target)
}

func (p *Policy) evalOr(rule string, ctx auth.Context, target map[string]string) bool {
	if target == nil {
		target = map[string]string{}
	}
	for _, part := range strings.Split(rule, " or ") {
		part = strings.TrimSpace(part)
		if p.evalAtom(part, ctx, target) {
			return true
		}
	}
	return false
}

func (p *Policy) evalAtom(part string, ctx auth.Context, target map[string]string) bool {
	switch part {
	case "authenticated":
		return ctx.UserID != ""
	case "role:admin":
		return ctx.HasRole("admin")
	case "user_match":
		return target["user_id"] == ctx.UserID
	case "domain_match":
		return target["domain_id"] != "" && target["domain_id"] == ctx.DomainID
	case "project_match":
		return target["project_id"] != "" && target["project_id"] == ctx.ProjectID
	default:
		return false
	}
}
