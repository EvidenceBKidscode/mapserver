package main

// Compile using following command to build .so and .h files :
// go build -o mapserver.a -buildmode=c-archive

/*
#include <stdlib.h>

typedef void (*EventCallbackType)(int);
void callCallback(EventCallbackType callback, int);
*/
import "C"

import (
	"mapserver/control"
	"mapserver/app"
	"mapserver/params"
	"path/filepath"
)

type EventCallbackType func(int)

type ControlListener struct {}

var the_app *app.App
var the_control *control.Control
var listeners []C.EventCallbackType
var control_listener = ControlListener{}

var mapstatus = "starting"
var mapprogress = 0

//export MapserverRun
func MapserverRun(cworldpath *C.char) {

	if MapserverStatus() != control.STATUS_STOPPED {
		return
	}

	worldpath := C.GoString(cworldpath)

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

//export MapserverMapStatus
func MapserverMapStatus() int {
	if the_control == nil {
		return control.MAP_NOTREADY
	} else {
		return the_control.MapStatus()
	}
}

//export MapserverMapProgress
func MapserverMapProgress() float64 {
	if the_control == nil {
		return 0
	} else {
		return the_control.MapProgress()
	}
}

//export MapserverListen
func MapserverListen(l C.EventCallbackType) {
	listeners = append(listeners, l)
}

// Would have been much better to send an event name but wasn't able to do it.
// when passing C.char* C.CString(xxx) gives a 32 bits truncated pointer on
// C side provoking a SEGFAULT when used.
func emit(eventtype int) {
	for _, l := range listeners {
		C.callCallback(l, C.int(eventtype))
	}
}

func (self *ControlListener) OnEvent(eventtype string, o interface{}) {
	if eventtype == "app-status-changed" {
		emit(1)
	}
	if eventtype ==  "map-status-changed" {
		emit(2)
	}
}


//export MapserverGetPort
func MapserverGetPort() int {
	return the_app.Config.Port
}

//export MapserverGetWorldDir
func MapserverGetWorldDir() string {
	return the_app.WorldDir
}
