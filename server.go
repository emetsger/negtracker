package main

import (
	"context"
	"fmt"
	"github.com/emetsger/negtracker/handler/neg"
	"github.com/emetsger/negtracker/store"
	"github.com/emetsger/negtracker/store/mongo"
	"github.com/emetsger/negtracker/urlutil/strip"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"
)

const (
	STARTING = iota
	RUNNING
	STOPPING
	STOPPED
)

var s *http.Server
var config *Configuration
var state State

type Configuration struct {
	Host   string
	Port   string
	Secure bool
}

type State int

var mongoConfig = &mongo.MongoConfig{
	DbUri:         getEnvOrDefault(store.EnvDbUri, "mongodb://localhost:27017"),
	DbName:        getEnvOrDefault(store.EnvDbName, "negtracker"),
	NegCollection: getEnvOrDefault(store.EnvDbNegCollection, "neg"),
	Opts:          options.Client().SetAppName("negtracker").SetServerSelectionTimeout(5 * time.Second),
}

var mongoStore = &mongo.MongoStore{}

func main() {
	state = STARTING
	pong := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte("Pong!"))
	}

	mongoStore.Configure(mongoConfig)

	http.HandleFunc("/Ping", pong)
	negHandler := neg.NewHandler(mongoStore)
	http.HandleFunc("/neg", negHandler)
	http.HandleFunc("/neg/", negHandler)

	s = &http.Server{}
	config = configure(s)
	start(s, config)
}

func configure(s *http.Server) *Configuration {
	c := &Configuration{
		Host: getEnvOrDefault("LISTEN_HOST", "localhost"),
		Port: getEnvOrDefault("LISTEN_PORT", "0"),
	}
	s.Addr = fmt.Sprintf("%s:%s", c.Host, c.Port)
	return c
}

func start(s *http.Server, c *Configuration) {
	// net.Listen interprets ":0" as pick a random open port
	l, err := net.Listen("tcp", fmt.Sprintf("%s:%s", c.Host, c.Port))

	if err != nil {
		panic(err)
	} else {
		c.Port = strconv.Itoa(l.Addr().(*net.TCPAddr).Port)
	}

	log.Printf("Starting HTTP server on port %s", c.Port)
	state = RUNNING
	if err := s.Serve(l); err != http.ErrServerClosed {
		log.Fatal("HTTP server stopped unexpectedly", err)
	}
}

func stop(s *http.Server) {
	state = STOPPING
	log.Print("Stopping HTTP server")
	ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
	err := s.Shutdown(ctx)
	log.Printf("%s, %v", "HTTP server shutdown", err)
	state = STOPPED
}

func getEnvOrDefault(envVar, defaultValue string) string {
	if value, exists := os.LookupEnv(envVar); exists == false {
		return defaultValue
	} else {
		return value
	}
}

func (c *Configuration) ListenUrl() string {
	scheme := "http"
	if c.Secure {
		scheme = "https"
	}

	return strip.TrailingSlashes(fmt.Sprintf("%s://%s:%s/", scheme, c.Host, c.Port))
}
