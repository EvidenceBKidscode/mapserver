package main

import (
	"fmt"
	"mapserver/app"
	"mapserver/mapobject"
	"mapserver/params"
	"mapserver/tilerendererjob"
	"mapserver/web"
	"runtime"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

//go:generate sh -c "go run github.com/mjibson/esc -o vfs/static.go -prefix='static/' -pkg vfs static"

func main() {
	//Parse command line

	p := params.Parse()

	if p.Debug {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}

	if p.Help {
		params.PrintHelp()
		return
	}

	if p.Version {
		fmt.Print("Mapserver version: ")
		fmt.Println(app.Version)
		fmt.Print("OS: ")
		fmt.Println(runtime.GOOS)
		fmt.Print("Architecture: ")
		fmt.Println(runtime.GOARCH)
		return
	}

	worlddir := "."
	configfilepath := "mapserver.json"

	if p.World != "" {
		worlddir = p.World
		configfilepath = filepath.Join(worlddir, configfilepath)
	}

	if p.Config != "" {
		configfilepath = p.Config
	}


	//parse Config
	cfg, err := app.ParseConfig(configfilepath)
	if err != nil {
		panic(err)
	}

	//write back config with all values
	err = cfg.Save()
	if err != nil {
		panic(err)
	}

	//exit after creating the config
	if p.CreateConfig {
		return
	}

	//setup app context
	ctx := app.Setup(p, cfg, worlddir)

	//Set up mapobject events
	mapobject.Setup(ctx)

	//run initial rendering
	if ctx.Config.EnableRendering {
		go tilerendererjob.Job(ctx)
	}

	//Start http server
	web.Serve(ctx)

}
