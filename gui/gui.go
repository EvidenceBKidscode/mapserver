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
	"mapserver/tilerendererjob"
	"mapserver/web"
	"mapserver/mapobject"
	"net/url"
//"fmt"
)

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
	world string
	params params.ParamsType
	window fyne.Window
	status_bar *widget.ProgressBar
	status_text *widget.Label
	link *widget.Hyperlink
}

func (self *Gui) startMapServer() {
	worlddir := filepath.Join(getUserPath(), "worlds", self.world)

	//parse Config
	cfg, err := app.ParseConfig(filepath.Join(worlddir, "mapserver.json"))
	if err != nil {
		panic(err)
	}

	//setup app context
	ctx := app.Setup(self.params, cfg, worlddir)

	//TODO: Rather use event bus
	ctx.SetStatus = func(msg string, progress float64) {
		self.status_text.SetText(msg)
		if progress >= 0 {
			self.status_bar.Show()
			self.status_bar.SetValue(progress)
		} else {
			self.status_bar.Hide()
		}
	}

	//Set up mapobject events
	mapobject.Setup(ctx)

	//run initial rendering
	if ctx.Config.EnableRendering {
		go tilerendererjob.Job(ctx)
	}

	self.status_text.SetText("Lancement du cartographe.")
	self.link.Show()
	//Start http server
	web.Serve(ctx)
}

func (self *Gui) Run(p params.ParamsType) {
	worldpath := filepath.Join(getUserPath(), "worlds")
	self.app = fyneapp.New()
	self.params = p

	// Find available worlds
	var worlds []string
	files, err := ioutil.ReadDir(worldpath)
	if err == nil {
		for _, file := range files {
			if checkWorld(filepath.Join(worldpath, file.Name())) {
				worlds = append(worlds, file.Name())
			}
		}
	}

	// Show main window
	self.window = self.app.NewWindow("Cartographe Kidscode")
	self.window.SetPadded(true)
	self.status_text = widget.NewLabel("Cartographie non lancée.")
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

	if len(worlds) == 0 {
		self.window.SetContent(widget.NewVBox(
			widget.NewLabel("Désolé, aucun monde trouvé."),
			widget.NewButton("Quitter", func() { self.app.Quit() }),
		))
	} else if len(worlds) == 1 {
			self.world = worlds[0]
			go self.startMapServer()
	} else {
		w := self.app.NewWindow("Cartographe Kidscode")
		launch := widget.NewButton("Lancer le cartographe", func() {
			self.window.Show()
			w.Hide()
			go self.startMapServer()
		})
		launch.Disable()

		radio := widget.NewRadio(worlds, func(s string) {
			self.world = s
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

	self.app.Run()
}

func (self *Gui) SetStatus(message string, progress int) {

}
