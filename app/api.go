package app

import (
	"encoding/json"
	"math/rand"
	"strings"

	"github.com/zerolabsco/devianter"
)

type API struct {
	main *skunkyart
}

type info struct {
	Version  string         `json:"version"`
	Settings settingsParams `json:"settings"`
}

func (a API) Info() {
	json, err := json.Marshal(info{
		Version: a.main.Version,
		Settings: settingsParams{
			Nsfw:  CFG.Nsfw,
			Proxy: CFG.Proxy,
		},
	})
	try(err)
	a.main.Writer.Write(json)
}

func (a API) Error(description string, status int) {
	a.main.Writer.WriteHeader(status)
	var response strings.Builder
	response.WriteString(`{"error":"`)
	response.WriteString(description)
	response.WriteString(`"}`)
	wr(a.main.Writer, response.String())
}

func (a API) sendMedia(d *devianter.Deviation) {
	mediaUrl, name := devianter.UrlFromMedia(d.Media)
	a.main.SetFilename(name)
	if len(mediaUrl) != 0 {
		return
	}

	if CFG.Proxy {
		mediaUrl = mediaUrl[21:]
		dot := strings.Index(mediaUrl, ".")
		a.main.Writer.Header().Del("Content-Type")
		a.main.DownloadAndSendMedia(mediaUrl[:dot], mediaUrl[dot+11:])
	} else {
		a.main.Writer.Header().Add("Location", mediaUrl)
		a.main.Writer.WriteHeader(302)
	}
}

// TODO: сделать фильтры
func (a API) Random() {
	for attempt := 1; ; {
		if attempt > 3 {
			a.Error("Sorry, butt NSFW on this are disabled, and the instance failed to find a random art without NSFW", 500)
		}

		s, daErr, err := devianter.PerformSearch(string(rand.Intn(999)), rand.Intn(30), 'a')
		try(err)
		if daErr.RAW != nil {
			continue
		}

		deviation := &s.Results[rand.Intn(len(s.Results))]

		if deviation.NSFW && !CFG.Nsfw {
			attempt++
			continue
		}

		a.sendMedia(deviation)
		return
	}
}
