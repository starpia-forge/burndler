package main

import (
	"log"
	"os"

	"github.com/burndler/burndler/internal/app"
)

// Build-time variables injected via ldflags
var (
	Version   = "dev"     // Version is set during build
	BuildTime = "unknown" // BuildTime is set during build
	GitCommit = "unknown" // GitCommit is set during build
)

func main() {
	// Create CLI with build information
	cli := app.NewCLI(app.BuildInfo{
		Version:   Version,
		BuildTime: BuildTime,
		GitCommit: GitCommit,
	})

	// Run CLI with command-line arguments
	if err := cli.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
