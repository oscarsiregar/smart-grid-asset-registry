package product

import (
	"time"

	"backend-template/domain/dto"
	"backend-template/domain/entity"

	"github.com/gofiber/fiber/v2"
	pkg_middleware_fiber "repo.smf-indonesia.co.id/package-helper/go-common/pkg/middleware/fiber"
)

// caller returns the authenticated user's name and permissions placed in
// ctx.Locals by the Identity()/JWTVerifyStrict() middleware chain. An
// unidentified caller yields ("", []string{}) — the usecase then denies.
func caller(ctx *fiber.Ctx) (name string, permissions []string) {
	permissions = []string{}
	identity := pkg_middleware_fiber.GetUserIdentity(ctx)
	if identity == nil {
		return "", permissions
	}
	if identity.Permissions != nil {
		permissions = identity.Permissions
	}
	return identity.Name, permissions
}

func toProductResponse(p *entity.Product) dto.ProductResponse {
	return dto.ProductResponse{
		ID:            p.ID.String(),
		Code:          p.Code,
		Name:          p.Name,
		Description:   p.Description,
		Price:         p.Price,
		Status:        p.Status.String(),
		CreatedAt:     p.CreatedAt.Format(time.RFC3339),
		CreatedByName: p.CreatedByName,
		UpdatedAt:     p.UpdatedAt.Format(time.RFC3339),
		UpdatedByName: p.UpdatedByName,
	}
}
