package app

import (
	"encoding/json"
	"errors"
	"os"
)

type cache_config struct {
	Enabled  bool
	Path     string
	MaxSize  int64 `json:"max-size"`
	Lifetime int64
}

type config struct {
	cfg         string
	Listen      string
	BasePath    string `json:"base-path"`
	Cache       cache_config
	Proxy, Nsfw bool
	WixmpProxy  string `json:"wixmp-proxy"`
}

var CFG = config{
	cfg:      "config.json",
	Listen:   "127.0.0.1:3003",
	BasePath: "/",
	Cache: cache_config{
		Enabled: true,
		Path:    "cache",
	},
	Proxy: true,
	Nsfw:  true,
}

func ExecuteConfig() {
	try := func(err error, exitcode int) {
		if err != nil {
			println(err.Error())
			os.Exit(exitcode)
		}
	}

	a := os.Args
	if l := len(a); l > 1 {
		switch a[1] {
		case "-c", "--config":
			if l >= 3 {
				CFG.cfg = a[2]
			} else {
				try(errors.New("Not enought arguments"), 1)
			}
		case "-h", "--help":
			try(errors.New(`SkunkyArt v1.3 [refactoring]
Usage:
	- [-c|--config] - path to config
	- [-h|--help]	- returns this message
Example:
	./skunkyart -c config.json
Copyright lost+skunk, X11. https://git.macaw.me/skunky/skunkyart/src/tag/v1.3`), 0)
		default:
			try(errors.New("Unreconginzed argument: "+a[1]), 1)
		}
		if CFG.cfg != "" {
			f, err := os.ReadFile(CFG.cfg)
			try(err, 1)

			try(json.Unmarshal(f, &CFG), 1)
			if CFG.Cache.Enabled && !CFG.Proxy {
				try(errors.New("Incompatible settings detected: cannot use caching media content without proxy"), 1)
			}

			if CFG.Cache.MaxSize != 0 || CFG.Cache.Lifetime != 0 {
				go InitCacheSystem()
			}
		}
	}
}
