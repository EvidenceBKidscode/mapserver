package web

import (
//	"encoding/json"
	"mapserver/app"
	"net/http"
	"fmt"
	"os"
	"io/ioutil"
	"path/filepath"
)

type Draw struct {
	ctx *app.App
}

func (t *Draw) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	fmt.Printf("HTTP draw request\n")
	f, err := os.Create(filepath.Join(t.ctx.WorldDir, "zones.json"))
	if err != nil {
			fmt.Println(err)
			return
	}
	defer req.Body.Close()
	bodydata, err := ioutil.ReadAll(req.Body)
	body := string(bodydata)

	fmt.Printf("Body: %s\n",body)

	l, err := f.WriteString(body)
	if err != nil {
			fmt.Println(err)
			f.Close()
			return
	}
	fmt.Println(l, "bytes written successfully")
	err = f.Close()
	if err != nil {
			fmt.Println(err)
			return
	}

	resp.Header().Add("content-type", "text/plain")
}
