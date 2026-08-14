package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/shafqat-a/iplegence/internal/lookup"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: lookup <ip> [mmdb]")
		os.Exit(1)
	}
	ip := os.Args[1]
	path := "dist/Superior-IP.mmdb"
	if len(os.Args) > 2 {
		path = os.Args[2]
	}
	db, err := lookup.Open(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer db.Close()
	rec, ok, err := db.Lookup(ip)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if !ok {
		_ = enc.Encode(map[string]string{"error": "not found"})
		os.Exit(2)
	}
	if err := enc.Encode(rec); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
