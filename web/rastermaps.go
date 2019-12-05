package web

import (
	"mapserver/app"
	"net/http"
	"strings"
	"os"
	"path/filepath"
	"io"
	"github.com/sirupsen/logrus"
)

type RasterMaps struct {
	ctx *app.App
}

func (h *RasterMaps) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	str := strings.TrimPrefix(req.URL.Path, "/api/rastermaps/")
	parts := strings.Split(str, "/")

	if len(parts) != 1 {
		resp.WriteHeader(404)
		return
	}

	filename := filepath.Join(h.ctx.WorldDir, "worldmods", "minimap",
			"textures", parts[0])

	file, err := os.Open(filename)
	if err != nil {
		log.WithFields(logrus.Fields{"error": err}).Error("Serve raster file")
		resp.WriteHeader(404)
		resp.Write([]byte(parts[0]))
		return
	}
	defer file.Close()

	resp.Header().Set("Content-Type", "image/png")

	// Unfortunately on windows (or some windows?) using io.Copy leads to a "not
	// implemented" error
	buf := make([]byte, 16384)

	for {
		n, err := file.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.WithFields(logrus.Fields{"error": err}).Error("Serve raster file")
			resp.WriteHeader(500)
			return
		}
		resp.Write(buf[:n])
	}
}
