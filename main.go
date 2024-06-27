package main

import (
	"skunkyart/app"

	"git.macaw.me/skunky/devianter"
)

func main() {
	err := devianter.UpdateCSRF()
	if err != nil {
		println(err.Error())
	}

	app.Router()
}
