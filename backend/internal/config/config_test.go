package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplitList(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  []string
	}{
		{"empty", "", nil},
		{"blank", "  ", nil},
		{"one", "10.0.0.0/8", []string{"10.0.0.0/8"}},
		{
			"several with spaces and empty items",
			" 10.0.0.0/8, 172.16.0.0/12 ,,192.168.0.0/16 ",
			[]string{"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, splitList(tt.value))
		})
	}
}

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
