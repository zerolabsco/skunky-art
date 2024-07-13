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
func try_with_exitstatus(err error, code int) {
	if err != nil {
		exit(err.Error(), code)
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
	str.WriteString(CFG.BasePath)
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
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:123.0) Gecko/20100101 Firefox/123.0.0")

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
		os.Mkdir(CFG.Cache.Path, 0700)
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
	for {
		dir, e := os.Open(c.Path)
		try(e)
		stat, e := dir.Stat()
		try(e)

		dirnames, e := dir.Readdirnames(-1)
		try(e)
		for _, a := range dirnames {
			a = c.Path + "/" + a
			if c.Lifetime != 0 {
				now := time.Now().UnixMilli()

				f, _ := os.Stat(a)
				stat := f.Sys().(*syscall.Stat_t)
				time := time.Unix(stat.Ctim.Unix()).UnixMilli()

				if time+c.Lifetime <= now {
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
		try_with_exitstatus(e, 1)

		for _, x := range dir {
			n := dirname + "/" + x.Name()
			file, e := os.ReadFile(n)
			try_with_exitstatus(e, 1)
			Templates[n] = string(file)
		}
	}
}

/* PARSING HELPERS */
func ParseMedia(media devianter.Media) string {
	url := devianter.UrlFromMedia(media)
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
		toart := strings.Index(url, "/art/")
		if toart != -1 {
			output = UrlBuilder("post", url[:toart], url[toart+5:])
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
		switch tt {
		default:
			return ""
		case html.TextToken:
			return string(t.Text())
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
	// TODO: сделать понятнее
	// навигация по страницам
	var list strings.Builder
	list.WriteString("<br>")
	p := s.Page

	// функция для генерации ссылок
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

	// вперёд-назад
	if p > 1 {
		prevrev("<= Prev |", p-1, false)
	} else {
		p = 1
	}

	if c.Pages > 0 {
		// назад
		for x := p - 6; x < p && x > 0; x++ {
			prevrev(strconv.Itoa(x), x, false)
		}

		// вперёд
		for x := p; x <= p+6 && c.Pages > p+6; x++ {
			if x == p {
				prevrev("", x, true)
				x++
			}

			if x > p {
				prevrev(strconv.Itoa(x), x, false)
			}
		}
	}

	// вперёд-назад
	if c.More {
		prevrev("| Next =>", p+1, false)
	}

	return list.String()
}
