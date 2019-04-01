package wab

//go:generate go run -tags=dev webui_generate.go

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"strings"

	"time"

	"github.com/aybabtme/rgbterm"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	glog "github.com/labstack/gommon/log"
	log "github.com/sirupsen/logrus"
)

const (
	padding int = 6
)

// Options contains all the information about the web api. These are options like host, port, dev mode, etc.
type Options struct {
	Port        int
	Host        string
	DevMode     bool // If true, CORS headers will be not good.
	LogRequests bool
	BuildMode   string
}

// APIServer holds the internal fields for the HTTP Server (Echo) as well as the HTTP Client
// used for performing outbound requests.
type APIServer struct {
	e       *echo.Echo
	options *Options
	client  *http.Client
	// object  *api.Object
}

// NewAPIServer creates a new strip server with a given API/Options
func NewAPIServer(options *Options) *APIServer {
	server := &APIServer{
		e:       echo.New(),
		options: options,
		client: &http.Client{
			Timeout: time.Duration(1) * time.Minute,
		},
	}
	return server
}

// RunLoop is the primary run method for the StripServer
// NOTE: this method is blocking.
func (s *APIServer) RunLoop() {
	s.setupHTTPServer()
}

func red(text string) string {
	return rgbterm.FgString(text, 255, 0, 0)
}

func green(text string) string {
	return rgbterm.FgString(text, 0, 255, 0)
}

// func darkGreen(text string) string {
// 	return rgbterm.FgString(text, 69, 139, 0)
// }
//
// func darkCyan(text string) string {
// 	return rgbterm.FgString(text, 0, 139, 139)
// }

func cyan(text string) string {
	return rgbterm.FgString(text, 0, 255, 255)
}

func orange(text string) string {
	return rgbterm.FgString(text, 255, 165, 0)
}

func logrusMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			methodColor := req.Method
			switch req.Method {
			case "GET":
				methodColor = green(req.Method)
			case "POST":
				methodColor = cyan(req.Method)
			case "PUT":
				methodColor = orange(req.Method)
			case "DELETE":
				methodColor = red(req.Method)
			}
			fmtString := fmt.Sprintf("%s %%-%ds %%s", "-->", len(methodColor)-len(req.Method)+padding)
			log.Infof(fmtString, methodColor, req.URL.Path)

			start := time.Now()
			err := next(c)
			stop := time.Now()

			var statusCode int
			if err != nil {
				statusCode = http.StatusInternalServerError
				if he, ok := err.(*echo.HTTPError); ok {
					statusCode = he.Code
				}
			} else {
				statusCode = c.Response().Status
			}

			statusColor := func(text string) string { return text }
			logFcn := log.Infof
			if statusCode >= 500 {
				statusColor = red
			} else if statusCode >= 400 {
				statusColor = orange
			} else if statusCode >= 300 {
				statusColor = cyan
			} else if statusCode >= 200 {
				statusColor = green
			}
			coloredStatus := statusColor(fmt.Sprintf("%d", statusCode))
			fmtString = fmt.Sprintf("<-- %%-%ds %%s (took: %%s)", len(coloredStatus)-len(fmt.Sprintf("%d", statusCode))+padding)
			latency := stop.Sub(start).String()
			logFcn(fmtString, coloredStatus, http.StatusText(statusCode), latency)

			return err
		}
	}
}

func (s *APIServer) setupHTTPServer() {
	s.e = echo.New()

	if s.options.DevMode {
		log.Warn("DevMode enabled; your server is not secure against CORS based attacks.")
		s.e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins: []string{"*"},
			AllowMethods: []string{echo.GET, echo.PUT, echo.POST, echo.DELETE},
		}))
		//upgrader.CheckOrigin = func(r *http.Request) bool {
		//	return true
		//}
	}

	// Setup the quiet logger. The default Echo Logger spews junk with it's own formatting.
	s.e.HTTPErrorHandler = s.quietHTTPErrorHandler

	if s.options.LogRequests {
		// Setup the pretty logrus logger :)
		s.e.Use(logrusMiddleware())
	}

	// Make sure to catch anything bad that we might do to avoid crashes.
	s.e.Use(CustomRecoverWithConfig(middleware.RecoverConfig{
		DisableStackAll: false,
	}))

	s.e.Logger.SetLevel(glog.DEBUG)

	s.setupStaticHandler()

	s.setupRoutes()
	s.e.HideBanner = false
	s.e.HidePort = false
	if err := s.e.Start(fmt.Sprintf("%s:%d", s.options.Host, s.options.Port)); err != nil {
		s.e.Logger.Info(fmt.Sprintf("shutting down the server: %s", err))
		os.Exit(1)
	}
}

// This is an example custom handler that i'm not using anymore.
// func (s *APIServer) customHTTPErrorHandler(err error, c echo.Context) {
// 	code := http.StatusInternalServerError
// 	if he, ok := err.(*echo.HTTPError); ok {
// 		code = he.Code
// 	}
// 	errorPage := fmt.Sprintf("%d.html", code)
// 	if err := c.File(errorPage); err != nil {
// 		// c.Logger().Error(err)
// 		log.WithFields(log.Fields{"code": code, "err": err}).Errorf("Unable to load error page...")
// 	}
// 	// c.Logger().Error(err)
// 	log.WithFields(log.Fields{"code": code}).Errorf("Pooperz")
// }

// quietHTTPErrorHandler is identical to the built-in error handler, but I tailored it
func (s *APIServer) quietHTTPErrorHandler(err error, c echo.Context) {
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

	if !c.Response().Committed {
		if c.Request().Method == echo.HEAD { // Issue #608
			if err := c.NoContent(code); err != nil {
				goto ERROR
			}
		} else {
			if err := c.JSON(code, msg); err != nil {
				goto ERROR
			}
		}
	}
ERROR:
	// I perform all my logging in the request/response area. I don't want this extra print
	// e.Logger.Error(err)
	// log.WithFields(log.Fields{"code": code, "err": err}).Errorf("Unable to respond with error code")
}

func (s *APIServer) setupStaticHandler() {
	fs := http.FileServer(WebUI)
	s.e.GET("/*", echo.WrapHandler(http.StripPrefix("/", noListForYou(fs))))

	// Make sure we default to serving the index.html
	handle, err := WebUI.Open("/index.html")
	if err != nil {
		log.Fatal(err)
	}
	buf := []byte{}
	buf, err = ioutil.ReadAll(handle)
	if err != nil {
		log.Fatal(err)
	}

	s.e.GET("/", func(c echo.Context) error {
		return c.HTML(http.StatusOK, string(buf))
	})
}

// noListForYou prevents users from performing directory listings on static assets.
func noListForYou(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.RequestURI, "/") {
			http.NotFound(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *APIServer) getState() echo.HandlerFunc {
	return func(c echo.Context) error {
		result := map[string]interface{}{
			"all_on":  true,
			"dogs_on": false,
			"cats_on": true,
		}
		return c.JSON(http.StatusOK, result)
	}
}

func (s *APIServer) setupRoutes() {
	s.e.GET("/api/v1/state", s.getState())
}

// CustomRecoverWithConfig returns a Recover middleware with config.
// See: `Recover()`.
func CustomRecoverWithConfig(config middleware.RecoverConfig) echo.MiddlewareFunc {
	// Defaults
	if config.Skipper == nil {
		config.Skipper = middleware.DefaultRecoverConfig.Skipper
	}
	if config.StackSize == 0 {
		config.StackSize = middleware.DefaultRecoverConfig.StackSize
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}

			defer func() {
				if r := recover(); r != nil {
					err, ok := r.(error)
					if !ok {
						err = fmt.Errorf("%v", r)
					}
					stack := make([]byte, config.StackSize)
					length := runtime.Stack(stack, !config.DisableStackAll)
					if !config.DisablePrintStack {
						log.Errorf("[PANIC RECOVER] %v %s\n", err, stack[:length])
					}
					c.Error(err)
				}
			}()
			return next(c)
		}
	}
}
