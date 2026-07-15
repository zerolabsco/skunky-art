package app

import (
	"encoding/json"
	"os"
	"regexp"
	"skunkyart/static"
	"strconv"
	"time"

	"github.com/zerolabsco/devianter"
)

var Release struct {
	Version     string
	Description string
}

type cache_config struct {
	Enabled        bool
	MemCache       bool `json:"memcache"`
	Path           string
	MaxSize        int64 `json:"max-size"`
	Lifetime       string
	UpdateInterval int64 `json:"update-interval"`
}

type config struct {
	cfg           string
	Listen        string
	URI           string `json:"uri"`
	Cache         cache_config
	Proxy, Nsfw   bool
	UserAgent     string `json:"user-agent"`
	DownloadProxy string `json:"download-proxy"`
	StaticPath    string `json:"static-path"`
}

var CFG = config{
	cfg:    "config.json",
	Listen: "127.0.0.1:3003",
	URI:    "/",
	Cache: cache_config{
		Enabled:        false,
		Path:           "cache",
		UpdateInterval: 1,
	},
	StaticPath: "static",
	UserAgent:  "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/127.0.0.0 Safari/537.36",
	Proxy:      true,
	Nsfw:       true,
}

var lifetimeParsed int64

func ExecuteConfig() {
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
			CFG.Cache.MaxSize *= 1024 ^ 2
			go InitCacheSystem()
		}

		About = instanceAbout{
			Proxy: CFG.Proxy,
			Nsfw:  CFG.Nsfw,
		}

		static.StaticPath = CFG.StaticPath
		devianter.UserAgent = CFG.UserAgent
	}
}
