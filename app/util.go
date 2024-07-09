package app

import (
	"encoding/base64"
	"encoding/json"
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

// парсинг темплейтов
func (s skunkyart) ExecuteTemplate(file string, data any) {
	var buf strings.Builder
	tmp := template.New(file)
	tmp, e := tmp.Parse(Templates[file])
	err(e)
	err(tmp.Execute(&buf, &data))
	wr(s.Writer, buf.String())
}

func UrlBuilder(strs ...string) string {
	var str strings.Builder
	l := len(strs)
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
	msg.WriteString(UrlBuilder("gui", "css", "skunky.css"))
	msg.WriteString(`" /><h1>`)
	msg.WriteString(strconv.Itoa(status))
	msg.WriteString(" - ")
	msg.WriteString(http.StatusText(status))
	msg.WriteString("</h1></html>")

	wr(s.Writer, msg.String())
}

func (s skunkyart) ConvertDeviantArtUrlToSkunkyArt(url string) (output string) {
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

type text struct {
	TXT  string
	from int
	to   int
}

func tagval(t *html.Tokenizer) string {
	for tt := t.Next(); ; {
		switch tt {
		default:
			return ""
		case html.TextToken:
			return string(t.Text())
		}
	}
}

func ParseDescription(dscr devianter.Text) string {
	var parseddescription strings.Builder
	TagBuilder := func(tag string, content string) string {
		if tag != "" {
			var htm strings.Builder
			htm.WriteString("<")
			htm.WriteString(tag)
			htm.WriteString(">")

			htm.WriteString(content)

			htm.WriteString("</")
			htm.WriteString(tag)
			htm.WriteString(">")
			return htm.String()
		}
		return content
	}
	DeleteSpywareFromUrl := func(url string) string {
		if len(url) > 42 && url[:42] == "https://www.deviantart.com/users/outgoing?" {
			url = url[42:]
		}
		return url
	}

	if description, dl := dscr.Html.Markup, len(dscr.Html.Markup); dl != 0 &&
		description[0] == '{' &&
		description[dl-1] == '}' {
		var descr struct {
			Blocks []struct {
				Text, Type        string
				InlineStyleRanges []struct {
					Offset, Length int
					Style          string
				}
				EntityRanges []struct {
					Offset, Length int
					Key            int
				}
				Data struct {
					TextAlignment string
				}
			}
			EntityMap map[string]struct {
				Type string
				Data struct {
					Url    string
					Config struct {
						Aligment string
						Width    int
					}
					Data devianter.Deviation
				}
			}
		}
		e := json.Unmarshal([]byte(description), &descr)
		err(e)

		entities := make(map[int]devianter.Deviation)
		urls := make(map[int]string)
		for n, x := range descr.EntityMap {
			num, _ := strconv.Atoi(n)
			if x.Data.Url != "" {
				urls[num] = DeleteSpywareFromUrl(x.Data.Url)
			}
			entities[num] = x.Data.Data
		}

		for _, x := range descr.Blocks {
			ranges := make(map[int]text)

			for i, rngs := range x.InlineStyleRanges {
				var tag string

				switch rngs.Style {
				case "BOLD":
					tag = "b"
				case "UNDERLINE":
					tag = "u"
				case "ITALIC":
					tag = "i"
				}

				fromto := rngs.Offset + rngs.Length
				ranges[i] = text{
					TXT:  TagBuilder(tag, x.Text[rngs.Offset:fromto]),
					from: rngs.Offset,
					to:   fromto,
				}
			}

			switch x.Type {
			case "atomic":
				d := entities[x.EntityRanges[0].Key]
				parseddescription.WriteString(`<img width="50%" src="`)
				parseddescription.WriteString(ParseMedia(d.Media))
				parseddescription.WriteString(`" title="`)
				parseddescription.WriteString(d.Author.Username)
				parseddescription.WriteString(" - ")
				parseddescription.WriteString(d.Title)
				parseddescription.WriteString(`">`)
			case "unstyled":
				if len(ranges) != 0 {
					for _, r := range ranges {
						var tag string
						switch x.Type {
						case "header-two":
							tag = "h2"
						}

						parseddescription.WriteString(x.Text[:r.from])
						if len(urls) != 0 && len(x.EntityRanges) != 0 {
							ra := &x.EntityRanges[0]

							parseddescription.WriteString(`<a target="_blank" href="`)
							parseddescription.WriteString(urls[ra.Key])
							parseddescription.WriteString(`">`)
							parseddescription.WriteString(r.TXT)
							parseddescription.WriteString(`</a>`)
						} else {
							parseddescription.WriteString(r.TXT)
						}
						parseddescription.WriteString(TagBuilder(tag, x.Text[r.to:]))
					}
				} else {
					parseddescription.WriteString(x.Text)
				}
			}
			parseddescription.WriteString("<br>")
		}
	} else if dl != 0 {
		for tt := html.NewTokenizer(strings.NewReader(dscr.Html.Markup)); ; {
			switch tt.Next() {
			case html.ErrorToken:
				return parseddescription.String()
			case html.StartTagToken, html.EndTagToken, html.SelfClosingTagToken:
				token := tt.Token()
				switch token.Data {
				case "a":
					for _, a := range token.Attr {
						if a.Key == "href" {
							url := DeleteSpywareFromUrl(a.Val)
							parseddescription.WriteString(`<a target="_blank" href="`)
							parseddescription.WriteString(url)
							parseddescription.WriteString(`">`)
							parseddescription.WriteString(tagval(tt))
							parseddescription.WriteString("</a> ")
						}
					}
				case "img":
					var uri, title string
					for b, a := range token.Attr {
						switch a.Key {
						case "src":
							if len(a.Val) > 9 && a.Val[8:9] == "e" {
								uri = UrlBuilder("media", "emojitar", a.Val[37:len(a.Val)-4], "?type=e")
							}
						case "title":
							title = a.Val
						}
						if title != "" {
							for x := -1; x < b; x++ {
								parseddescription.WriteString(`<img src="`)
								parseddescription.WriteString(uri)
								parseddescription.WriteString(`" title="`)
								parseddescription.WriteString(title)
								parseddescription.WriteString(`">`)
							}
						}
					}
				case "br", "li", "ul", "p", "b":
					parseddescription.WriteString(token.String())
				case "div":
					parseddescription.WriteString("<p> ")
				}
			case html.TextToken:
				parseddescription.Write(tt.Text())
			}
		}
	}

	return parseddescription.String()
}

// навигация по страницам
type dlist struct {
	Pages int
	More  bool
}

// FIXME: на некоротрых артах первая страница может вызывать полное отсутствие панели навигации.
func (s skunkyart) NavBase(c dlist) string {
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
		for x := p; x <= p+6; x++ {
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
	if p != 417 || c.More {
		prevrev("| Next =>", p+1, false)
	}

	return list.String()
}

func (s skunkyart) DeviationList(devs []devianter.Deviation, content ...dlist) string {
	var list strings.Builder
	if s.Atom && s.Page > 1 {
		s.ReturnHTTPError(400)
		return ""
	} else if s.Atom {
		list.WriteString(`<?xml version="1.0" encoding="UTF-8"?><feed xmlns:media="http://search.yahoo.com/mrss/" xmlns="http://www.w3.org/2005/Atom">`)
		list.WriteString(`<title>SkunkyArt</title>`)
		// list.WriteString(`<link rel="alternate" href="HOMEPAGE_URL"/><link href="FEED_URL" rel="self"/>`)
	} else {
		list.WriteString(`<div class="content">`)
	}
	for _, data := range devs {
		if !(data.NSFW && !CFG.Nsfw) {
			url := ParseMedia(data.Media)
			if s.Atom {
				id := strconv.Itoa(data.ID)
				list.WriteString(`<entry><author><name>`)
				list.WriteString(data.Author.Username)
				list.WriteString(`</name></author><title>`)
				list.WriteString(data.Title)
				list.WriteString(`</title><link rel="alternate" type="text/html" href="`)
				list.WriteString(UrlBuilder("post", data.Author.Username, "atom-"+id))
				list.WriteString(`"/><id>`)
				list.WriteString(id)
				list.WriteString(`</id><published>`)
				list.WriteString(data.PublishedTime.UTC().Format("Mon, 02 Jan 2006 15:04:05 -0700"))
				list.WriteString(`</published>`)
				list.WriteString(`<media:group><media:title>`)
				list.WriteString(data.Title)
				list.WriteString(`</media:title><media:thumbinal url="`)
				list.WriteString(url)
				list.WriteString(`"/></media:group><content type="xhtml"><div xmlns="http://www.w3.org/1999/xhtml"><a href="`)
				list.WriteString(data.Url)
				list.WriteString(`"><img src="`)
				list.WriteString(url)
				list.WriteString(`"/></a><p>`)
				list.WriteString(ParseDescription(data.TextContent))
				list.WriteString(`</p></div></content></entry>`)
			} else {
				list.WriteString(`<div class="block">`)
				if url != "" {
					list.WriteString(`<a title="open/download" href="`)
					list.WriteString(url)
					list.WriteString(`"><img loading="lazy" src="`)
					list.WriteString(url)
					list.WriteString(`" width="15%"></a>`)
				} else {
					list.WriteString(`<h1>[ TEXT ]</h1>`)
				}
				list.WriteString(`<br><a href="`)
				list.WriteString(s.ConvertDeviantArtUrlToSkunkyArt(data.Url))
				list.WriteString(`">`)
				list.WriteString(data.Author.Username)
				list.WriteString(" - ")
				list.WriteString(data.Title)

				// шильдики нсфв, аи и ежедневного поста
				if data.NSFW {
					list.WriteString(` [<span class="nsfw">NSFW</span>]`)
				}
				if data.AI {
					list.WriteString(" [🤖]")
				}
				if data.DD {
					list.WriteString(` [<span class="dd">DD</span>]`)
				}

				list.WriteString("</a></div>")
			}
		}
	}

	if s.Atom {
		list.WriteString("</feed>")
		s.Writer.Write([]byte(list.String()))
		return ""
	} else {
		list.WriteString("</div>")
		if content != nil {
			list.WriteString(s.NavBase(content[0]))
		}
	}

	return list.String()
}

func (s skunkyart) ParseComments(c devianter.Comments) string {
	var cmmts strings.Builder
	replied := make(map[int]string)

	cmmts.WriteString("<details><summary>Comments: <b>")
	cmmts.WriteString(strconv.Itoa(c.Total))
	cmmts.WriteString("</b></summary>")
	for _, x := range c.Thread {
		replied[x.ID] = x.User.Username
		cmmts.WriteString(`<div class="msg`)
		if x.Parent > 0 {
			cmmts.WriteString(` reply`)
		}
		cmmts.WriteString(`"><p id="`)
		cmmts.WriteString(strconv.Itoa(x.ID))
		cmmts.WriteString(`"><img src="`)
		cmmts.WriteString(UrlBuilder("media", "emojitar", x.User.Username, "?type=a"))
		cmmts.WriteString(`" width="30px" height="30px"><a href="`)
		cmmts.WriteString(UrlBuilder("group_user", "?q=", x.User.Username, "&type=a"))
		cmmts.WriteString(`"><b`)
		cmmts.WriteString(` class="`)
		if x.User.Banned {
			cmmts.WriteString(`banned`)
		}
		if x.Author {
			cmmts.WriteString(`author`)
		}
		cmmts.WriteString(`">`)
		cmmts.WriteString(x.User.Username)
		cmmts.WriteString("</b></a> ")

		if x.Parent > 0 {
			cmmts.WriteString(` In reply to <a href="#`)
			cmmts.WriteString(strconv.Itoa(x.Parent))
			cmmts.WriteString(`">`)
			if replied[x.Parent] == "" {
				cmmts.WriteString("???")
			} else {
				cmmts.WriteString(replied[x.Parent])
			}
			cmmts.WriteString("</a>")
		}
		cmmts.WriteString(" [")
		cmmts.WriteString(x.Posted.UTC().String())
		cmmts.WriteString("]<p>")

		cmmts.WriteString(ParseDescription(x.TextContent))
		cmmts.WriteString("<p>👍: ")
		cmmts.WriteString(strconv.Itoa(x.Likes))
		cmmts.WriteString(" ⏩: ")
		cmmts.WriteString(strconv.Itoa(x.Replies))
		cmmts.WriteString("</p></div>\n")
	}
	cmmts.WriteString(s.NavBase(dlist{
		Pages: 0,
		More:  c.HasMore,
	}))
	cmmts.WriteString("</details>")
	return cmmts.String()
}

func ParseMedia(media devianter.Media) string {
	url := devianter.UrlFromMedia(media)
	if len(url) != 0 {
		url = url[21:]
		dot := strings.Index(url, ".")

		return UrlBuilder("media", "file", url[:dot], "/", url[dot+10:])
	}
	return ""
}

func (s skunkyart) DownloadAndSendMedia(subdomain, path string) {
	var url strings.Builder
	url.WriteString("https://images-wixmp-")
	url.WriteString(subdomain)
	url.WriteString(".wixmp.com/")
	url.WriteString(path)
	url.WriteString("?token=")
	url.WriteString(s.Args.Get("token"))

	download := func() (body []byte, status int, headers http.Header) {
		cli := &http.Client{}
		if CFG.WixmpProxy != "" {
			u, e := u.Parse(CFG.WixmpProxy)
			err(e)
			cli.Transport = &http.Transport{Proxy: http.ProxyURL(u)}
		}

		req, e := http.NewRequest("GET", url.String(), nil)
		err(e)
		req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:123.0) Gecko/20100101 Firefox/123.0.0")

		resp, e := cli.Do(req)
		err(e)
		defer resp.Body.Close()

		b, e := io.ReadAll(resp.Body)
		err(e)
		return b, resp.StatusCode, resp.Header
	}

	if CFG.Cache.Enabled {
		os.Mkdir(CFG.Cache.Path, 0700)
		fname := CFG.Cache.Path + "/" + base64.StdEncoding.EncodeToString([]byte(subdomain+path))
		file, e := os.Open(fname)

		if e != nil {
			b, status, headers := download()
			if status == 200 && headers["Content-Type"][0][:5] == "image" {
				err(os.WriteFile(fname, b, 0700))
				s.Writer.Write(b)
			}
		} else {
			file, e := io.ReadAll(file)
			err(e)
			s.Writer.Write(file)
		}
	} else if CFG.Proxy {
		b, _, _ := download()
		s.Writer.Write(b)
	} else {
		s.Writer.WriteHeader(403)
		s.Writer.Write([]byte("Sorry, butt proxy on this instance disabled."))
	}
}

func InitCacheSystem() {
	c := &CFG.Cache
	for {
		dir, e := os.Open(c.Path)
		err(e)
		stat, e := dir.Stat()
		err(e)

		dirnames, e := dir.Readdirnames(-1)
		err(e)
		for _, a := range dirnames {
			a = c.Path + "/" + a
			rm := func() {
				err(os.RemoveAll(a))
			}
			if c.Lifetime != 0 {
				now := time.Now().UnixMilli()

				f, _ := os.Stat(a)
				stat := f.Sys().(*syscall.Stat_t)
				time := time.Unix(stat.Ctim.Unix()).UnixMilli()

				if time+c.Lifetime <= now {
					rm()
				}
			}
			if c.MaxSize != 0 && stat.Size() > c.MaxSize {
				rm()
			}
		}

		dir.Close()
		time.Sleep(time.Second * time.Duration(CFG.Cache.UpdateInterval))
	}
}

func CopyTemplatesToMemory() {
	try := func(e error) {
		if e != nil {
			panic(e.Error())
		}
	}

	dir, e := os.ReadDir(CFG.TemplatesDir)
	try(e)

	for _, x := range dir {
		n := CFG.TemplatesDir + "/" + x.Name()
		file, e := os.ReadFile(n)
		try(e)
		Templates[n] = string(file)
	}
}
