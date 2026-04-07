package v3

import "github.com/gin-gonic/gin"

func registerV3ProjectsAndUsers(v3 *gin.RouterGroup, h *Hub) {
	registerAny(v3, "/projects/:project_id/tags", "/v3/projects/:project_id/tags", h)
	registerAny(v3, "/projects/:project_id/tags/:value", "/v3/projects/:project_id/tags/:value", h)
	registerAny(v3, "/projects/:project_id/users/:user_id/roles", "/v3/projects/:project_id/users/:user_id/roles", h)
	registerAny(v3, "/projects/:project_id/users/:user_id/roles/:role_id", "/v3/projects/:project_id/users/:user_id/roles/:role_id", h)
	registerAny(v3, "/projects/:project_id/groups/:group_id/roles", "/v3/projects/:project_id/groups/:group_id/roles", h)
	registerAny(v3, "/projects/:project_id/groups/:group_id/roles/:role_id", "/v3/projects/:project_id/groups/:group_id/roles/:role_id", h)

	registerAny(v3, "/users/:user_id/password", "/v3/users/:user_id/password", h)
	registerAny(v3, "/users/:user_id/groups", "/v3/users/:user_id/groups", h)
	registerAny(v3, "/users/:user_id/projects", "/v3/users/:user_id/projects", h)
	registerAny(v3, "/users/:user_id/credentials/OS-EC2", "/v3/users/:user_id/credentials/OS-EC2", h)
	registerAny(v3, "/users/:user_id/credentials/OS-EC2/:credential_id", "/v3/users/:user_id/credentials/OS-EC2/:credential_id", h)

	registerAny(v3, "/users/:user_id/OS-OAUTH1/access_tokens", "/v3/users/:user_id/OS-OAUTH1/access_tokens", h)
	registerAny(v3, "/users/:user_id/OS-OAUTH1/access_tokens/:access_token_id", "/v3/users/:user_id/OS-OAUTH1/access_tokens/:access_token_id", h)
	registerAny(v3, "/users/:user_id/OS-OAUTH1/access_tokens/:access_token_id/roles", "/v3/users/:user_id/OS-OAUTH1/access_tokens/:access_token_id/roles", h)
	registerAny(v3, "/users/:user_id/OS-OAUTH1/access_tokens/:access_token_id/roles/:role_id", "/v3/users/:user_id/OS-OAUTH1/access_tokens/:access_token_id/roles/:role_id", h)

	registerAny(v3, "/users/:user_id/application_credentials", "/v3/users/:user_id/application_credentials", h)
	registerAny(v3, "/users/:user_id/application_credentials/:application_credential_id", "/v3/users/:user_id/application_credentials/:application_credential_id", h)
	registerAny(v3, "/users/:user_id/access_rules", "/v3/users/:user_id/access_rules", h)
	registerAny(v3, "/users/:user_id/access_rules/:access_rule_id", "/v3/users/:user_id/access_rules/:access_rule_id", h)
}
