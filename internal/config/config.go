// При Написание configs использовался pattern Builder(Строитель/функц. опции)
package config

import (
	"flag"
	"os"
)

const (
	ModeDev   = "dev"
	ModeDebug = "debug"
	ModeProd  = "prod"
)

// Options хранит настройки приложения, собранные из флагов(высокий приоритет) командной строки,
// переменных окружения(средний приоритет), файла конфигурации(низкий приоритет)
// и значений по умолчанию.
// generate:reset
type Options struct {
	RunAddr   string
	DBConnStr string
	JWTSecret string
	Mode      string
}

// NewOptions создаёт Options со значениями по умолчанию и применяет
// переданные опции (pattern Builder / функциональные опции).
func NewOptions(opts ...func(*Options)) *Options {
	o := &Options{
		RunAddr:   "localhost:8080",
		JWTSecret: "jwt_secret_key",
		Mode:      ModeDev,
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

	if err := fs.Parse(os.Args[1:]); err != nil {
		return nil, err
	}

	explicit := map[string]bool{}
	fs.Visit(func(f *flag.Flag) {
		explicit[f.Name] = true
	})

	opts := NewOptions()

	return opts, nil
}

// WithRunAddr задаёт адрес и порт, на которых запускается сервер HTTP.
func WithRunAddr(v string) func(*Options) { return func(o *Options) { o.RunAddr = v } }

// WithDBConnStr задаёт строку подключения к Postgres.
func WithDBConnStr(v string) func(*Options) { return func(o *Options) { o.DBConnStr = v } }

// WithJWTSecret задаёт секретный ключ для подписи JWT-токенов.
func WithJWTSecret(v string) func(*Options) { return func(o *Options) { o.JWTSecret = v } }

// WithMode задаёт режим работы приложения (например, "dev" или "debug").
func WithMode(v string) func(*Options) { return func(o *Options) { o.Mode = v } }
