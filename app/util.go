package app

import (
	"encoding/base64"
	"io"
	"net/http"
	u "net/url"
	"os"
	"strconv"
	"strings"
	"syscall"
	"text/template"
	"time"

	"git.macaw.me/skunky/devianter"
	"golang.org/x/net/html"
)

/* INTERNAL */
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

func RefreshInstances() {
	for {
		func() {
			defer func() {
				if r := recover(); r != nil {
					recover()
				}
			}()
			Templates["instances.json"] = string(Download("https://git.macaw.me/skunky/SkunkyArt/raw/branch/master/instances.json").Body)
		}()
		time.Sleep(1 * time.Hour)
	}
}

// some crap for frontend
func (s skunkyart) ExecuteTemplate(file string, data any) {
	var buf strings.Builder
	tmp := template.New(file)
	tmp, e := tmp.Parse(Templates[file])
	try(e)
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
		if n+1 < l && !(strs[n+1][0] == '?' || strs[n+1][0] == '&') && !(x[0] == '?' || x[0] == '&') {
			str.WriteString("/")
		}
	}
	return str.String()
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

type Downloaded struct {
	Headers http.Header
	Status  int
	Body    []byte
}

func Download(url string) (d Downloaded) {
	cli := &http.Client{}
	if CFG.DownloadProxy != "" {
		u, e := u.Parse(CFG.DownloadProxy)
		try(e)
		cli.Transport = &http.Transport{Proxy: http.ProxyURL(u)}
	}

	req, e := http.NewRequest("GET", url, nil)
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

// caching
func (s skunkyart) DownloadAndSendMedia(subdomain, path string) {
	var url strings.Builder
	url.WriteString("https://images-wixmp-")
	url.WriteString(subdomain)
	url.WriteString(".wixmp.com/")
	url.WriteString(path)
	url.WriteString("?token=")
	url.WriteString(s.Args.Get("token"))

	if CFG.Cache.Enabled {
		fname := CFG.Cache.Path + "/" + base64.StdEncoding.EncodeToString([]byte(subdomain+path))
		file, e := os.Open(fname)

		if e != nil {
			dwnld := Download(url.String())
			if dwnld.Status == 200 && dwnld.Headers["Content-Type"][0][:5] == "image" {
				try(os.WriteFile(fname, dwnld.Body, 0700))
				s.Writer.Write(dwnld.Body)
			}
		} else {
			file, e := io.ReadAll(file)
			try(e)
			s.Writer.Write(file)
		}
	} else if CFG.Proxy {
		dwnld := Download(url.String())
		s.Writer.Write(dwnld.Body)
	} else {
		s.Writer.WriteHeader(403)
		s.Writer.Write([]byte("Sorry, butt proxy on this instance are disabled."))
	}
}

func InitCacheSystem() {
	c := &CFG.Cache
	os.Mkdir(CFG.Cache.Path, 0700)
	for {
		dir, e := os.Open(c.Path)
		try(e)
		stat, e := dir.Stat()
		try(e)

		dirnames, e := dir.Readdirnames(-1)
		try(e)
		for _, a := range dirnames {
			a = c.Path + "/" + a
			if c.Lifetime != "" {
				now := time.Now().UnixMilli()

				f, _ := os.Stat(a)
				stat := f.Sys().(*syscall.Stat_t)
				time := time.Unix(stat.Ctim.Unix()).UnixMilli()

				if time+lifetimeParsed <= now {
					try(os.RemoveAll(a))
				}
			}
			if c.MaxSize != 0 && stat.Size() > c.MaxSize {
				try(os.RemoveAll(a))
			}
		}

		dir.Close()
		time.Sleep(time.Second * time.Duration(CFG.Cache.UpdateInterval))
	}
}

func CopyTemplatesToMemory() {
	for _, dirname := range CFG.Dirs {
		dir, e := os.ReadDir(dirname)
		tryWithExitStatus(e, 1)

		for _, x := range dir {
			file, e := os.ReadFile(dirname + "/" + x.Name())
			tryWithExitStatus(e, 1)
			Templates[x.Name()] = string(file)
		}
	}
}

/* PARSING HELPERS */
func ParseMedia(media devianter.Media, thumb ...int) string {
	url := devianter.UrlFromMedia(media, thumb...)
	if len(url) != 0 && CFG.Proxy {
		url = url[21:]
		dot := strings.Index(url, ".")

		return UrlBuilder("media", "file", url[:dot], url[dot+11:])
	}
	return url
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
			list.WriteString(`<a href="?p=`)
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
