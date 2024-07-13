package app

import (
	"encoding/json"
	"os"
	"time"
)

type cache_config struct {
	Enabled        bool
	Path           string
	MaxSize        int64 `json:"max-size"`
	Lifetime       int64
	UpdateInterval int64 `json:"update-interval"`
}

type config struct {
	cfg           string
	Listen        string
	BasePath      string `json:"base-path"`
	Cache         cache_config
	Proxy, Nsfw   bool
	DownloadProxy string   `json:"download-proxy"`
	Dirs          []string `json:"dirs-to-memory"`
}

var CFG = config{
	cfg:      "config.json",
	Listen:   "127.0.0.1:3003",
	BasePath: "/",
	Cache: cache_config{
		Enabled:        true,
		Path:           "cache",
		UpdateInterval: 1,
	},
	Dirs:  []string{"html", "css"},
	Proxy: true,
	Nsfw:  true,
}

func ExecuteConfig() {
	go func() {
		for {
			Templates["instances.json"] = string(Download("https://git.macaw.me/skunky/SkunkyArt/raw/branch/master/instances.json").Body)
			time.Sleep(1 * time.Hour)
		}
	}()

	const helpmsg = `SkunkyArt v1.3 [refactoring]
Usage:
	- [-c|--config] - path to config
	- [-h|--help]	- returns this message
Example:
	./skunkyart -c config.json
Copyright lost+skunk, X11. https://git.macaw.me/skunky/skunkyart/src/tag/v1.3`

	a := os.Args
	for n, x := range a {
		switch x {
		case "-c", "--config":
			if len(a) >= 3 {
				CFG.cfg = a[n+1]
			} else {
				exit("Not enought arguments", 1)
			}
		case "-h", "--help":
			exit(helpmsg, 0)
		}
	}

	if CFG.cfg != "" {
		f, err := os.ReadFile(CFG.cfg)
		try_with_exitstatus(err, 1)

		try_with_exitstatus(json.Unmarshal(f, &CFG), 1)
		if CFG.Cache.Enabled && !CFG.Proxy {
			exit("Incompatible settings detected: cannot use caching media content without proxy", 1)
		}

		if CFG.Cache.MaxSize != 0 || CFG.Cache.Lifetime != 0 {
			go InitCacheSystem()
		}
	}
}
