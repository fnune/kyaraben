package main

import (
	"flag"
	"log"
	"os"

	"github.com/fnune/kyaraben/relay/internal/server"
)

func main() {
	addr := flag.String("addr", ":8080", "listen address")
	flag.Parse()

	if port := os.Getenv("PORT"); port != "" {
		*addr = ":" + port
	}

	cfg := server.DefaultConfig()
	cfg.Addr = *addr

	srv := server.New(cfg)
	if err := srv.Start(); err != nil {
		log.Fatal(err)
	}
}
