package handlers

import "github.com/gin-gonic/gin"

func registerV3Inherit(v3 *gin.RouterGroup, h *Hub) {
	inh := v3.Group("/OS-INHERIT")

	registerAny(
		inh,
		"/domains/:domain_id/groups/:group_id/roles/:role_id/inherited_to_projects",
		"/v3/OS-INHERIT/domains/:domain_id/groups/:group_id/roles/:role_id/inherited_to_projects",
		h,
	)
	registerAny(
		inh,
		"/domains/:domain_id/groups/:group_id/roles/inherited_to_projects",
		"/v3/OS-INHERIT/domains/:domain_id/groups/:group_id/roles/inherited_to_projects",
		h,
	)
	registerAny(
		inh,
		"/domains/:domain_id/users/:user_id/roles/:role_id/inherited_to_projects",
		"/v3/OS-INHERIT/domains/:domain_id/users/:user_id/roles/:role_id/inherited_to_projects",
		h,
	)
	registerAny(
		inh,
		"/domains/:domain_id/users/:user_id/roles/inherited_to_projects",
		"/v3/OS-INHERIT/domains/:domain_id/users/:user_id/roles/inherited_to_projects",
		h,
	)
	registerAny(
		inh,
		"/projects/:project_id/users/:user_id/roles/:role_id/inherited_to_projects",
		"/v3/OS-INHERIT/projects/:project_id/users/:user_id/roles/:role_id/inherited_to_projects",
		h,
	)
	registerAny(
		inh,
		"/projects/:project_id/groups/:group_id/roles/:role_id/inherited_to_projects",
		"/v3/OS-INHERIT/projects/:project_id/groups/:group_id/roles/:role_id/inherited_to_projects",
		h,
	)
}
