package app

import (
	"math/rand"
	"strings"

	"git.macaw.me/skunky/devianter"
)

type API struct {
  skunkyartLink *skunkyart
}

func (a API) sendMedia(d *devianter.Deviation) {
  mediaUrl, name := devianter.UrlFromMedia(d.Media)
  
  var filename strings.Builder
	filename.WriteString(`filename="`)
  filename.WriteString(name)
	filename.WriteString(`"`)
	a.skunkyartLink.Writer.Header().Add("Content-Disposition", filename.String())
	
  if len(mediaUrl) != 0 {
    mediaUrl = mediaUrl[21:]
		dot := strings.Index(mediaUrl, ".")
    a.skunkyartLink.DownloadAndSendMedia(mediaUrl[:dot], mediaUrl[dot+11:])
  }
}

func (a API) Random() {
  s, err := devianter.PerformSearch(string(rand.Intn(999)), rand.Intn(30), 'a')
  try(err) 
  a.sendMedia(&s.Results[rand.Intn(len(s.Results))])
}
