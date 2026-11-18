package fiber

import gofiber "github.com/gofiber/fiber/v2"

// TEMPORARY: this file exists only because the template's `product` example
// handler still reads a caller identity. The product has no users
// (Template Exception 1); T108 deletes this file together with `product`.

// UserIdentityKey is the ctx.Locals key holding the caller's identity.
const UserIdentityKey = "user_identity"

// UserIdentity is the caller identity go-common's Identity middleware stores.
type UserIdentity struct {
	ID          string   `json:"id"`
	Email       string   `json:"email"`
	Name        string   `json:"name"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
}

// GetUserIdentity returns the identity stored under UserIdentityKey, or nil.
func GetUserIdentity(c *gofiber.Ctx) *UserIdentity {
	identity, ok := c.Locals(UserIdentityKey).(*UserIdentity)
	if !ok {
		return nil
	}
	return identity
}
