package app

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"text/template"

	"git.macaw.me/skunky/devianter"
)

// парсинг темплейтов
func (s skunkyart) ExecuteTemplate(file string, data any) {
	var buf strings.Builder
	tmp, e := template.ParseFiles(file)
	err(e)
	err(tmp.Execute(&buf, &data))
	wr(s.Writer, buf.String())
}

func (s skunkyart) ReturnHTTPError(status int) {
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

type text struct {
	TXT  string
	from int
	to   int
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

	if description, dl := dscr.Html.Markup, len(dscr.Html.Markup); dl != 0 &&
		description[0] == '{' &&
		description[dl-1] == '}' {
		var descr struct {
			Blocks []struct {
				Key, Text, Type   string
				InlineStyleRanges []struct {
					Offset, Length int
					Style          string
				}
			}
		}
		e := json.Unmarshal([]byte(description), &descr)
		err(e)

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

			for _, r := range ranges {
				var tag string
				switch x.Type {
				case "header-two":
					tag = "h2"
				case "unstyled":
					tag = "p"
				}
				parseddescription.WriteString(r.TXT)
				parseddescription.WriteString(TagBuilder(tag, x.Text[r.to:]))
			}
		}
	} else if dl != 0 {
		parseddescription.WriteString(description)
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

// FIXME: первый комментарий не отображается.
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
		cmmts.WriteString(`"><img src="/media/`)
		cmmts.WriteString(x.User.Username)
		cmmts.WriteString(`?type=a" width="30px" height="30px"><a href="/group_user?type=about&q=`)
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
	return cmmts.String()
}
