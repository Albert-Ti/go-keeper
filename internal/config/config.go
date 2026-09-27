// При Написание configs использовался pattern Builder(Строитель/функц. опции)
package config

import (
	"encoding/json"
	"flag"
	"log/slog"
	"os"
)

type Mode string

const (
	ModeDev   Mode = "dev"
	ModeDebug Mode = "debug"
	ModeProd  Mode = "prod"
)

// Options хранит настройки приложения, собранные из флагов(высокий приоритет) командной строки,
// переменных окружения(средний приоритет), файла конфигурации(низкий приоритет)
// и значений по умолчанию.
// generate:reset
type Options struct {
	RunAddr   string
	DBConnStr string
	JWTSecret string
	Mode      Mode

	EnableSMTP bool
	SMTPOpt    *smtpFileOptions

	CacheClientRunAddr string
	CacheClientPass    string

	ObjStorageRunAddr   string
	ObjStorageAccessKey string
	ObjStorageSecretKey string
}

type smtpFileOptions struct {
	Port     int    `json:"port"`
	Host     string `json:"host"`
	Username string `json:"username"`
	Pass     string `json:"pass"`
}

// NewOptions создаёт Options со значениями по умолчанию и применяет
func NewOptions(opts ...func(*Options)) *Options {
	o := &Options{
		RunAddr:             "localhost:8080",
		JWTSecret:           "jwt_secret_key",
		Mode:                ModeDev,
		CacheClientRunAddr:  "localhost:6379",
		CacheClientPass:     "redis",
		ObjStorageRunAddr:   "http://localhost:8333",
		ObjStorageAccessKey: "s3-access",
		ObjStorageSecretKey: "s3-secret",
	}

	for _, opt := range opts {
		opt(o)
	}
	return o
}

// Build собирает Options из флагов командной строки, переменных окружения
// и файла конфигурации. Приоритет: явный флаг > env > файл конфига > дефолт.
func Build() (*Options, error) {
	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	var raw Options
	fs.StringVar(&raw.RunAddr, "a", "localhost:8080", "адрес и порт запуска HTTP-сервера, например: -a=localhost:8080")
	fs.StringVar(&raw.DBConnStr, "d", "", "строка подключения к БД, например: -d=\"postgres://user:pass@localhost:5432/shortener\"")
	fs.BoolVar(&raw.EnableSMTP, "e", false, "")

	if err := fs.Parse(os.Args[1:]); err != nil {
		return nil, err
	}

	explicit := map[string]bool{}
	fs.Visit(func(f *flag.Flag) {
		explicit[f.Name] = true
	})

	opts := NewOptions()

	opts.RunAddr = pickString(explicit["a"], "SERVER_ADDRESS", raw.RunAddr)
	opts.DBConnStr = pickString(explicit["d"], "DB_CONN_STRING", raw.DBConnStr)
	opts.EnableSMTP = pickBool(explicit["e"], "ENABLE_SMTP", raw.EnableSMTP)

	var smtpOpts smtpFileOptions
	if opts.EnableSMTP {
		parsed, err := parseFileOptions("smtp_config.json")
		if err != nil {
			slog.Error("parse config file", "path", "smtp_config.json", "error", err)
		} else {
			smtpOpts = *parsed
		}
		opts.SMTPOpt = &smtpOpts
	}

	return opts, nil
}

// WithRunAddr задаёт адрес и порт, на которых запускается сервер HTTP.
func WithRunAddr(v string) func(*Options) { return func(o *Options) { o.RunAddr = v } }

// WithDBConnStr задаёт строку подключения к Postgres.
func WithDBConnStr(v string) func(*Options) { return func(o *Options) { o.DBConnStr = v } }

// WithJWTSecret задаёт секретный ключ для подписи JWT-токенов.
func WithJWTSecret(v string) func(*Options) { return func(o *Options) { o.JWTSecret = v } }

// WithMode задаёт режим работы приложения (например, "dev" или "debug").
func WithMode(v Mode) func(*Options) { return func(o *Options) { o.Mode = v } }

// WithEnableSMTP.
func WithEnableSMTP(v bool) func(*Options) { return func(o *Options) { o.EnableSMTP = v } }

func pickString(explicitFlag bool, envStr string, flagVal string) string {
	if explicitFlag {
		return flagVal
	}
	if v := os.Getenv(envStr); v != "" {
		return v
	}

	return flagVal
}

func pickBool(explicitFlag bool, envStr string, flagVal bool) bool {
	if explicitFlag {
		return flagVal
	}
	if v := os.Getenv(envStr); v != "" {
		return v == "true" || v == "1"
	}

	return flagVal
}

func parseFileOptions(fname string) (*smtpFileOptions, error) {
	file, err := os.Open(fname)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var fc smtpFileOptions
	if err := json.NewDecoder(file).Decode(&fc); err != nil {
		return nil, err
	}
	return &fc, nil
}
