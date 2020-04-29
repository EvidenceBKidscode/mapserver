package main

// Compile using following command to build .so and .h files :
// go build -o mapserver.so -buildmode=c-archive

import "C"

import (
	"mapserver/control"
	"mapserver/app"
	"mapserver/params"
	"path/filepath"
)

var the_app *app.App
var the_control *control.Control

//export MapserverRun
func MapserverRun(worldpath string) {

	//parse Config
	cfg, err := app.ParseConfig(filepath.Join(worldpath, "mapserver.json"))
	if err != nil {
		panic(err)
	}

	//setup app context
	the_app := app.Setup(params.ParamsType{}, cfg, worldpath)

	//control app
	the_control = control.New(the_app)

	// Run mapserver!
	the_control.Run()
}

//export MapserverStop
func MapserverStop() {
	the_control.Stop()
}
