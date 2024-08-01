package main

import (
	"skunkyart/app"
	"time"

	"git.macaw.me/skunky/devianter"
)

func main() {
	go app.RefreshInstances()

	app.ExecuteCommandLineArguments()
	app.ExecuteConfig()
	app.CopyTemplatesToMemory()

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
