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
	"os"
	"os/exec"
	"strings"

	"github.com/kataras/golog"
	"github.com/labstack/echo/v4"
	md "github.com/labstack/echo/v4/middleware"
)

const (
	XForwardedFor = "X-Forwarded-For"
	XRealIP       = "X-Real-IP"
)

type Result struct {
	Version  string `json:"version,omitempty"`
	Message  string `json:"message,omitempty"`
	DeviceOn bool   `json:"deviceon,omitempty"`
	DefColor int64  `json:"defcolor,omitempty"`
	DefIcon  string `json:"deficon,omitempty"`
	Btns     []Btn  `json:"btns,omitempty"`
}

func Logger(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		var (
			code  int
			err   error
			msg   string
			valid bool
		)
		req := c.Request()
		if req.URL.String() == `/` {
			return next(c)
		}
		remoteAddr := req.RemoteAddr
		if ip := req.Header.Get(XRealIP); len(ip) > 6 {
			remoteAddr = ip
		} else if ip = req.Header.Get(XForwardedFor); len(ip) > 6 {
			remoteAddr = ip
		}
		if strings.Contains(remoteAddr, ":") {
			remoteAddr, _, _ = net.SplitHostPort(remoteAddr)
		}
		sign := strings.ToLower(c.QueryParam(`hash`))
		forHash := cfg.Password
		device := c.QueryParam(`device`)
		key := c.QueryParam(`key`)
		if len(cfg.Devices) > 0 {
			for _, device := range cfg.Devices {
				hash := md5.Sum([]byte(forHash + device + key))
				if sign == strings.ToLower(hex.EncodeToString(hash[:])) {
					valid = true
					break
				}
			}
		} else {
			hash := md5.Sum([]byte(forHash + key))
			valid = sign == strings.ToLower(hex.EncodeToString(hash[:]))
		}
		if len(device) > 0 && valid {
			err = next(c)
			if err != nil {
				code = http.StatusInternalServerError
				if he, ok := err.(*echo.HTTPError); ok {
					code = he.Code
				}
				msg = http.StatusText(code)
			} else {
				code = c.Response().Status
			}
		} else {
			code = http.StatusUnauthorized
			msg = http.StatusText(code)
		}
		if len(msg) > 0 {
			c.JSON(code, Result{Message: msg})
		}
		url := req.URL.String()
		if ind := strings.IndexByte(url, '?'); ind >= 0 {
			url = url[:ind]
		}
		out := fmt.Sprintf("%s,%s,%s,%d", url, remoteAddr, device, code)
		cmd := c.Get("cmd")
		if cmd != nil {
			out += `,` + cmd.(string)
		}
		isError := c.Get("error")
		if code != http.StatusOK || (isError != nil && isError.(bool)) {
			golog.Warn(out)
		} else {
			golog.Info(out)
		}
		return err
	}
}

func ping(c echo.Context) error {
	return c.JSON(http.StatusOK, Result{
		Version:  Version,
		DeviceOn: len(cfg.Devices) > 0,
	})
}

func list(c echo.Context) error {
	return c.JSON(http.StatusOK, Result{
		Btns:     cfg.Btns,
		DefColor: cfg.DefColor,
		DefIcon:  cfg.DefIcon,
	})
}

func run(c echo.Context) error {
	var (
		curDir string
		err    error
	)
	errRun := func(msg string) error {
		c.Set("error", true)
		return c.JSON(http.StatusOK, Result{
			Message: msg,
		})
	}
	key := c.QueryParam(`key`)
	c.Set("error", false)
	if iCmd, ok := cmds[key]; ok {
		cmd := iCmd.Cmd
		if len(iCmd.Params) > 0 {
			cmd += ` ` + fmt.Sprint(iCmd.Params)
		}
		c.Set("cmd", cmd)
		if len(iCmd.Dir) > 0 {
			curDir, _ = os.Getwd()
			if err = os.Chdir(iCmd.Dir); err != nil {
				return errRun(fmt.Sprintf("Cannot set dir %s", iCmd.Dir))
			}
			defer func() {
				os.Chdir(curDir)
			}()
		}
		if err = exec.Command(iCmd.Cmd, iCmd.Params...).Start(); err != nil {
			return errRun(fmt.Sprintf("Cannot run %s", iCmd.Cmd))
		}
		return c.JSON(http.StatusOK, Result{})
	}
	return errRun("Unknown command")
}

func RunServer() {
	e := echo.New()

	e.HideBanner = true
	e.Use(Logger)
	e.Use(md.Recover())

	e.GET("/", ping)
	e.GET("/list", list)
	e.GET("/run", run)

	if err := e.Start(fmt.Sprintf(":%d", cfg.Port)); err != nil {
		golog.Fatal(err)
	}
}
