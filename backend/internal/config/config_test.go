package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_Validate(t *testing.T) {
	t.Run("all required set", func(t *testing.T) {
		c := &Config{
			AppName: "svc", DBHost: "h", DBUser: "u", DBPassword: "p", DBName: "d",
			InternalServiceSecret: "s", JWTSecret: "j",
		}
		assert.NoError(t, c.Validate())
	})

	t.Run("lists every missing variable", func(t *testing.T) {
		c := &Config{AppName: "svc", DBHost: "h", DBUser: "u", DBName: "d", JWTSecret: "  "}
		err := c.Validate()
		assert.EqualError(t, err, "[CONFIG] validation failed. missing variables: DB_PASSWORD, INTERNAL_SERVICE_SECRET, JWT_SECRET")
	})
}
