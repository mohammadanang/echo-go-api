package main

import (
	"fmt"
	"net/http"

	"time"

	"github.com/labstack/echo"
	log "github.com/sirupsen/logrus"
)

type M map[string]interface{}

func makeLogEntry(c echo.Context) *log.Entry {
	if c == nil {
		return log.WithFields(log.Fields{
			"at": time.Now().Format("2006-01-02 15:04:05"),
		})
	}

	return log.WithFields(log.Fields{
		"at":     time.Now().Format("2006-01-02 15:04:05"),
		"method": c.Request().Method,
		"uri":    c.Request().URL.String(),
		"ip":     c.Request().RemoteAddr,
	})
}

func middlewareLogging(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		makeLogEntry(c).Info("incoming request")
		return next(c)
	}
}

func errorHandler(err error, c echo.Context) {
	report, ok := err.(*echo.HTTPError)
	if ok {
		report.Message = fmt.Sprintf("http error %d - %v", report.Code, report.Message)
	} else {
		report = echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	makeLogEntry(c).Error(report.Message)
	c.HTML(report.Code, report.Message.(string))
}

func main() {
	r := echo.New()

	// middleware here
	r.Use(middlewareLogging)
	r.HTTPErrorHandler = errorHandler

	r.GET("/", func(c echo.Context) error {
		data := M{"message": "Hello", "status": "OK", "code": 200}

		return c.JSON(http.StatusOK, data)
	})

	lock := make(chan error)
	go func(lock chan error) { lock <- r.Start(":9000") }(lock)

	time.Sleep(1 * time.Millisecond)
	makeLogEntry(nil).Warning("application started without ssl/tls enabled")

	err := <-lock
	if err != nil {
		makeLogEntry(nil).Panic("failed to start application")
	}
}
