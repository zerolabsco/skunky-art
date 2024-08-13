package main

import (
	"skunkyart/app"
	"skunkyart/static"
	"time"

	"git.macaw.me/skunky/devianter"
)

func main() {
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
