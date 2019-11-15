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
	Version        = "1.0.0"
	DefaultCfgName = `run-buttons.yaml`
	DefaultPort    = 1321
	DefaultColor   = 0x006699
	DefaultIcon    = `play_circle_outline`
)

type Btn struct {
	Key    string   `yaml:"key" json:"key"`                 // optional key
	Cmd    string   `yaml:"cmd" json:"cmd"`                 // application
	Params []string `yaml:"params" json:"params,omitempty"` // command line parameters
	Dir    string   `yaml:"dir" json:"dir,omitempty"`       // working directory
	Title  string   `yaml:"title" json:"title,omitempty"`   // title of the button
	Desc   string   `yaml:"desc" json:"desc,omitempty"`     // description
	Color  int64    `yaml:"color" json:"color,omitempty"`   // color of the icon
	Icon   string   `yaml:"icon" json:"icon,omitempty"`     // material icon
}

type Settings struct {
	Password string   `yaml:"password"`
	Port     int      `yaml:"port"`     // if empty, then DefaultPort is used
	LogFile  string   `yaml:"logfile"`  // name of the logfile
	DefColor int64    `yaml:"defcolor"` // default color of icons
	DefIcon  string   `yaml:"deficon"`  // default icon
	Devices  []string `yaml:"devices"`  // allowed mobile devices. if empty, then any phone can  connect.
	Btns     []Btn    `yaml:"btns"`     // the list of buttons

	//	LocalOnly bool     `yaml:"localonly"`
}

var (
	cfg  Settings
	cmds = make(map[string]*Btn)
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
	if cfg.DefColor == 0 {
		cfg.DefColor = DefaultColor
	}
	if len(cfg.DefIcon) == 0 {
		cfg.DefIcon = DefaultIcon
	}
	if len(cfg.Btns) == 0 {
		golog.Fatal(`'btns' is undefined in settings file`)
	}
	for i, item := range cfg.Btns {
		cfg.Btns[i].Key = fmt.Sprint(crc32.ChecksumIEEE([]byte(item.Cmd +
			strings.Join(item.Params, ``) + item.Dir)))
		if _, ok := cmds[item.Key]; ok {
			golog.Fatalf(`btns[%d].key %s is duplicated`, i, item.Key)
		}
		if len(item.Cmd) == 0 {
			golog.Fatalf(`btns[%d].cmd is empty`, i)
		}
		if len(item.Title) == 0 {
			cfg.Btns[i].Title = item.Cmd
			if len(item.Desc) == 0 {
				cfg.Btns[i].Desc = strings.Join(item.Params, ` `)
			}
		} else if len(item.Desc) == 0 {
			cfg.Btns[i].Desc = item.Cmd
		}

		cmds[cfg.Btns[i].Key] = &cfg.Btns[i]
	}
	localIP := GetLocalIP()
	if cfg.Port != DefaultPort {
		localIP += fmt.Sprintf(`:%d`, cfg.Port)
	}
	if len(cfg.LogFile) > 0 {
		f, err := os.OpenFile(cfg.LogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			golog.Fatal(err)
		}
		golog.Infof("Log output has been redirected to %s", cfg.LogFile)
		golog.SetOutput(f)
		golog.Infof("Started Run Buttons %s:%d", localIP, cfg.Port)
	}
	fmt.Println(`==========================================
     IP-address: ` + localIP + `
==========================================`)
	RunServer()
}
