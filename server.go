// Copyright 2019 Alexey Krivonogov. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/kataras/golog"
	"github.com/labstack/echo/v4"
	md "github.com/labstack/echo/v4/middleware"
)

const (
	XForwardedFor = "X-Forwarded-For"
	XRealIP       = "X-Real-IP"
)

func Logger(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		req := c.Request()
		remoteAddr := req.RemoteAddr
		if ip := req.Header.Get(XRealIP); len(ip) > 6 {
			remoteAddr = ip
		} else if ip = req.Header.Get(XForwardedFor); len(ip) > 6 {
			remoteAddr = ip
		}
		if strings.Contains(remoteAddr, ":") {
			remoteAddr, _, _ = net.SplitHostPort(remoteAddr)
		}
		sign := strings.ToLower(c.QueryParam(`sign`))
		device := c.QueryParam(`device`)
		id := c.QueryParam(`id`)

		hash := md5.Sum([]byte(id + device + cfg.Password))
		if sign != strings.ToLower(hex.EncodeToString(hash[:])) {

		}
		var code int
		err := next(c)
		if err != nil {
			code = http.StatusInternalServerError
			//			msg := http.StatusText(code)
			if he, ok := err.(*echo.HTTPError); ok {
				code = he.Code
				//				msg = he.Error()
			}
			/*				if !c.Response().Committed() {
							if code == http.StatusNotFound {
								c.Response().Header().Set("Status", "404 Not Found")
								c.Render(http.StatusNotFound, "404.tpl", pages[`404`])
							} else {
								http.Error(c.Response(), msg, code)
							}
						}*/
		} else {
			code = c.Response().Status
		}
		out := fmt.Sprintf("%s,%s,%s,%d", req.URL.String(), remoteAddr, device, code)
		if code != http.StatusOK {
			golog.Warn(out)
		} else {
			golog.Info(out)
		}
		return err
	}
}

func connect(c echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!")
}

func run(c echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!")
}

func RunServer() {
	e := echo.New()

	e.Use(Logger)
	e.Use(md.Recover())

	e.GET("/", connect)
	e.GET("/run", run)

	if err := e.Start(fmt.Sprintf(":%d", cfg.Port)); err != nil {
		golog.Fatal(err)
	}
}
