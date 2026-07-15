package main

import (
	"skunkyart/app"
	"skunkyart/static"
	"time"

	"github.com/zerolabsco/devianter"
)

func main() {
	app.Release.Version = "1.3.2"
	app.Release.Description = "Two API endpoints and template embedding into binary"

	app.ExecuteCommandLineArguments()
	app.ExecuteConfig()
	static.CopyTemplatesToMemory()

	// Rate/concurrency-limit + time-out outbound DeviantArt requests so bot floods
	// can't exhaust the process or get our egress IP banned by CloudFront/WAF.
	app.InstallDAThrottle()

	// Only once the config is loaded and the throttle installed: this fetches over
	// the network, so starting it earlier both raced ExecuteConfig's writes to CFG
	// and let the request escape the throttle and the configured User-Agent.
	go app.RefreshInstances()

	go func() {
		for {
			err := devianter.UpdateCSRF()
			if err != nil {
				println(err.Error())
			}
			time.Sleep(12 * time.Hour)
		}
	}()

	app.Router()
}
