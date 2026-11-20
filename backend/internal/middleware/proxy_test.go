package middleware

import (
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// app.Test connects from 0.0.0.0, so "0.0.0.0/32" plays the reverse proxy.
const testPeer = "0.0.0.0"

func newClientIPApp(trustedProxies []string) *fiber.App {
	app := fiber.New(WithTrustedProxies(fiber.Config{}, trustedProxies))
	app.Get("/ip", func(c *fiber.Ctx) error {
		return c.SendString(c.IP())
	})

	return app
}

func TestWithTrustedProxies_ClientIP(t *testing.T) {
	tests := []struct {
		name           string
		trustedProxies []string
		forwardedFor   string
		want           string
	}{
		{
			name:           "trusted proxy: first forwarded address is the client",
			trustedProxies: []string{testPeer + "/32"},
			forwardedFor:   "203.0.113.7, 172.18.0.5",
			want:           "203.0.113.7",
		},
		{
			name:           "trusted proxy without the header: the peer",
			trustedProxies: []string{testPeer + "/32"},
			want:           testPeer,
		},
		{
			name:           "trusted proxy with an invalid header: the peer",
			trustedProxies: []string{testPeer + "/32"},
			forwardedFor:   "not-an-ip",
			want:           testPeer,
		},
		{
			name:           "untrusted peer: the header is ignored",
			trustedProxies: []string{"10.0.0.0/8"},
			forwardedFor:   "203.0.113.7",
			want:           testPeer,
		},
		{
			name:         "no trusted proxies: the header is ignored",
			forwardedFor: "203.0.113.7",
			want:         testPeer,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/ip", nil)
			if tt.forwardedFor != "" {
				req.Header.Set(fiber.HeaderXForwardedFor, tt.forwardedFor)
			}

			resp, err := newClientIPApp(tt.trustedProxies).Test(req)
			require.NoError(t, err)

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			assert.Equal(t, tt.want, string(body))
		})
	}
}

func TestWithTrustedProxies_KeepsOtherSettings(t *testing.T) {
	got := WithTrustedProxies(fiber.Config{AppName: "svc", BodyLimit: 42}, []string{"10.0.0.0/8"})

	assert.Equal(t, "svc", got.AppName)
	assert.Equal(t, 42, got.BodyLimit)
}

// Behind the proxy every request comes from the same peer; the limiter must
// still count each visitor on their own.
func TestLimiterConfig_BehindTrustedProxyCountsEachClient(t *testing.T) {
	app := fiber.New(WithTrustedProxies(fiber.Config{}, []string{testPeer + "/32"}))
	app.Use(LimiterConfig())
	app.Get("/x", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusBadRequest).SendString("nope")
	})

	request := func(client string) int {
		req := httptest.NewRequest("GET", "/x", nil)
		req.Header.Set(fiber.HeaderXForwardedFor, client)
		resp, err := app.Test(req)
		require.NoError(t, err)

		return resp.StatusCode
	}

	for i := 0; i < 20; i++ {
		assert.Equal(t, fiber.StatusBadRequest, request("203.0.113.7"))
	}

	assert.Equal(t, fiber.StatusTooManyRequests, request("203.0.113.7"), "the noisy client is limited")
	assert.Equal(t, fiber.StatusBadRequest, request("198.51.100.9"), "another client is not")
}
