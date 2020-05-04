package control

import (
	"mapserver/app"
	"mapserver/mapobject"
	"mapserver/tilerendererjob"
	"mapserver/web"
	"mapserver/eventbus"
	"context"
	"sync"
)

const(
	STATUS_RUNNING = iota
	STATUS_STOPPING
	STATUS_STOPPED
)

const(
	MAP_NOTREADY = iota
	MAP_INITIAL
	MAP_INCREMENTAL
	MAP_READY
)

const(
	PIECE_WEBSERVER = iota
	PIECE_RENDERINGJOB
)

type Control struct {
	ctx *app.App
	renderingJobStatus int
	webServerStatus int
	appStatus int
	mapStatus int
	mapProgress float64
	run_job bool

	EventBus *eventbus.Eventbus

	renderingJobWaitGroup sync.WaitGroup
	webServerWaitGroup sync.WaitGroup
}

func New(ctx *app.App) *Control {
	c := Control{}
	c.ctx = ctx
	c.renderingJobStatus = STATUS_STOPPED
	c.webServerStatus = STATUS_STOPPED
	c.appStatus = STATUS_STOPPED
	c.mapStatus = MAP_NOTREADY
	c.mapProgress = 0
	c.run_job = false
	c.EventBus = eventbus.New()

	ctx.WebEventbus.AddListener(&c)

	return &c
}

func (self *Control) SetStatus(piece int, status int) {
	if piece == PIECE_RENDERINGJOB && self.renderingJobStatus != status {
		self.renderingJobStatus = status
		self.EventBus.Emit("rendering-job-status-changed", status)
	}
	if piece == PIECE_WEBSERVER && self.webServerStatus != status {
		self.webServerStatus = status
		self.EventBus.Emit("web-server-status-changed", status)
	}

	var newStatus int
	if self.renderingJobStatus < self.webServerStatus {
		newStatus = self.renderingJobStatus
	} else {
		newStatus = self.webServerStatus
	}

	if newStatus != self.appStatus {
		self.appStatus = newStatus
		self.EventBus.Emit("app-status-changed", 0)
	}
}

func (self *Control) RenderingJobShouldRun() bool {
	return self.renderingJobStatus == STATUS_RUNNING
}

func (self *Control) Run() {
	if self.Status() != STATUS_STOPPED {
		return
	}
	//Set up mapobject events
	mapobject.Setup(self.ctx)

	//Start rendering job
	if self.ctx.Config.EnableRendering {
		self.renderingJobWaitGroup.Add(1)
		go func() {
			self.SetStatus(PIECE_RENDERINGJOB, STATUS_RUNNING)
			self.run_job = true
			tilerendererjob.Job(self.ctx, &self.run_job)
			self.SetStatus(PIECE_RENDERINGJOB, STATUS_STOPPED)
			self.renderingJobWaitGroup.Done()
		}()
	}

	//Start http server
	if self.webServerStatus == STATUS_STOPPED {
		self.webServerWaitGroup.Add(1)
		go func() {
			self.SetStatus(PIECE_WEBSERVER, STATUS_RUNNING)
			web.Serve(self.ctx)
			self.SetStatus(PIECE_WEBSERVER, STATUS_STOPPED)
			self.webServerWaitGroup.Done()
		}()
	}
}

func (self *Control) Stop() {
	if self.renderingJobStatus == STATUS_RUNNING {
		self.SetStatus(PIECE_RENDERINGJOB, STATUS_STOPPING)
		self.run_job = false
	}

	if self.webServerStatus == STATUS_RUNNING {
		self.SetStatus(PIECE_WEBSERVER, STATUS_STOPPING)
		self.ctx.WebServer.Shutdown(context.Background())
	}
}

func (self *Control) Wait() {
	self.renderingJobWaitGroup.Wait()
	self.webServerWaitGroup.Wait()
}

func (self *Control) Status() int {
	return self.appStatus
}

func (self *Control) MapStatus() int {
	return self.mapStatus
}

func (self *Control) MapProgress() float64 {
	return self.mapProgress
}

// Listen WebEventBus to make Control.EventBus events
func (self *Control) OnEvent(eventtype string, o interface{}) {
	mapStatus := self.mapStatus
	mapProgress := self.mapProgress

	if eventtype == "initial-render-progress" {
		ev := o.(*tilerendererjob.InitialRenderEvent)
		if ev.Progress < 1 {
			mapStatus = MAP_INITIAL
			mapProgress = ev.Progress
		} else {
			mapStatus = MAP_READY
			mapProgress = 1
		}
	}

	if eventtype == "incremental-render-progress" {
		ev := o.(*tilerendererjob.IncrementalRenderEvent)
		if ev.Progress < 1 {
			mapStatus = MAP_INCREMENTAL
			mapProgress = ev.Progress
		} else {
			mapStatus = MAP_READY
			mapProgress = 1
		}
	}

	if mapStatus != self.mapStatus || mapProgress != self.mapProgress {
		self.mapStatus = mapStatus
		self.mapProgress = mapProgress
		self.EventBus.Emit("map-status-changed", 0)
	}
}
