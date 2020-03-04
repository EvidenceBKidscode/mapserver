package app

import (
	"encoding/json"
	"io/ioutil"
	"mapserver/layer"
	"os"
	"runtime"
	"sync"
)

type Config struct {
	ConfigVersion          int                     `json:"configversion"`
	Port                   int                     `json:"port"`
	EnablePrometheus       bool                    `json:"enableprometheus"`
	EnableRendering        bool                    `json:"enablerendering"`
	EnableSearch           bool                    `json:"enablesearch"`
	EnableInitialRendering bool                    `json:"enableinitialrendering"`
	EnableTransparency     bool                    `json:"enabletransparency"`
	EnableMediaRepository  bool                    `json:"enablemediarepository"`
	Webdev                 bool                    `json:"webdev"`
	WebApi                 *WebApiConfig           `json:"webapi"`
	Layers                 []*layer.Layer          `json:"layers"`
	InitialFetchLimit      int                     `json:"initialfetchlimit"`
	IncrementalFetchLimit  int                     `json:"incrementalfetchlimit"`
	RenderingJobs          int                     `json:"renderingjobs"`
	RenderingQueue         int                     `json:"renderingqueue"`
	MapObjects             *MapObjectConfig        `json:"mapobjects"`
	MapBlockAccessorCfg    *MapBlockAccessorConfig `json:"mapblockaccessor"`
	DefaultOverlays        []string                `json:"defaultoverlays"`
	ConfigFilePath         string
}

type MapBlockAccessorConfig struct {
	Expiretime string `json:"expiretime"`
	Purgetime  string `json:"purgetime"`
	MaxItems   int    `json:"maxitems"`
}

type MapObjectConfig struct {
	Areas              bool `json:"areas"`
	Bones              bool `json:"bones"`
	Protector          bool `json:"protector"`
	XPProtector        bool `json:"xpprotector"`
	PrivProtector      bool `json:"privprotector"`
	TechnicQuarry      bool `json:"technic_quarry"`
	TechnicSwitch      bool `json:"technic_switch"`
	TechnicAnchor      bool `json:"technic_anchor"`
	TechnicReactor     bool `json:"technic_reactor"`
	LuaController      bool `json:"luacontroller"`
	Digiterms          bool `json:"digiterms"`
	Digilines          bool `json:"digilines"`
	Travelnet          bool `json:"travelnet"`
	MapserverPlayer    bool `json:"mapserver_player"`
	MapserverPOI       bool `json:"mapserver_poi"`
	MapserverLabel     bool `json:"mapserver_label"`
	MapserverTrainline bool `json:"mapserver_trainline"`
	MapserverBorder    bool `json:"mapserver_border"`
	TileServerLegacy   bool `json:"tileserverlegacy"`
	Mission            bool `json:"mission"`
	Jumpdrive          bool `json:"jumpdrive"`
	Smartshop          bool `json:"smartshop"`
	Fancyvend          bool `json:"fancyvend"`
	ATM                bool `json:"atm"`
	Train              bool `json:"train"`
	TrainSignal        bool `json:"trainsignal"`
	Minecart           bool `json:"minecart"`
	Locator            bool `json:"locator"`
	Symbols            bool `json:"symbols"`
}

type WebApiConfig struct {
	//mapblock debugging
	EnableMapblock bool `json:"enablemapblock"`

	//mod http bridge secret
	SecretKey string `json:"secretkey"`
}

var lock sync.Mutex

func (cfg *Config) Save() error {
	lock.Lock()
	defer lock.Unlock()

	f, err := os.Create(cfg.ConfigFilePath)
	if err != nil {
		return err
	}

	defer f.Close()

	str, err := json.MarshalIndent(cfg, "", "	")
	if err != nil {
		return err
	}

	f.Write(str)

	return nil
}

func ParseConfig(configfilepath string) (*Config, error) {
	webapi := WebApiConfig{
		EnableMapblock: false,
		SecretKey:      RandStringRunes(16),
	}

	layers := []*layer.Layer{
		&layer.Layer{
			Id:   0,
			Name: "Ground",
			From: -1,
			To:   100,
		},
	}

	mapobjs := MapObjectConfig{
		Areas:              false,
		Bones:              false,
		Protector:          false,
		XPProtector:        false,
		PrivProtector:      false,
		TechnicQuarry:      false,
		TechnicSwitch:      false,
		TechnicAnchor:      false,
		TechnicReactor:     false,
		LuaController:      false,
		Digiterms:          false,
		Digilines:          false,
		Travelnet:          false,
		MapserverPlayer:    true,
		MapserverPOI:       false,
		MapserverLabel:     false,
		MapserverTrainline: false,
		MapserverBorder:    false,
		TileServerLegacy:   false,
		Mission:            false,
		Jumpdrive:          false,
		Smartshop:          false,
		Fancyvend:          false,
		ATM:                false,
		Train:              false,
		TrainSignal:        false,
		Minecart:           false,
		Locator:            false,
		Symbols:            true,
	}

	mapblockaccessor := MapBlockAccessorConfig{
		Expiretime: "15s",
		Purgetime:  "30s",
		MaxItems:   500,
	}

	defaultoverlays := []string{
		"mapserver_poi",
		"mapserver_label",
		"mapserver_player",
	}

	cfg := Config{
		ConfigVersion:          1,
		Port:                   8080,
		EnableRendering:        true,
		EnablePrometheus:       true,
		EnableSearch:           false,
		EnableInitialRendering: true,
		EnableTransparency:     false,
		EnableMediaRepository:  false,
		Webdev:                 false,
		WebApi:                 &webapi,
		Layers:                 layers,
		InitialFetchLimit:      2000,
		IncrementalFetchLimit:  500,
		RenderingJobs:          runtime.NumCPU(),
		RenderingQueue:         100,
		MapObjects:             &mapobjs,
		MapBlockAccessorCfg:    &mapblockaccessor,
		DefaultOverlays:        defaultoverlays,
		ConfigFilePath:         configfilepath,
	}

	info, err := os.Stat(cfg.ConfigFilePath)
	if info != nil && err == nil {
		data, err := ioutil.ReadFile(cfg.ConfigFilePath)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(data, &cfg)
		if err != nil {
			return nil, err
		}

		//write back config with all values
		err = cfg.Save()
		if err != nil {
			panic(err)
		}
	}

	//write back config with all values
	err = cfg.Save()
	if err != nil {
		panic(err)
	}

	return &cfg, nil
}
