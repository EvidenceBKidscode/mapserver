package main

// Compile using following command to build .so and .h files :
// go build -o mapserver.a -buildmode=c-archive

/*
typedef void (*EventCallbackType)(int i);
void callCallback(EventCallbackType callback, int value);
*/
import "C"

import (
	"mapserver/control"
	"mapserver/app"
	"mapserver/params"
	"path/filepath"
)

type EventCallbackType func(C.int)

type ControlListener struct {}

var the_app *app.App
var the_control *control.Control
var listeners []C.EventCallbackType
var control_listener = ControlListener{}

//export MapserverRun
func MapserverRun(worldpath string) {

	if MapserverStatus() != control.STATUS_STOPPED {
		return
	}

	//parse Config
	cfg, err := app.ParseConfig(filepath.Join(worldpath, "mapserver.json"))
	if err != nil {
		panic(err)
	}

	//setup app context
	the_app := app.Setup(params.ParamsType{}, cfg, worldpath)

	//control app
	the_control = control.New(the_app)

	// listen for control events
	the_control.EventBus.AddListener(&control_listener)

	// Run mapserver!
	the_control.Run()
}

//export MapserverStop
func MapserverStop() {
	the_control.Stop()
}

//export MapserverStatus
func MapserverStatus() int {
	if the_control == nil {
		return control.STATUS_STOPPED
	} else {
		return the_control.Status()
	}
}

//export MapserverListen
func MapserverListen(l C.EventCallbackType) {
	listeners = append(listeners, l)
}

func (self *ControlListener) OnEvent(eventtype string, o interface{}) {
	if eventtype == "rendering-job-status-changed" ||
			eventtype == "web-server-status-changed" {
		for _, l := range listeners {
			C.callCallback(l, C.int(MapserverStatus()))
		}
	}
}
