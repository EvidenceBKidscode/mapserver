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
//	"fmt"
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
	basedir string
	app fyne.App
	world string
	params params.ParamsType
	window fyne.Window
}


func (self *Gui) startMapServer() {
	self.window.SetContent(widget.NewVBox(
		widget.NewLabel("Cartographie active sur le monde \"" + self.world + "\"."),
		widget.NewButton("Arrêter", func() { self.app.Quit() }),
	))
	app.WorldDir = filepath.Join(self.basedir, "..", "worlds", self.world)

	//parse Config
	cfg, err := app.ParseConfig()
	if err != nil {
		panic(err)
	}

	//setup app context
	ctx := app.Setup(self.params, cfg)

	//Set up mapobject events
	mapobject.Setup(ctx)

	//run initial rendering
	if ctx.Config.EnableRendering {
		go tilerendererjob.Job(ctx)
	}

	//Start http server
	web.Serve(ctx)
}

func (self *Gui) Run(p params.ParamsType) {
	self.basedir = filepath.Dir(os.Args[0])
	self.app = fyneapp.New()
	self.params = p

	// Find available worlds
	var worlds []string
	files, err := ioutil.ReadDir(filepath.Join(self.basedir, "..", "worlds"))
	if err == nil {
		for _, file := range files {
			if checkWorld(filepath.Join(self.basedir, "..", "worlds", file.Name())) {
				worlds = append(worlds, file.Name())
			}
		}
	}

	// Show main window
	self.window = self.app.NewWindow("Cartographe Kidscode")
	self.window.SetPadded(true)
	self.window.SetContent(widget.NewVBox(
		widget.NewLabel("Cartographie non lancée."),
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
