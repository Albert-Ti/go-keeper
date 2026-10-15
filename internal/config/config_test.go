package config_test

import (
	"os"
	"testing"

	"github.com/Albert-Ti/go-keeper/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var configStr = `{
  "server_address": "localhost:8080",
  "database_db": "",
}`

func CreateConfigFileTmp() error {

	err := os.WriteFile("testCfg.json", []byte(configStr), 0666)
	if err != nil {
		return err
	}

	return nil
}

func TestBuild(t *testing.T) {
	defer os.Remove("testCfg.json")

	type want struct {
		runAddr string
		db      string
	}

	allEnvKeys := []string{
		"SERVER_ADDRESS",
		"BASE_URL",
		"DATABASE_CONN_STRING",
		"JWT_SECRET_KEY",
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
				"BASE_URL":       "http://localhost:8888",
			},
			want: want{
				runAddr: "localhost:8888",
			},
		},
		{
			name: "flag only",
			args: []string{"test", "-a=localhost:9090", "-b=http://localhost:9090"},
			want: want{
				runAddr: "localhost:9090",
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

			if tt.name == "Test file config" {
				err := CreateConfigFileTmp()
				require.NoError(t, err)
			}

			cfg, err := config.Build()
			require.NoError(t, err)

			if tt.want.runAddr != "" {
				assert.Equal(t, tt.want.runAddr, cfg.RunAddr)
			}
			assert.Equal(t, tt.want.db, cfg.DBConnStr)
		})
	}
}
