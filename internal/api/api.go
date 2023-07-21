package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/modfin/henry/slicez"
	"github.com/prometheus/client_golang/prometheus"
	"idly"
	"idly/internal/config"
	"idly/internal/dao"
	"time"
)

func Start(c context.Context, db dao.DAO) {
	e := echo.New()

	e.Use(middleware.Recover(), middleware.Logger())
	e.Use(echoprometheus.NewMiddleware("idly"))    // adds middleware to gather metrics
	e.GET("/metrics", echoprometheus.NewHandler()) // adds route to serve gathered metrics

	failedLoginCounter := prometheus.NewCounterVec( // create new counter metric. This is replacement for `prometheus.Metric` struct
		prometheus.CounterOpts{
			Name: "logins_failed",
			Help: "Numer of failed logins",
		}, []string{"service"},
	)
	knownLoginCounter := prometheus.NewCounterVec( // create new counter metric. This is replacement for `prometheus.Metric` struct
		prometheus.CounterOpts{
			Name: "logins_known",
			Help: "Numer of known logins",
		}, []string{"service"},
	)
	unknownLoginCounter := prometheus.NewCounterVec( // create new counter metric. This is replacement for `prometheus.Metric` struct
		prometheus.CounterOpts{
			Name: "logins_unknown",
			Help: "Numer of unknown logins",
		}, []string{"service"},
	)

	prometheus.MustRegister(failedLoginCounter, knownLoginCounter, unknownLoginCounter)

	e.POST("/login-fail", func(c echo.Context) error {
		var login idly.Login
		err := json.NewDecoder(c.Request().Body).Decode(&login)
		_ = c.Request().Body.Close()
		if err != nil {
			return err
		}
		fmt.Printf("Failed login for %s, from %s\n", login.Prefix(), login.IPAddress)
		failedLoginCounter.With(map[string]string{"service": login.Service}).Inc()
		return nil
	})

	// TODO access key?
	e.GET("/login/:service/:uid", func(c echo.Context) error {
		service := c.Param("service")
		uid := c.Param("uid")
		logins, err := db.ListLogins(fmt.Sprintf("%s/%s", service, uid))
		if err != nil {
			return err
		}
		// TODO format this as, who, when, where, and what device...
		return c.JSON(200, logins)
	})

	e.POST("/login", func(c echo.Context) error {

		var err error
		status := struct {
			Status string `json:"status"`
		}{
			Status: "logged",
		}
		defer func() {
			code := 200
			if err != nil {
				code = 500
				status.Status = "error"
			}
			_ = c.JSON(code, status)
		}()

		var login idly.Login
		err = json.NewDecoder(c.Request().Body).Decode(&login)
		_ = c.Request().Body.Close()
		if err != nil {
			return err
		}

		if !idly.IsIPAddress(login.IPAddress) {
			err = fmt.Errorf("supplied ip address, %s, is not an ipv4 address", login.IPAddress)
			return err
		}

		if !idly.IsEmail(login.Email) {
			err = fmt.Errorf("a valid email address were not supplied, (%s)", login.Email)
			return err
		}

		login.At = time.Now()
		login.SentAlert = false
		if login.UID == "" {
			login.UID = login.Email
		}

		logins, err := db.ListLogins(login.Prefix())
		if err != nil {
			return err
		}

		err = db.StoreLogin(login, config.Get().LoginTTL)
		if err != nil {
			return err
		}

		// The first login and we should Not alert
		if len(logins) == 0 && !config.Get().AlertOnInit {
			knownLoginCounter.With(map[string]string{"service": login.Service}).Inc()
			return nil
		}

		// The first login and we should alert
		if len(logins) == 0 {
			unknownLoginCounter.With(map[string]string{"service": login.Service}).Inc()
			status.Status = "alert sent"
			login.SentAlert = true
			err = Alert(login)
			if err != nil {
				return err
			}
			return db.StoreLogin(login, config.Get().LoginTTL)
		}

		if Suspect(login, logins) {
			unknownLoginCounter.With(map[string]string{"service": login.Service}).Inc()
			status.Status = "alert sent"
			login.SentAlert = true
			err = Alert(login)
			if err != nil {
				return err
			}
			return db.StoreLogin(login, config.Get().LoginTTL)
		}
		knownLoginCounter.With(map[string]string{"service": login.Service}).Inc()

		return nil
	})

	go func() {
		<-c.Done()
		cc, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		fmt.Println("[Shutdown] Shutting down web server")
		_ = e.Shutdown(cc)
		fmt.Println("[Shutdown] Webserver is shutdown")
	}()

	_ = e.Start(fmt.Sprintf(":%d", config.Get().HttpPort))

	return
}

func Suspect(login idly.Login, logins []idly.Login) bool {
	return !slicez.ContainsFunc(logins, func(e idly.Login) bool {
		return e.IPAddress == login.IPAddress
	})
}
