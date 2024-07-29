package app

import (
	"encoding/json"
	"os"
	"regexp"
	"strconv"
	"time"

	"git.macaw.me/skunky/devianter"
)

type cache_config struct {
	Enabled        bool
	Path           string
	MaxSize        int64 `json:"max-size"`
	Lifetime       string
	UpdateInterval int64 `json:"update-interval"`
}

type config struct {
	cfg           string
	Listen        string
	BasePath      string `json:"base-path"`
	Cache         cache_config
	Proxy, Nsfw   bool
	UserAgent     string   `json:"user-agent"`
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
	Dirs:      []string{"html", "css"},
	UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/127.0.0.0 Safari/537.36",
	Proxy:     true,
	Nsfw:      true,
}

var lifetimeParsed int64

func ExecuteConfig() {
	go func() {
		for {
			func() {
				defer func() {
					if r := recover(); r != nil {
						recover()
					}
				}()
				Templates["instances.json"] = string(Download("https://git.macaw.me/skunky/SkunkyArt/raw/branch/master/instances.json").Body)
			}()
			time.Sleep(1 * time.Hour)
		}
	}()

	const helpmsg = `SkunkyArt v1.3.1 [CSS improvements for mobile and strips on Daily Deviations]
Usage:
	- [-c|--config] - path to config
	- [-h|--help]	- returns this message
Example:
	./skunkyart -c config.json
Copyright lost+skunk, X11. https://git.macaw.me/skunky/skunkyart/src/tag/v1.3.1`

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
		tryWithExitStatus(err, 1)

		tryWithExitStatus(json.Unmarshal(f, &CFG), 1)
		if CFG.Cache.Enabled && !CFG.Proxy {
			exit("Incompatible settings detected: cannot use caching media content without proxy", 1)
		}

		if CFG.Cache.Enabled {
			if CFG.Cache.Lifetime != "" {
				var duration int64
				day := 24 * time.Hour.Milliseconds()
				numstr := regexp.MustCompile("[0-9]+").FindAllString(CFG.Cache.Lifetime, -1)
				num, _ := strconv.Atoi(numstr[len(numstr)-1])

				switch unit := CFG.Cache.Lifetime[len(CFG.Cache.Lifetime)-1:]; unit {
				case "i":
					duration = time.Minute.Milliseconds()
				case "h":
					duration = time.Hour.Milliseconds()
				case "d":
					duration = day
				case "w":
					duration = day * 7
				case "m":
					duration = day * 30
				case "y":
					duration = day * 360
				default:
					exit("Invalid unit specified: "+unit, 1)
				}

				lifetimeParsed = duration * int64(num)
			}
			CFG.Cache.MaxSize /= 1024 ^ 2
			go InitCacheSystem()
		}
		devianter.UserAgent = CFG.UserAgent
	}
}
