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
	"time"

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
	if c.More {
		prevrev("| Next =>", p+1, false)
	}

	return list.String()
}

func (s skunkyart) GRUser() {
	var group struct {
		GR           devianter.GRuser
		CreationDate string

		About struct {
			A devianter.About

			DescriptionFormatted string
			Interests            string
			Social               string
			BG                   devianter.Deviation
		}
	}

	if len(s.Query) < 1 {
		s.httperr(400)
		return
	}

	var g devianter.Group
	g.Name = s.Query
	group.GR = g.GroupFunc()

	if g := group.GR; !g.Owner.Group {
		for _, x := range g.Gruser.Page.Modules {
			var about = group.About.A
			if x.ModuleData.About.RegDate != 0 {
				about = x.ModuleData.About
			}
			group.About.DescriptionFormatted = ParseDescription(about.Description)

			for _, val := range x.ModuleData.About.Interests {
				var interest strings.Builder
				interest.WriteString(val.Label)
				interest.WriteString(": <b>")
				interest.WriteString(val.Value)
				interest.WriteString("</b><br>")
				group.About.Interests += interest.String()
			}

			for _, val := range x.ModuleData.About.SocialLinks {
				var social strings.Builder
				social.WriteString(`<a target="_blank" href="`)
				social.WriteString(val.Value)
				social.WriteString(`">`)
				social.WriteString(val.Value)
				social.WriteString("</a><br>")
				group.About.Social += social.String()
			}

			if rd := x.ModuleData.About.RegDate; rd != 0 {
				group.CreationDate = time.Unix(time.Now().Unix()-rd, 0).UTC().String()
			}
		}
	} else {

	}

	s.exe("html/gruser.htm", &group)
}

func (s skunkyart) DeviationList(devs []devianter.Deviation, content ...dlist) string {
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
	list.WriteString(s.NavBase(content[0]))

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

		post.Post.Description = ParseDescription(post.Post.Deviation.TextContent)
		// время публикации
		post.StringTime = post.Post.Deviation.PublishedTime.UTC().String()

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

		// FIXME: первый комментарий не отображается.
		// генерация комментов
		var cmmts strings.Builder
		replied := make(map[int]string)
		c := devianter.CommentsFunc(id, post.Post.Comments.Cursor, s.Page, 1)

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
			cmmts.WriteString(`"><img src="/media/`)
			cmmts.WriteString(x.User.Username)
			cmmts.WriteString(`?type=a" width="30px" height="30px"><a href="/user/`)
			cmmts.WriteString(x.User.Username)
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

			cmmts.WriteString(x.Comment)
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

		post.Comments = cmmts.String()

		s.exe("html/deviantion.htm", &post)
	} else {
		s.httperr(400)
	}
}

func (s skunkyart) DD() {
	dd := devianter.DailyDeviationsFunc(s.Page)
	s.exe("html/list.htm", s.DeviationList(dd.Deviations, dlist{
		Pages: 0,
		More:  dd.HasMore,
	}))
}

func (s skunkyart) Search() {
	// тут всё и так понятно
	switch s.Type {
	case 'a', 't', 'g':
		var srch struct {
			Search devianter.Search
			List   string
		}

		var e error
		srch.Search, e = devianter.SearchFunc(s.Query, s.Page, s.Type)
		err(e)
		srch.List = s.DeviationList(srch.Search.Results, dlist{
			Pages: srch.Search.Pages,
			More:  srch.Search.HasMore,
		})

		s.exe("html/search.htm", &srch)
	default:
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
