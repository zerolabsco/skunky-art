package app

import (
	"encoding/json"
	"math/rand"
	"strconv"
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
	// Bounded retries: the loop used to be unbounded, and the DeviantArt-error
	// path never incremented attempt, so a single request could spin forever
	// hammering the API (and get this instance's egress IP banned).
	const maxAttempts = 3

	for attempt := 0; attempt < maxAttempts; attempt++ {
		// strconv.Itoa, not string(): string(65) is "A", not "65".
		s, daErr, err := devianter.PerformSearch(strconv.Itoa(rand.Intn(999)), rand.Intn(30), 'a')
		try(err)
		if daErr.RAW != nil {
			continue
		}

		// rand.Intn panics on 0, so an empty result set must be skipped.
		if len(s.Results) == 0 {
			continue
		}

		deviation := &s.Results[rand.Intn(len(s.Results))]
		if deviation.NSFW && !CFG.Nsfw {
			continue
		}

		a.sendMedia(deviation)
		return
	}

	a.Error("Sorry, butt NSFW on this are disabled, and the instance failed to find a random art without NSFW", 500)
}
