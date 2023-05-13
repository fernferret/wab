package wab

//go:generate go run -tags=dev webui_generate.go

import (
	"fmt"
	"io/fs"
	"net/http"
	"os"

	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	glog "github.com/labstack/gommon/log"
	"go.uber.org/zap"

	"github.com/fernferret/wab/internal/wabmw"
	"github.com/fernferret/wab/ui"
)

// Options contains all the information about the web api. These are options like host, port, dev mode, etc.
type Options struct {
	Port          int
	Host          string
	DevMode       bool // If true, CORS headers will be not good.
	LogRequests   bool
	BuildMode     string
	DisableGRPCUI bool
}

// WebServer holds the internal fields for the HTTP Server (Echo) as well as the HTTP Client
// used for performing outbound requests.
type WebServer struct {
	e       *echo.Echo
	options *Options
	client  *http.Client
	log     *zap.SugaredLogger
	// object  *api.Object
}

// NewAPIServer creates a new APIServer which is the primary web
// application/API.
func NewAPIServer(options *Options) *WebServer {
	server := &WebServer{
		e:       echo.New(),
		options: options,
		log:     zap.S(),
		client: &http.Client{
			Timeout: time.Duration(1) * time.Minute,
		},
	}

	return server
}

// RunLoop is the primary run method for the APIServer, call it when you're
// ready to start the application.
//
// NOTE: this method is blocking.
func (s *WebServer) RunLoop() {
	// Start a GRPCServer and setup the webui for debugging.
	//
	// TODO: Disable grpcweb if we disable the embedded UI
	grpcWebHdlr, debugUIHdlr := ServeGRPC(5050, true, !s.options.DisableGRPCUI)

	s.setupHTTPServer(grpcWebHdlr, debugUIHdlr)
}

// setupHTTPServer creates a new Echo HTTP server. If an http.Handler is passed
// in, it is assumed to be a grpcui based handler for debugging GRPC user
// interfaces with GRPCURL.
//
// The GRPCUI is embedded into the process and is totally standalone.
func (s *WebServer) setupHTTPServer(grpcWebHdlr, debugHandler http.Handler) {
	s.e = echo.New()

	if s.options.DevMode {
		zap.S().Warn("DevMode enabled; your server is not secure against CORS based attacks.")
		s.e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins: []string{"*"},
			AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
		}))
	}

	// Setup the quiet logger. The default Echo Logger spews junk with it's own formatting.
	s.e.HTTPErrorHandler = s.quietHTTPErrorHandler

	if s.options.LogRequests {
		// Setup the pretty zap logger :)
		s.e.Use(wabmw.ZapMiddleware())
	}

	// Make sure to catch anything bad that we might do to avoid crashes.
	s.e.Use(wabmw.CustomRecoverWithConfig(middleware.RecoverConfig{
		DisableStackAll: false,
	}))

	s.e.Logger.SetLevel(glog.DEBUG)

	const grpcPath = "/grpc"

	var grpcwebHandler echo.HandlerFunc

	if grpcWebHdlr != nil {
		s.log.Infof("Setup grpcweb at %s", grpcPath)
		grpcwebHandler = echo.WrapHandler(http.StripPrefix(grpcPath, debugHandler))
	} else {
		grpcwebHandler = noGRPCWebHandler
	}

	s.e.Any(fmt.Sprintf("%s/*", grpcPath), grpcwebHandler)

	// Optionally setup the GRPCUI embedded debug server. This is (intentionally)
	// totally independent of the web-app below.
	//
	// If a function was not passed in we'll embed a helpful message to users that
	// they have everything setup correctly, but we did not embed the UI. In
	// production you should probably not compile this in (via a different
	// buildtag) but here in WAB I'm just doing it at runtime.
	const grpcUIPath = "/grpc-ui"

	var grpcUIDebugHandler echo.HandlerFunc

	if debugHandler != nil {
		s.log.Infof("Setup GRPC UI at %s", grpcUIPath)
		grpcUIDebugHandler = echo.WrapHandler(http.StripPrefix(grpcUIPath, debugHandler))
	} else {
		grpcUIDebugHandler = noGRPCUIHandler
	}

	s.e.Any(fmt.Sprintf("%s/*", grpcUIPath), grpcUIDebugHandler)

	// Setup the handler that will serve the embedded VueJS application.
	s.setupStaticHandler()

	s.setupRoutes()
	s.e.HideBanner = true
	s.e.HidePort = true

	if err := s.e.Start(fmt.Sprintf("%s:%d", s.options.Host, s.options.Port)); err != nil {
		s.e.Logger.Info(fmt.Sprintf("shutting down the server: %s", err))
		os.Exit(1)
	}
}

// quietHTTPErrorHandler is identical to the built-in error handler, but I tailored it
func (s *WebServer) quietHTTPErrorHandler(err error, ectx echo.Context) {
	var (
		code = http.StatusInternalServerError
		msg  interface{}
	)

	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
		msg = he.Message
	} else if s.e.Debug {
		msg = err.Error()
	} else {
		msg = http.StatusText(code)
	}

	if _, ok := msg.(string); ok {
		msg = echo.Map{"message": msg}
	}

	if !ectx.Response().Committed {
		if ectx.Request().Method == echo.HEAD { // Issue #608
			if err := ectx.NoContent(code); err != nil {
				goto ERROR
			}
		} else {
			if err := ectx.JSON(code, msg); err != nil {
				goto ERROR
			}
		}
	}
ERROR:
	// I perform all my logging in the request/response area. I don't want this extra print
	// e.Logger.Error(err)
	// log.WithFields(log.Fields{"code": code, "err": err}).Errorf("Unable to respond with error code")

	return
}

func (s *WebServer) setupStaticHandler() {
	assets := ui.GetAssets()
	if assets == nil {
		s.e.GET("/*", noEmbeddedUIHandler)
	} else {
		// Create a handler for the static files
		// This is done so we can all *ALL* other routes
		// to go to index.html. This lets Vue.js do history-based
		// routing that looks great and works great.
		staticHandler := http.FileServer(http.FS(assets))
		s.e.GET("/static/*", echo.WrapHandler(staticHandler))
		s.e.GET("/assets/*", echo.WrapHandler(staticHandler))

		// Load the index.html file. This is the entrypoint for the
		// entire user-interface and sub-routing actually happens
		// in here! For example if a user sees:
		// http://localhost:1323/about in the browser, this is still
		// serving up index.html but then vue.js is loading the about
		// sub-page.
		indexFile, err := fs.ReadFile(assets, "index.html")
		if err != nil {
			s.log.Fatalw("Failed to load critical index.html file, cannot continue", "err", err)
		}
		s.e.GET("/*", func(ectx echo.Context) error {
			err := ectx.Blob(http.StatusOK, "text/html", indexFile)
			if err != nil {
				return fmt.Errorf("failed to load blob %q %w", ectx.Request().RequestURI, err)
			}

			return nil
		})
	}
}

func (s *WebServer) getState() echo.HandlerFunc {
	return func(ctx echo.Context) error {
		result := map[string]interface{}{
			"all_on":  true,
			"dogs_on": false,
			"cats_on": true,
		}

		return ctx.JSON(http.StatusOK, result)
	}
}

func (s *WebServer) setupRoutes() {
	s.e.GET("/api/v1/state", s.getState())
}

// Base Handlers
func noEmbeddedUIHandler(ectx echo.Context) error {
	return ectx.String(http.StatusOK, "No UI embedded in this copy of wab")
}

func noGRPCUIHandler(ectx echo.Context) error {
	return ectx.String(http.StatusOK, "GRPCUI is not enabled in this copy of wab")
}

func noGRPCWebHandler(ectx echo.Context) error {
	return ectx.String(http.StatusOK, "The grpcweb extension is not enabled in this copy of wab")
}
