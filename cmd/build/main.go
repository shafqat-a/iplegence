package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"

	"github.com/shafqat-a/iplegence/internal/pipeline"
)

func main() {
	configPath := flag.String("config", "configs/sources.yaml", "sources config")
	priorityPath := flag.String("priority", "configs/priority.yaml", "field priority config")
	skip := flag.Bool("skip-download", false, "use already-downloaded files")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if err := pipeline.Run(ctx, pipeline.Options{
		ConfigPath:   *configPath,
		PriorityPath: *priorityPath,
		SkipDownload: *skip,
	}); err != nil {
		log.Fatal(err)
	}
}
