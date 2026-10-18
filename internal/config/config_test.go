package config_test

import (
	"os"
	"testing"

	"github.com/Albert-Ti/go-keeper/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuild(t *testing.T) {
	type want struct {
		runAddr string
		db      string
	}

	allEnvKeys := []string{
		"SERVER_ADDRESS",
		"DB_CONN_STRING",
		"ENABLE_SMTP",
	}

	tests := []struct {
		name string
		args []string
		env  map[string]string
		want want
	}{
		{
			name: "env only",
			args: []string{"test"},
			env: map[string]string{
				"SERVER_ADDRESS": "localhost:8888",
				"DB_CONN_STRING": "postgres://env",
			},
			want: want{
				runAddr: "localhost:8888",
				db:      "postgres://env",
			},
		},
		{
			name: "flag only",
			args: []string{"test", "-a=localhost:9090", "-d=postgres://flag"},
			want: want{
				runAddr: "localhost:9090",
				db:      "postgres://flag",
			},
		},
		{
			name: "flag has priority over env",
			args: []string{"test", "-a=localhost:9090", "-d=postgres://flag"},
			env: map[string]string{
				"SERVER_ADDRESS": "localhost:8888",
				"DB_CONN_STRING": "postgres://env",
			},
			want: want{
				runAddr: "localhost:9090",
				db:      "postgres://flag",
			},
		},
		{
			name: "defaults, nothing set",
			args: []string{"test"},
			want: want{
				runAddr: "localhost:8080",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, key := range allEnvKeys {
				t.Setenv(key, "")
			}
			for k, v := range tt.env {
				t.Setenv(k, v)
			}

			origArgs := os.Args
			defer func() { os.Args = origArgs }()
			os.Args = tt.args

			cfg, err := config.Build()
			require.NoError(t, err)

			assert.Equal(t, tt.want.runAddr, cfg.RunAddr)
			assert.Equal(t, tt.want.db, cfg.DBConnStr)
		})
	}
}
