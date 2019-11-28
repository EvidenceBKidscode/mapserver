package web

import (
	"mapserver/app"
	"net/http"
	"strings"
	"os"
	"path"
	"io"
)

type RasterMaps struct {
	ctx *app.App
}

func (h *RasterMaps) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	str := strings.TrimPrefix(req.URL.Path, "/api/rastermaps/")
	parts := strings.Split(str, "/")
	if len(parts) != 1 {
		resp.WriteHeader(500)
		resp.Write([]byte("wrong number of arguments"))
		return
	}

	filename := parts[0]
	file, err := os.Open(path.Join(h.ctx.WorldDir, "worldmods", "minimap", "textures", filename))
	if err != nil {
		resp.WriteHeader(404)
		resp.Write([]byte("Not found"))
		resp.Write([]byte(filename))
		return
	}

	defer file.Close()
	resp.Header().Set("Content-Type", "image/png")
	_, err = io.Copy(resp, file);
	if err != nil {
		resp.WriteHeader(500)
		resp.Write([]byte("error while sending data"))
		return
	}
}
