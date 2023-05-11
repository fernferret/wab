package main

import (
	"fmt"
	"os"

	"github.com/fernferret/wab"
	"github.com/fernferret/wab/internal/util"
	"go.uber.org/zap"

	flag "github.com/spf13/pflag"
)

func usage() {
	fmt.Fprintf(os.Stderr, "WAB Version: %s\n\nusage: %s\n", version, os.Args[0])
	flag.PrintDefaults()
}

var (
	versionFull = "dev"
	version     = "dev"
	versionMode = "dev"
)

func main() {
	// HTTP Server options
	options := &wab.Options{}
	// Default log level
	logLevelString := flag.String("level", "info", "log level, can be one of: trace, debug, info, warn, error, fatal, panic")

	flag.IntVarP(&options.Port, "port", "p", 8080, "set the port for the HTTP server")
	flag.StringVarP(&options.Host, "bind", "b", "127.0.0.1", "set the bind host for the HTTP server")
	flag.BoolVar(&options.DevMode, "dev", false, "if true, CORS headers will be insecure, use if you're splitting the API/Server for now.")
	flag.BoolVar(&options.LogRequests, "log-requests", false, "if true, http requests will be logged, pretty loud")
	printVersion := flag.Bool("version", false, "print the version and exit")
	flag.Usage = usage
	flag.Parse()

	if *printVersion {
		fmt.Printf("WAB Version: %v\n", versionFull)
		os.Exit(0)
	}

	// Setup the zap logger
	util.SetupLogger(*logLevelString)
	log := zap.S()

	// Print some info about the server
	hostPretty := options.Host
	log.Infof("Starting HTTP server: http://%s:%d", hostPretty, options.Port)
	server := wab.NewAPIServer(options)
	server.RunLoop()
}
