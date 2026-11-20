package middleware

import (
	"github.com/gofiber/fiber/v2"
)

// WithTrustedProxies makes c.IP() return the visitor's address when the app
// runs behind reverse proxies (deploy/: Caddy, then the web container's
// nginx), so the limiter counts each visitor rather than the proxy.
//
// The address is the first valid entry of X-Forwarded-For, read only when the
// request comes from one of trustedProxies (CIDR ranges or addresses);
// otherwise, or when the header holds no valid address, the connection's
// address is used. Without trusted proxies the header is never read, so a
// client cannot choose its own address by sending it.
func WithTrustedProxies(cfg fiber.Config, trustedProxies []string) fiber.Config {
	if len(trustedProxies) == 0 {
		return cfg
	}

	cfg.ProxyHeader = fiber.HeaderXForwardedFor
	cfg.EnableTrustedProxyCheck = true
	cfg.TrustedProxies = trustedProxies
	cfg.EnableIPValidation = true

	return cfg
}
