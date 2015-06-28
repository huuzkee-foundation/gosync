// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// +build !plan9,!solaris
package main

import (
	"flag"
	log "github.com/cihub/seelog"
	"github.com/Nitecon/gosync/fswatcher"
	"github.com/Nitecon/gosync/nodeinfo"
	"github.com/Nitecon/gosync/replicator"
	"github.com/Nitecon/gosync/utils"
	"net/http"
	"os"
    "path/filepath"
)

func getLoggerConfig() string {
	cfg := utils.GetConfig()
	var loggerConfig = ""
	if cfg.ServerConfig.LogLocation != "stdout" {
        if _, err := os.Stat(filepath.Dir(cfg.ServerConfig.LogLocation)); os.IsNotExist(err) {
            os.Mkdir(filepath.Dir(cfg.ServerConfig.LogLocation), 0775)
        }
		loggerConfig = `<seelog type="asynctimer" asyncinterval="1000">
    <outputs formatid="main">
        <filter levels="`+cfg.ServerConfig.LogLevel+`">
          <file path="` + cfg.ServerConfig.LogLocation + `" />
        </filter>
    </outputs>
    <formats>
        <format id="main" format="%Date %Time [%LEVEL] %Msg%n"/>
    </formats>
    </seelog>`
	} else {
		loggerConfig = `<seelog type="asynctimer" asyncinterval="1000">
    <outputs formatid="main">
        <console/>
    </outputs>
    <formats>
        <format id="main" format="%Date %Time [%LEVEL] %Msg (%RelFile:%Func)%n"/>
    </formats>
    </seelog>`
	}

	return loggerConfig
}

func StartWebFileServer(cfg *utils.Configuration) {
	nodeinfo.Initialize()
	log.Info("Starting Web File server and setting node as active")
	nodeinfo.SetAlive()

	var listenPort = ":" + cfg.ServerConfig.ListenPort
	for name, item := range cfg.Listeners {
		var section = "/" + name + "/"
		log.Infof("Adding section listener: %s, to serve directory: %s", section, item.Directory)
		http.Handle(section, http.StripPrefix(section, http.FileServer(http.Dir(item.Directory))))
	}
	log.Debug(http.ListenAndServe(listenPort, nil))
}

func init() {
	var ConfigFile string
	flag.StringVar(&ConfigFile, "config", "config.cfg.example",
		"Please provide the path to the config file, defaults to: /etc/gosync/config.cfg")
	flag.Parse()
	if _, err := os.Stat(ConfigFile); os.IsNotExist(err) {
		log.Criticalf("Configuration file does not exist or cannot be loaded: (%s)", ConfigFile)
        os.Exit(1)
	} else {
		utils.ReadConfigFromFile(ConfigFile)
	}

	logger, err := log.LoggerFromConfigAsString(getLoggerConfig())

	if err == nil {
		log.ReplaceLogger(logger)
	}

}

func main() {

	cfg := utils.GetConfig()
	replicator.InitialSync()
	for _, item := range cfg.Listeners {
		log.Info("Working with: " + item.Directory)
		go replicator.CheckIn(item.Directory)
		go fswatcher.SysPathWatcher(item.Directory)
	}
	StartWebFileServer(cfg)

}
