package main

import (
	"fmt"
	"log"
	"net/http"
)

type Config struct {
}

const webPort = "80"

func main() {
	app := Config{}

	log.Printf("Starting server on port %s", webPort)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	err := srv.ListenAndServe()
	if err != nil {
		log.Panic(err)
	}
}
