package policy

// extraAdminOnlyRules are Keystone-style actions defaulting to admin-only (unless overridden in baseRules).
func extraAdminOnlyRules() []string {
	return []string{
		"identity:list_groups", "identity:get_group", "identity:create_group", "identity:update_group", "identity:delete_group",
		"identity:list_group_users", "identity:add_user_to_group", "identity:remove_user_from_group",

		"identity:list_policies", "identity:get_policy", "identity:create_policy", "identity:update_policy", "identity:delete_policy",
		"identity:list_policy_endpoints", "identity:add_policy_endpoint", "identity:remove_policy_endpoint",
		"identity:list_policy_services", "identity:add_policy_service", "identity:remove_policy_service_region",

		"identity:list_registered_limits", "identity:get_registered_limit", "identity:create_registered_limit", "identity:update_registered_limit", "identity:delete_registered_limit",
		"identity:list_limits", "identity:get_limit", "identity:create_limit", "identity:update_limit", "identity:delete_limit",
		"identity:get_limits_model",

		"identity:list_credentials", "identity:get_credential", "identity:create_credential", "identity:update_credential", "identity:delete_credential",

		"identity:list_domain_user_roles", "identity:assign_domain_user_role", "identity:remove_domain_user_role",
		"identity:list_domain_group_roles", "identity:assign_domain_group_role", "identity:remove_domain_group_role",

		"identity:list_group_project_roles", "identity:assign_group_project_role", "identity:remove_group_project_role",

		"identity:list_project_tags", "identity:add_project_tag", "identity:remove_project_tag",

		"identity:list_trusts", "identity:get_trust", "identity:create_trust", "identity:delete_trust",
		"identity:list_trust_roles", "identity:add_trust_role", "identity:remove_trust_role",

		"identity:list_identity_providers", "identity:get_identity_provider", "identity:create_identity_provider", "identity:update_identity_provider", "identity:delete_identity_provider",
		"identity:list_mappings", "identity:get_mapping", "identity:create_mapping", "identity:update_mapping", "identity:delete_mapping",
		"identity:list_service_providers", "identity:get_service_provider", "identity:create_service_provider", "identity:update_service_provider", "identity:delete_service_provider",

		"identity:list_oauth_consumers", "identity:get_oauth_consumer", "identity:create_oauth_consumer", "identity:update_oauth_consumer", "identity:delete_oauth_consumer",

		"identity:list_revoke_events", "identity:create_revoke_event",

		"identity:list_oauth1_token_roles", "identity:assign_oauth1_token_role", "identity:remove_oauth1_token_role",

		"identity:list_system_user_roles", "identity:assign_system_user_role", "identity:remove_system_user_role",
		"identity:list_system_group_roles", "identity:assign_system_group_role", "identity:remove_system_group_role",

		"identity:list_inherited_roles", "identity:get_domain_config", "identity:set_domain_config", "identity:delete_domain_config",

		"identity:list_endpoint_groups", "identity:list_project_endpoint_filters", "identity:add_project_endpoint_filter", "identity:remove_project_endpoint_filter",
		"identity:list_endpoint_group_endpoints", "identity:add_endpoint_group_endpoint", "identity:list_endpoint_group_projects", "identity:add_endpoint_group_project",

		"identity:get_auth_catalog", "identity:get_auth_system",
	}
}
