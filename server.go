package main

import (
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"text/template"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	flag "github.com/spf13/pflag"
)

// StartTime gives the start time of server
var StartTime = time.Now()

const defaultAppPort string = "1323"

func uptime() string {
	elapsedTime := time.Since(StartTime)
	return fmt.Sprintf("%d:%d:%d", int(math.Round(elapsedTime.Hours())), int(math.Round(elapsedTime.Minutes())), int(math.Round(elapsedTime.Seconds())))
}

func homePage(c echo.Context) error {
	host, _ := os.Hostname()
	return c.Render(http.StatusOK, "index.html", map[string]interface{}{
		"host":           fmt.Sprintf("[HOST: %s] (uptime: %s)]", host, uptime()),
		"requestHeaders": c.Request().Header,
		"response":       "Hello Universe",
	})
}

// Render renders a template document
func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func main() {
	var requiredLogLevel string
	flag.StringVarP(&requiredLogLevel, "log-level", "l", "info", "Required log level: debug/info/warn/error. Defaults to info")
	var printHelp bool
	flag.BoolVarP(&printHelp, "help", "h", false, "Prints this help content.")
	flag.Parse()
	if printHelp {
		flag.Usage()
		return
	}
	e := echo.New()
	renderer := &TemplateRenderer{
		templates: template.Must(template.ParseGlob("templates/*.html")),
	}
	e.HideBanner = true
	e.Debug = true
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: `{"time":"${time_rfc3339}","id":"${id}","host":"${host}",` +
			`,"uri":"${uri}","status":${status},"error":"${error}","latency":${latency_human},` +
			`"bytes_out":${bytes_out}}` + "\n",
		Output: os.Stdout,
	}))
	switch requiredLogLevel {
	case "debug":
		e.Logger.SetLevel(log.DEBUG)
	case "info":
		e.Logger.SetLevel(log.INFO)
	case "warn":
		e.Logger.SetLevel(log.WARN)
	case "error":
		e.Logger.SetLevel(log.ERROR)
	default:
		e.Logger.SetLevel(log.INFO)
	}
	e.Logger.Debugf("Loglevel is set to %s", requiredLogLevel)
	e.Logger.SetLevel(log.DEBUG)
	var port, isEnvVarSet = os.LookupEnv("APP_PORT")
	if !isEnvVarSet {
		port = defaultAppPort
		e.Logger.Infof("Port is defaulted to %s", port)
	}
	e.Renderer = renderer
	e.GET("/", homePage)
	e.Logger.Fatal(e.Start(fmt.Sprintf("[::]:%s", port)))
}