package wabmw

import (
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

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
						zap.S().Errorf("[PANIC RECOVER] %v %s\n", err, stack[:length])
					}
					c.Error(err)
				}
			}()
			return next(c)
		}
	}
}

const (
	padding int = 6
)

func ZapMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ectx echo.Context) error {
			req := ectx.Request()
			methodColor := req.Method

			switch req.Method {
			case http.MethodGet:
				methodColor = green(req.Method)
			case http.MethodPost:
				methodColor = cyan(req.Method)
			case http.MethodPut:
				methodColor = orange(req.Method)
			case http.MethodDelete:
				methodColor = red(req.Method)
			}

			fmtString := fmt.Sprintf("%s %%-%ds %%s", "-->", len(methodColor)-len(req.Method)+padding)
			zap.S().Infof(fmtString, methodColor, req.URL.Path)

			start := time.Now()
			err := next(ectx)
			stop := time.Now()

			var statusCode int
			if err != nil {
				statusCode = http.StatusInternalServerError
				if he, ok := err.(*echo.HTTPError); ok {
					statusCode = he.Code
				}
			} else {
				statusCode = ectx.Response().Status
			}

			statusColor := func(text string) string { return text }
			logFcn := zap.S().Infof

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
