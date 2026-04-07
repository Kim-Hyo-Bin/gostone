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
	rules := map[string]string{
		"identity:get_user":    "role:admin or user_match",
		"identity:list_users":  "role:admin",
		"identity:create_user": "role:admin",
		"identity:update_user": "role:admin or user_match",
		"identity:delete_user": "role:admin",

		"identity:change_user_password":      "role:admin or user_match",
		"identity:list_groups_for_user":      "role:admin or user_match",
		"identity:list_projects_for_user":    "role:admin or user_match",
		"identity:list_credentials_for_user": "role:admin or user_match",
		"identity:create_ec2_credential":     "role:admin or user_match",
		"identity:get_ec2_credential":        "role:admin or user_match",
		"identity:delete_ec2_credential":     "role:admin or user_match",

		"identity:list_auth_projects": "authenticated",
		"identity:list_auth_domains":  "authenticated",
		"identity:get_simple_ca":      "authenticated",

		"identity:list_application_credentials":  "role:admin or user_match",
		"identity:create_application_credential": "role:admin or user_match",
		"identity:get_application_credential":    "role:admin or user_match",
		"identity:delete_application_credential": "role:admin or user_match",
		"identity:list_access_rules":             "role:admin or user_match",
		"identity:create_access_rule":            "role:admin or user_match",
		"identity:get_access_rule":               "role:admin or user_match",
		"identity:delete_access_rule":            "role:admin or user_match",
		"identity:list_oauth1_access_tokens":     "role:admin or user_match",
		"identity:create_oauth1_access_token":    "role:admin or user_match",
		"identity:get_oauth1_access_token":       "role:admin or user_match",
		"identity:delete_oauth1_access_token":    "role:admin or user_match",

		"identity:list_domains":           "role:admin",
		"identity:get_domain":             "role:admin or domain_match",
		"identity:create_domain":          "role:admin",
		"identity:update_domain":          "role:admin",
		"identity:delete_domain":          "role:admin",
		"identity:list_projects":          "role:admin",
		"identity:get_project":            "role:admin or project_match",
		"identity:create_project":         "role:admin",
		"identity:update_project":         "role:admin",
		"identity:delete_project":         "role:admin",
		"identity:list_roles":             "role:admin",
		"identity:get_role":               "role:admin or authenticated",
		"identity:create_role":            "role:admin",
		"identity:update_role":            "role:admin",
		"identity:delete_role":            "role:admin",
		"identity:list_role_assignments":  "role:admin",
		"identity:create_role_assignment": "role:admin",

		"identity:list_regions":  "role:admin",
		"identity:get_region":    "role:admin",
		"identity:create_region": "role:admin",
		"identity:update_region": "role:admin",
		"identity:delete_region": "role:admin",

		"identity:list_services":  "role:admin",
		"identity:get_service":    "role:admin",
		"identity:create_service": "role:admin",
		"identity:update_service": "role:admin",
		"identity:delete_service": "role:admin",

		"identity:list_endpoints":  "role:admin",
		"identity:get_endpoint":    "role:admin",
		"identity:create_endpoint": "role:admin",
		"identity:update_endpoint": "role:admin",
		"identity:delete_endpoint": "role:admin",

		"identity:list_project_user_roles":  "role:admin",
		"identity:assign_project_user_role": "role:admin",
		"identity:remove_project_user_role": "role:admin",
	}
	for _, k := range extraAdminOnlyRules() {
		if _, ok := rules[k]; !ok {
			rules[k] = "role:admin"
		}
	}
	return &Policy{Rules: rules, DefaultRule: "authenticated"}
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
