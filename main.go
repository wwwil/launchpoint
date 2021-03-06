package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/wwwil/launchpoint/pkg/gpio"
	"github.com/wwwil/launchpoint/pkg/launchpoint"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// Version is the version of the app. This is injected during build.
var Version = "development"

// Commit is the commit hash of the build. This is injected during build.
var Commit string

// BuildDate is the date of the build. This is injected during build.
var BuildDate string

// GoVersion is the Go version used for the build. This is injected during build.
var GoVersion string

// Platform is the target platform for this build. This is injected during build.
var Platform string

// Variables set by argument flags.
var configFilePath = flag.String("config", "./launchpoint.yaml", "Path to the configuration file")

func main() {
	// Don't show time and date on log messages.
	log.SetFlags(0)

	// Print build information.
	versionString := "Launchpoint"
	versionString = fmt.Sprintf("%s\n  Version: %s %s", versionString, Version, Platform)
	versionString = fmt.Sprintf("%s\n  Commit: %s", versionString, Commit)
	versionString = fmt.Sprintf("%s\n  Built: %s", versionString, BuildDate)
	versionString = fmt.Sprintf("%s\n  Go: %s", versionString, GoVersion)
	log.Println(versionString)

	// Parse command line flags.
	flag.Parse()

	config, err := launchpoint.LoadConfigFromFile(*configFilePath)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("using configuration file: %s\n", *configFilePath)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var wg sync.WaitGroup

	// Start listeners.
	wg.Add(1)
	go gpio.Run(ctx, &wg, config)

	// Capture exit signals to ensure resources are released on exit.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(quit)

	// Run until told to quit.
	<-quit
	log.Println("Launchpoint will now quit.")
	// Cancel the context to stop the other threads.
	cancel()
	// Wait for other threads to finish.
	wg.Wait()
}
