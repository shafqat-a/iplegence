package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"

	"github.com/shafqat-a/iplegence/internal/serve"
)

func main() {
	addr := flag.String("addr", envOr("LISTEN_ADDR", ":8080"), "listen address")
	mmdb := flag.String("mmdb", envOr("MMDB_PATH", "dist/Superior-IP.mmdb"), "path to Superior-IP.mmdb")
	flag.Parse()

	srv, err := serve.New(*addr, *mmdb)
	if err != nil {
		log.Fatal(err)
	}
	defer srv.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	log.Printf("iplegence listening on %s (mmdb=%s)", *addr, *mmdb)
	if err := srv.ListenAndServe(ctx); err != nil {
		log.Fatal(err)
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
