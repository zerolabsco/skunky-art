package app

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"text/template"

	"git.macaw.me/skunky/devianter"
)

var wr = io.WriteString

type skunkyart struct {
	Writer http.ResponseWriter
	Args   url.Values
	Type   rune
	Query  string
	Page   int
}

// парсинг темплейтов
func (s skunkyart) exe(file string, data any) {
	var buf bytes.Buffer
	tmp, e := template.ParseFiles(file)
	err(e)
	tmp.Execute(&buf, &data)
	wr(s.Writer, buf.String())
}

func (s skunkyart) httperr(status int) {
	s.Writer.WriteHeader(status)

	// пострйока с помощью strings.Builder, потому что такой метод быстрее обычного сложения
	var msg strings.Builder
	msg.WriteString(`<html><link rel="stylesheet" href="/gui/css/skunky.css" />`)
	msg.WriteString("<h1>")
	msg.WriteString(strconv.Itoa(status))
	msg.WriteString(" - ")
	msg.WriteString(http.StatusText(status))
	msg.WriteString("</h1></html>")

	wr(s.Writer, msg.String())
}

func (s skunkyart) DeviationList(devs []devianter.Deviation) string {
	var list strings.Builder
	list.WriteString(`<div class="content">`)
	for _, data := range devs {
		url := devianter.UrlFromMedia(data.Media)

		list.WriteString(`<a title="open/download" href="`)
		list.WriteString(url)
		list.WriteString(`"><div class="block"><img src="`)
		list.WriteString(url)
		list.WriteString(`" width="15%"></a><br><a href="`)
		list.WriteString("/post/")
		list.WriteString(data.Author.Username)
		list.WriteString("/")
		list.WriteString(data.Url[27:][strings.Index(data.Url[27:], "/art/")+5:])
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
	list.WriteString("</div>")
	return list.String()
}

// посты
func (s skunkyart) Deviation(author, postname string) {
	// поиск ID
	re := regexp.MustCompile("[0-9]+").FindAllString(postname, -1)
	if len(re) >= 1 {
		var post struct {
			Post       devianter.Post
			StringTime string
			Tags       string
			Comments   string
		}

		id := re[len(re)-1]
		post.Post = devianter.DeviationFunc(id, author)

		// время публикации
		post.StringTime = post.Post.Deviation.PublishedTime.UTC().String()

		println(post.Post.Description)
		// хештэги
		for _, x := range post.Post.Deviation.Extended.Tags {
			var tag strings.Builder
			tag.WriteString(` <a href="/search?q=`)
			tag.WriteString(x.Name)
			tag.WriteString(`&type=tag">#`)
			tag.WriteString(x.Name)
			tag.WriteString("</a>")

			post.Tags += tag.String()
		}

		// генерация комментов
		var cmmts strings.Builder
		var replied map[int]string
		_ = replied
		c := devianter.CommentsFunc(id, post.Post.Comments.Cursor, s.Page, 1)

		cmmts.WriteString("<details><summary>Comments: <b>")
		cmmts.WriteString(strconv.Itoa(c.Total))
		cmmts.WriteString("</b></summary>")
		for _, x := range c.Thread {
			cmmts.WriteString(`<div class="msg"><p id="`)
			cmmts.WriteString(strconv.Itoa(x.ID))
			cmmts.WriteString(`"><img src="/media/`)
			cmmts.WriteString(x.User.Username)
			cmmts.WriteString(`?type=a" width="30px" height="30px"><a href="/user/`)
			cmmts.WriteString(x.User.Username)
			cmmts.WriteString(`"><b`)
			if x.User.Banned {
				cmmts.WriteString(` class="banned"`)
			}
			if x.Author {
				cmmts.WriteString(` class="author"`)
			}
			cmmts.WriteString(">")
			cmmts.WriteString(x.User.Username)
			cmmts.WriteString("</b></a> ")
			cmmts.WriteString(x.Posted.UTC().String())
			cmmts.WriteString("<p>")
			cmmts.WriteString(x.Comment)
			cmmts.WriteString("<p>👍: ")
			cmmts.WriteString(strconv.Itoa(x.Likes))
			cmmts.WriteString(" ⏩: ")
			cmmts.WriteString(strconv.Itoa(x.Replies))
			cmmts.WriteString("</p></div>\n")
		}
		cmmts.WriteString("</details>")
		post.Comments = cmmts.String()

		s.exe("html/deviantion.htm", &post)
	} else {
		s.httperr(400)
	}
}

func (s skunkyart) Search() {
	// тут всё и так понятно
	if s.Type == 'a' || s.Type == 't' || s.Type == 'g' {
		var srch struct {
			Search devianter.Search
			List   string
		}

		var e error
		srch.Search, e = devianter.SearchFunc(s.Query, s.Page, s.Type)
		err(e)
		srch.List = s.DeviationList(srch.Search.Results)

		s.exe("html/search.htm", &srch)
	} else {
		s.httperr(400)
	}
}

func (s skunkyart) Emojitar(name string) {
	if name != "" && (s.Type == 'a' || s.Type == 'e') {
		ae, e := devianter.AEmedia(name, s.Type)
		if e != nil {
			s.httperr(404)
			println(e.Error())
		}
		wr(s.Writer, ae)
	} else {
		s.httperr(400)
	}
}
