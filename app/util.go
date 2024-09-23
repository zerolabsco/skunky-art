package app

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"skunkyart/static"
	"strconv"
	"strings"
	"text/template"
	"time"

	"git.macaw.me/skunky/devianter"
	"golang.org/x/net/html"
)

/* INTERNAL */
var wr = io.WriteString

func exit(msg string, code int) {
	println(msg)
	os.Exit(code)
}
func try(e error) {
	if e != nil {
		println(e.Error())
	}
}
func tryWithExitStatus(err error, code int) {
	if err != nil {
		exit(err.Error(), code)
	}
}

func restore() {
	if r := recover(); r != nil {
		recover()
	}
}

var instances []byte
var About instanceAbout

func RefreshInstances() {
	for {
		func() {
			defer restore()
			instances = Download("https://git.macaw.me/skunky/SkunkyArt/raw/branch/master/instances.json").Body
			try(json.Unmarshal(instances, &About))
		}()
		time.Sleep(1 * time.Hour)
	}
}

// some crap for frontend
type instanceAbout struct {
	Proxy     bool
	Nsfw      bool
	Instances []settings
}

type skunkyart struct {
	Writer http.ResponseWriter

	Args url.Values
	Page int
	Type rune
	Atom bool

	BasePath, Endpoint string
	Query, QueryRaw    string

	API     API
	Version string

	Templates struct {
		About instanceAbout

		SomeList  string
		DDStrips  string
		Deviation struct {
			Post       devianter.Post
			Related    string
			StringTime string
			Tags       string
			Comments   string
		}

		GroupUser struct {
			GR           devianter.GRuser
			Admins       string
			Group        bool
			CreationDate string

			About struct {
				A devianter.About

				DescriptionFormatted string
				Interests, Social    string
				Comments             string
				BG                   string
				BGMeta               devianter.Deviation
			}

			Gallery struct {
				Folders string
				Pages   int
				List    string
			}
		}
		Search struct {
			Content devianter.Search
			List    string
		}
	}
}

func (s skunkyart) ExecuteTemplate(file, dir string, data any) {
	var buf strings.Builder
	tmp := template.New(file)
	tmp, err := tmp.ParseFS(static.Templates, dir+"/*")
	if err != nil {
		s.Writer.WriteHeader(500)
		wr(s.Writer, err.Error())
		return
	}
	try(tmp.Execute(&buf, &data))
	wr(s.Writer, buf.String())
}

func UrlBuilder(strs ...string) string {
	var str strings.Builder
	l := len(strs)
	str.WriteString(Host)
	str.WriteString(CFG.URI)
	for n, x := range strs {
		str.WriteString(x)
		if n := n + 1; n < l && len(strs[n]) != 0 && !(strs[n][0] == '?' || strs[n][0] == '&') && !(x[0] == '?' || x[0] == '&') {
			str.WriteString("/")
		}
	}
	return str.String()
}

func (s skunkyart) Error(dAerr devianter.Error) {
	s.Writer.WriteHeader(502)

	var msg strings.Builder
	msg.WriteString(`<html><link rel="stylesheet" href="`)
	msg.WriteString(UrlBuilder("stylesheet"))
	msg.WriteString(`" /><h3>DeviantArt error — '`)
	msg.WriteString(dAerr.Error)
	msg.WriteString("'</h3></html>")

	wr(s.Writer, msg.String())
}

func (s skunkyart) ReturnHTTPError(status int) {
	s.Writer.WriteHeader(status)

	var msg strings.Builder
	msg.WriteString(`<html><link rel="stylesheet" href="`)
	msg.WriteString(UrlBuilder("stylesheet"))
	msg.WriteString(`" /><h1>`)
	msg.WriteString(strconv.Itoa(status))
	msg.WriteString(" - ")
	msg.WriteString(http.StatusText(status))
	msg.WriteString("</h1></html>")

	wr(s.Writer, msg.String())
}

func (s skunkyart) SetFilename(name string) {
	var filename strings.Builder
	filename.WriteString(`filename="`)
	filename.WriteString(name)
	filename.WriteString(`"`)
	s.Writer.Header().Add("Content-Disposition", filename.String())
}

type Downloaded struct {
	Headers http.Header
	Status  int
	Body    []byte
}

func Download(urlString string) (d Downloaded) {
	cli := &http.Client{}
	if CFG.DownloadProxy != "" {
		u, e := url.Parse(CFG.DownloadProxy)
		try(e)
		cli.Transport = &http.Transport{Proxy: http.ProxyURL(u)}
	}

	req, e := http.NewRequest("GET", urlString, nil)
	try(e)
	req.Header.Set("User-Agent", CFG.UserAgent)

	resp, e := cli.Do(req)
	try(e)
	defer resp.Body.Close()
	b, e := io.ReadAll(resp.Body)
	try(e)

	d.Body = b
	d.Status = resp.StatusCode
	d.Headers = resp.Header
	return
}

/* PARSING HELPERS */
func ParseMedia(media devianter.Media, thumb ...int) string {
	mediaUrl, filename := devianter.UrlFromMedia(media, thumb...)
	if len(mediaUrl) != 0 && CFG.Proxy {
		mediaUrl = mediaUrl[21:]
		dot := strings.Index(mediaUrl, ".")
		if filename == "" {
			filename = "image.gif"
		}
		return UrlBuilder("media", "file", mediaUrl[:dot], mediaUrl[dot+11:], "&filename=", filename)
	} else if !CFG.Proxy {
		return mediaUrl
	}
	return ""
}

func ConvertDeviantArtUrlToSkunkyArt(url string) (output string) {
	if len(url) > 32 && url[27:32] != "stash" {
		url = url[27:]
		firstshash := strings.Index(url, "/")
		lastshash := firstshash + strings.Index(url[firstshash+1:], "/")
		if lastshash != -1 {
			output = UrlBuilder("post", url[:firstshash], url[lastshash+2:])
		}
	}
	return
}

func BuildUserPlate(name string) string {
	var htm strings.Builder
	htm.WriteString(`<div class="user-plate"><img src="`)
	htm.WriteString(UrlBuilder("media", "emojitar", name, "?type=a"))
	htm.WriteString(`"><a href="`)
	htm.WriteString(UrlBuilder("group_user", "?type=about&q=", name))
	htm.WriteString(`">`)
	htm.WriteString(name)
	htm.WriteString(`</a></div>`)
	return htm.String()
}

func GetValueOfTag(t *html.Tokenizer) string {
	for tt := t.Next(); ; {
		if tt == html.TextToken {
			return string(t.Text())
		} else {
			return ""
		}
	}
}

// навигация по страницам
type DeviationList struct {
	Pages int
	More  bool
}

// FIXME: на некоротрых артах первая страница может вызывать полное отсутствие панели навигации.
func (s skunkyart) NavBase(c DeviationList) string {
	var list strings.Builder

	list.WriteString("<br>")
	prevrev := func(msg string, page int, onpage bool) {
		if !onpage {
			list.WriteString(`<a href="`)
			list.WriteString(Path)
			list.WriteString(`?p=`)
			list.WriteString(strconv.Itoa(page))
			if s.Type != 0 {
				list.WriteString("&type=")
				list.WriteRune(s.Type)
			}
			if s.Query != "" {
				list.WriteString("&q=")
				list.WriteString(s.Query)
			}
			if f := s.Args.Get("folder"); f != "" {
				list.WriteString("&folder=")
				list.WriteString(f)
			}
			list.WriteString(`">`)
			list.WriteString(msg)
			list.WriteString("</a> ")
		} else {
			list.WriteString(strconv.Itoa(page))
			list.WriteString(" ")
		}
	}

	p := s.Page

	if p > 1 {
		prevrev("<= Prev |", p-1, false)
	} else {
		p = 1
	}

	for i, x := p-6, 0; (i <= c.Pages && i <= p+6) && x < 12; i++ {
		if i > 0 {
			var onPage bool
			if i == p {
				onPage = true
			}

			prevrev(strconv.Itoa(i), i, onPage)
			x++
		}
	}

	if c.More {
		prevrev("| Next =>", p+1, false)
	}

	return list.String()
}
