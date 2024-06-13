package app

import (
	"os"
)

type cache_config struct {
	Enabled            bool
	Path               string
	Max_size, Lifetime int64
}

type config struct {
	cfg              string
	Listen, Base_uri string
	Cache            cache_config
	Proxy, Nsfw      bool
}

var CFG = config{
	cfg:      "config.json",
	Listen:   "127.0.0.1:3003",
	Base_uri: "/",
	Cache: cache_config{
		Enabled: true,
		Path:    "cache",
	},
	Proxy: true,
	Nsfw:  true,
}

func execcfg() {
	a := os.Args
	for num, val := range a {
		switch val {
		case "-conf":
			CFG.cfg = a[num]
		case "-help":
			println(`SkunkyArt v 1.3 [refactoring]
Usage:
	- -conf - path to config
	- -help this message
Example:
	./skunkyart -conf config.json
Copyright lost+skunk, X11. https://git.macaw.me/skunky/skunkyart/src/tag/v1.3`)
		}
	}
}
