// Copyright 2019 Alexey Krivonogov. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package main

import (
	"fmt"
	"hash/crc32"
	"io/ioutil"
	"os"
	"strings"

	"github.com/kataras/golog"
	"gopkg.in/yaml.v3"
)

const (
	Version        = `1.0.0`
	DefaultCfgName = `run-buttons.yaml`
	DefaultPort    = 1321
)

type Run struct {
	Key    string   `yaml:"key"`    // optional key
	Cmd    string   `yaml:"cmd"`    // application
	Params []string `yaml:"params"` // command line parameters
	Dir    string   `yaml:"dir"`    // working directory
	Title  string   `yaml:"title"`  // title of the button
	Desc   string   `yaml:"desc"`   // description
	Color  int64    `yaml:"color"`  // color of the icon
}

type Settings struct {
	Password string   `yaml:"password"`
	Port     int      `yaml:"port"`    // if empty, then DefaultPort is used
	LogFile  string   `yaml:"logfile"` // name of the logfile
	Devices  []string `yaml:"devices"` // allowed mobile devices. if empty, then any phone can  connect.
	Runs     []Run    `yaml:"runs"`    // the list of buttons

	//	LocalOnly bool     `yaml:"localonly"`
}

var (
	cfg  Settings
	cmds = make(map[string]*Run)
)

func main() {
	golog.SetTimeFormat("2006/01/02 15:04:05")
	cfgFile := DefaultCfgName
	if len(os.Args) > 1 {
		cfgFile = os.Args[1]
	}

	cfgData, err := ioutil.ReadFile(cfgFile)
	if err != nil {
		golog.Fatal(err)
	}
	if err = yaml.Unmarshal(cfgData, &cfg); err != nil {
		golog.Fatal(err)
	}
	if cfg.Port == 0 {
		cfg.Port = DefaultPort
	}
	if len(cfg.Runs) == 0 {
		golog.Fatal(`'runs' is undefined in settings file`)
	}
	for i, item := range cfg.Runs {
		if len(item.Key) == 0 {
			cfg.Runs[i].Key = fmt.Sprint(crc32.ChecksumIEEE([]byte(item.Cmd +
				strings.Join(item.Params, ``) + item.Dir)))
		}
		if _, ok := cmds[item.Key]; ok {
			golog.Fatalf(`runs[%d].key %s is duplicated`, i, item.Key)
		}
		if len(item.Cmd) == 0 {
			golog.Fatalf(`runs[%d].cmd is empty`, i)
		}
		cmds[cfg.Runs[i].Key] = &cfg.Runs[i]
	}
	golog.Println(cfg, cmds)
	golog.Infof("Started Run Buttons IP-address: %s, Port: %d", GetLocalIP(), cfg.Port)

	RunServer()
}
