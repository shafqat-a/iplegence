package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/shafqat-a/iplegence/internal/validate"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: validate <mmdb> [golden.json]")
		os.Exit(1)
	}
	dbPath := os.Args[1]
	golden := "testdata/golden.json"
	if len(os.Args) > 2 {
		golden = os.Args[2]
	}
	cases, err := validate.LoadCases(golden)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	errs := validate.Check(dbPath, cases)
	failed := false
	for _, c := range cases {
		var hit error
		prefix := c.IP + ":"
		for _, e := range errs {
			if e.Error() == c.IP+": not found" || strings.HasPrefix(e.Error(), prefix) {
				hit = e
				break
			}
		}
		if hit == nil {
			fmt.Printf("OK   %s\n", c.IP)
			continue
		}
		fmt.Printf("FAIL %s\n", hit)
		failed = true
	}
	if failed {
		os.Exit(1)
	}
}
