package main

import (
	"skunkyart/app"

	"git.macaw.me/skunky/devianter"
)

func main() {
	devianter.UpdateCSRF()

	app.Router()
}
