package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	redis "gopkg.in/redis.v3"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Host string
	Port int

	RedisKey string

	RedisADDR     string
	RedisPassword string
	RedisDB       int64 `envconfig`

	LivenessStatus int
	ReadyStatus    int
}

const APP_NAME = "visit"

var version = "changeme"

type App struct {
	redis  *redis.Client
	config *Config

	hostname string

	*http.ServeMux
}

func NewApp(redis *redis.Client, config *Config, hostname string) *App {
	app := &App{
		redis:    redis,
		config:   config,
		hostname: hostname,
		ServeMux: http.NewServeMux(),
	}

	app.HandleFunc("/", app.Visit)
	app.HandleFunc("/health", app.Health)
	app.HandleFunc("/ready", app.Ready)

	return app
}

func (a *App) Visit(w http.ResponseWriter, r *http.Request) {
	cmd := a.redis.Incr(a.config.RedisKey)
	if cmd.Err() != nil {
		w.WriteHeader(500)
		fmt.Fprintf(w, "Error contacting redis: %s\n", cmd.Err())
		return
	}

	count, err := cmd.Result()
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintf(w, "Error contacting redis: %s\n", cmd.Err())
		return
	}

	fmt.Fprintf(w, "The current visit count is %d on %s running version %s.\n", count, a.hostname, version)
}

func (a *App) Health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(a.config.LivenessStatus)
}

func (a *App) Ready(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(a.config.ReadyStatus)
}

func main() {
	var config Config
	err := envconfig.Process(APP_NAME, &config)
	if err != nil {
		log.Fatalf("error reading config: %s", err)
	}

	if config.LivenessStatus == 0 {
		config.LivenessStatus = 200
	}

	if config.ReadyStatus == 0 {
		config.ReadyStatus = 200
	}

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown-host"
	}

	log.Printf("starting version %s with config: %+v", version, config)

	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("error listening on %q: %s", addr, err)
	}
	defer listener.Close()

	log.Printf("listening on %q", listener.Addr())

	client := redis.NewClient(&redis.Options{
		Addr:     config.RedisADDR,
		Password: config.RedisPassword,
		DB:       config.RedisDB,
	})

	app := NewApp(client, &config, hostname)

	if err := http.Serve(listener, app); err != nil {
		log.Fatalf("error serving http: %s", err)
	}
}
