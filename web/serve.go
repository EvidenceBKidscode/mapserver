package web

import (
	"mapserver/app"
	"mapserver/vfs"
	"mapserver/upnp"
	"net/http"
	"strconv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

func Serve(ctx *app.App) {
	fields := logrus.Fields{
		"port":   ctx.Config.Port,
		"webdev": ctx.Config.Webdev,
	}
	logrus.WithFields(fields).Info("Starting http server")

	// UPNP Announce
//	upnp.Announce(ctx)

	mux := http.NewServeMux()

	mux.Handle("/", http.FileServer(vfs.FS(ctx.Config.Webdev)))

	tiles := &Tiles{ctx: ctx}
	tiles.Init()
	mux.Handle("/upnp/", &upnp.UpnpHandler{Ctx: ctx})
	mux.Handle("/api/tile/", tiles)
	mux.Handle("/api/config", &ConfigHandler{ctx: ctx})
	mux.Handle("/api/media/", &MediaHandler{ctx: ctx})
	mux.Handle("/api/minetest", &Minetest{ctx: ctx})
	mux.Handle("/api/mapobjects/", &MapObjects{ctx: ctx})
	mux.Handle("/api/rastermaps/", &RasterMaps{ctx: ctx})

	if ctx.Config.MapObjects.Areas {
		mux.Handle("/api/areas", &AreasHandler{ctx: ctx})
	}

	if ctx.Config.EnablePrometheus {
		mux.Handle("/metrics", promhttp.Handler())
	}

	ws := NewWS(ctx)
	mux.Handle("/api/ws", ws)

	ctx.Tilerenderer.Eventbus.AddListener(ws)
	ctx.WebEventbus.AddListener(ws)

	if ctx.Config.WebApi.EnableMapblock {
		//mapblock endpoint
		mux.Handle("/api/mapblock/", &MapblockHandler{ctx: ctx})
	}

	mux.Handle("/api/draw/", &Draw{ctx: ctx})

	ctx.WebServer = &http.Server{
		Addr: ":"+strconv.Itoa(ctx.Config.Port),
		Handler:mux}

	err := ctx.WebServer.ListenAndServe()

	if err != http.ErrServerClosed {
		panic(err)
	}
}
