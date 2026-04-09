package auth

import "github.com/gin-gonic/gin"

// Context is the authenticated subject attached to a request (after token middleware).
type Context struct {
	UserID    string
	DomainID  string
	ProjectID string
	// ScopeDomainID is set for Keystone domain-scoped tokens (empty for project/unscoped).
	ScopeDomainID string
	Roles         []string
}

// GinKey is the Gin context key for Context.
const GinKey = "gostone_auth"

// FromGin returns auth context if middleware ran successfully.
func FromGin(c *gin.Context) (Context, bool) {
	v, ok := c.Get(GinKey)
	if !ok {
		return Context{}, false
	}
	ctx, ok := v.(Context)
	return ctx, ok
}

// HasRole reports whether the token carries the given role name.
func (c Context) HasRole(name string) bool {
	for _, r := range c.Roles {
		if r == name {
			return true
		}
	}
	return false
}
