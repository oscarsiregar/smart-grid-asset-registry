package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_Validate(t *testing.T) {
	t.Run("all required set", func(t *testing.T) {
		c := &Config{
			AppName: "svc", DBHost: "h", DBUser: "u", DBPassword: "p", DBName: "d",
		}
		assert.NoError(t, c.Validate())
	})

	t.Run("lists every missing variable", func(t *testing.T) {
		c := &Config{AppName: "  ", DBHost: "h", DBUser: "u", DBName: "d"}
		err := c.Validate()
		assert.EqualError(t, err, "[CONFIG] validation failed. missing variables: APP_NAME, DB_PASSWORD")
	})
}
