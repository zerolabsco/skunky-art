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
	go app.RefreshInstances()

	app.ExecuteCommandLineArguments()
	app.ExecuteConfig()
	static.CopyTemplatesToMemory()

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
