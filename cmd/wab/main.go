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
)

func main() {
	// HTTP Server options
	options := &wab.Options{}
	// Default log level
	logLevelString := flag.String("level", "info", "log level, can be one of: trace, debug, info, warn, error, fatal, panic")

	flag.StringVarP(&options.Bind, "bind", "b", "127.0.0.1:8080", "set the bind host for the http server")
	flag.StringVarP(&options.BindGRPC, "bind-grpc", "g", "127.0.0.1:5050", "set the bind address for the gRPC server")
	flag.BoolVar(&options.DisableGRPC, "no-grpc", false, "disable the native gRPC binding, grpcweb will still be available")
	flag.BoolVar(&options.DisableReflection, "no-reflection", false, "disable gRPC reflection, this will prevent gRPCurl from working")
	flag.BoolVar(&options.DevMode, "dev", false, "if true, CORS headers will be insecure, use if you're splitting the API/Server for now.")
	flag.BoolVar(&options.LogRequests, "log-requests", false, "if true, http requests will be logged, pretty loud")
	flag.BoolVar(&options.DisableGRPCUI, "no-grpcui", false, "disable the GRPCUI debug endpoint at /grpc-ui/")
	flag.BoolVar(&options.DisableGRPCWeb, "no-grpcweb", false, "disable the grpcweb endpoint at /grpc/, this means the embedded Vue app won't work")
	printVersion := flag.Bool("version", false, "print the version and exit")
	flag.Usage = usage
	flag.CommandLine.SortFlags = false
	flag.Parse()

	if *printVersion {
		fmt.Printf("WAB Version: %v\n", versionFull)
		os.Exit(0)
	}

	// Setup the zap logger by marking it as zap's global logger and setting some
	// pretty options.
	util.SetupLogger(*logLevelString)

	log := zap.S()

	// Print some info about the server
	log.Infof("Starting HTTP server: http://%s", options.Bind)
	server := wab.NewAPIServer(options)
	server.RunLoop()
}
