package main

import (
	"fmt"
	"mapserver/control"
	"mapserver/app"
	"mapserver/params"
	"runtime"
//	"mapserver/gui"
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

//	if p.NoGui {
		Run(p)
//	} else {
//		g := gui.Gui{}
//		g.Run(p)
//	}
}

func Run(p params.ParamsType) {

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

	//exit after creating the config
	if p.CreateConfig {
		return
	}

	//setup app context
	ctx := app.Setup(p, cfg, worlddir)

	//control it
	ctrl := control.New(ctx)

	ctrl.Run()
	ctrl.Wait()
}
