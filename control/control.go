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
	PIECE_WEBSERVER = iota
	PIECE_RENDERINGJOB
)

type Control struct {
	ctx *app.App
	renderingJobStatus int
	webServerStatus int
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
	c.run_job = false
	c.EventBus = eventbus.New()

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
	if self.renderingJobStatus < self.webServerStatus {
		return self.renderingJobStatus
	} else {
		return self.webServerStatus
	}
}
