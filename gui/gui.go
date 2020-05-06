// +build gui

// Use go build -tags gui to build GUI

// On linux, xorg-dev needed to compile

package gui

import (
	"fyne.io/fyne"
	fyneapp "fyne.io/fyne/app"
	"fyne.io/fyne/widget"
	"path/filepath"
	"io/ioutil"
	"os"
	"mapserver/app"
	"mapserver/params"
	"mapserver/control"
	"net/url"
)

var Active = true

func checkWorld(path string) bool {
	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		return false
	}
	_, err2 := os.Stat(path + "/map.sqlite")

	if err2 != nil {
		return false
	}
	return true
}

type Gui struct {
	app fyne.App
	worldpath string
	params params.ParamsType
	window fyne.Window
	status_bar *widget.ProgressBar
	status_text *widget.Label
	link *widget.Hyperlink

	control *control.Control
}

func (self *Gui) startMapServer() {

	//parse Config
	cfg, err := app.ParseConfig(filepath.Join(self.worldpath, "mapserver.json"))
	if err != nil {
		panic(err)
	}

	self.status_text.SetText("Cartographe lancé.")
	self.link.Show()

	//setup app context
	ctx := app.Setup(self.params, cfg, self.worldpath)

	//control app
	self.control = control.New(ctx)

	// listen for control events
	self.control.EventBus.AddListener(self)

	// Run mapserver!
	self.control.Run()
}

func (self *Gui) Run(p params.ParamsType) {
	worldbasepath := filepath.Join(getUserPath(), "worlds")
	self.app = fyneapp.New()
	self.params = p

	// Show main window
	self.window = self.app.NewWindow("Cartographe Kidscode")
	self.window.SetPadded(true)
	self.status_text = widget.NewLabel("Cartographe non lancé.")
	self.status_bar = widget.NewProgressBar()
	self.status_bar.Hide()
	self.link = widget.NewHyperlink("Lien du cartographe", &url.URL{Scheme: "http", Host: "localhost:8080"})
	self.link.Hide()

	self.window.SetContent(widget.NewVBox(
		self.status_text,
		self.status_bar,
		self.link,
		widget.NewButton("Quitter", func() { self.app.Quit() }),
	))
	self.window.Show()

	// World chosen from command line
	if p.World != "" {
		if checkWorld(p.World) {
			self.worldpath = p.World
			go self.startMapServer()
		}
		if checkWorld(filepath.Join(worldbasepath, p.World)) {
			self.worldpath = filepath.Join(worldbasepath, p.World)
			go self.startMapServer()
		}
	}
	if self.worldpath == "" {
		// Find available worlds
		var worlds []string
		files, err := ioutil.ReadDir(worldbasepath)
		if err == nil {
			for _, file := range files {
				if checkWorld(filepath.Join(worldbasepath, file.Name())) {
					worlds = append(worlds, file.Name())
				}
			}
		}

		// No world
		if len(worlds) == 0 {
			self.window.SetContent(widget.NewVBox(
				widget.NewLabel("Désolé, aucun monde trouvé."),
				widget.NewButton("Quitter", func() { self.app.Quit() }),
			))
		// Only one world, start on it
		} else if len(worlds) == 1 {
				self.worldpath = filepath.Join(worldbasepath, worlds[0])
				go self.startMapServer()
		// Several world, ask to choose
		} else {
			w := self.app.NewWindow("Cartographe Kidscode")
			launch := widget.NewButton("Lancer le cartographe", func() {
				self.window.Show()
				w.Hide()
				go self.startMapServer()
			})
			launch.Disable()

			radio := widget.NewRadio(worlds, func(s string) {
				self.worldpath = filepath.Join(worldbasepath, s)
				launch.Enable()
			})
			radio.Horizontal = false

			w.SetContent(widget.NewVBox(
				widget.NewLabel("Choisir un monde :"),
				radio, launch,
				widget.NewButton("Annuler", func() { self.app.Quit() }),
			))
			w.Show()
			self.window.Hide()
		}
	}

	self.app.Run()
}


func (self *Gui) OnEvent(eventtype string, ev interface{}) {
	switch eventtype {
	case "map-status-changed":
		status := self.control.MapStatus()
		progress := self.control.MapProgress()

		switch(status) {
		case control.MAP_NOTREADY:
			self.status_text.SetText("En cours d'initialisation")
			self.status_bar.Hide()
		case control.MAP_INITIAL:
			self.status_text.SetText("Rendu initial de la carte")
			self.status_bar.Show()
			self.status_bar.SetValue(progress)
		case control.MAP_INCREMENTAL:
			self.status_text.SetText("Rendu initial de la carte")
			self.status_bar.Show()
			self.status_bar.SetValue(progress)
		case control.MAP_READY:
			self.status_text.SetText("Cartographe lancé")
			self.status_bar.Hide()
		}
	}
}
