package main

import (
	"fmt"
	"log"
	"net"
	"net/http"

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
}

const APP_NAME = "visit"

type App struct {
	redis  *redis.Client
	config *Config
}

func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

	fmt.Fprintf(w, "The current visit count is %d.\n", count)
}

func main() {
	var config Config
	err := envconfig.Process(APP_NAME, &config)
	if err != nil {
		log.Fatalf("error reading config: %s", err)
	}

	log.Printf("starting with config: %+v", config)

	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	log.Printf("listening on %q", addr)

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("error listening on %q: %s", addr, err)
	}
	defer listener.Close()

	client := redis.NewClient(&redis.Options{
		Addr:     config.RedisADDR,
		Password: config.RedisPassword,
		DB:       config.RedisDB,
	})

	app := App{
		redis:  client,
		config: &config,
	}

	if err := http.Serve(listener, &app); err != nil {
		log.Fatalf("error serving http: %s", err)
	}
}
