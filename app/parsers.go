package app

import (
	"encoding/json"
	"strconv"
	"strings"

	"git.macaw.me/skunky/devianter"
	"golang.org/x/net/html"
)

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
	cmmts.WriteString(s.NavBase(DeviationList{
		Pages: 0,
		More:  c.HasMore,
	}))
	cmmts.WriteString("</details>")
	return cmmts.String()
}

func (s skunkyart) DeviationList(devs []devianter.Deviation, content ...DeviationList) string {
	var list strings.Builder
	if s.Atom && s.Page > 1 {
		s.ReturnHTTPError(400)
		return ""
	} else if s.Atom {
		list.WriteString(`<?xml version="1.0" encoding="UTF-8"?><feed xmlns:media="http://search.yahoo.com/mrss/" xmlns="http://www.w3.org/2005/Atom">`)
		list.WriteString(`<title>`)
		if s.Type == 0 {
			list.WriteString("Daily Deviations")
		} else if len(devs) != 0 {
			list.WriteString(devs[0].Author.Username)
		} else {
			list.WriteString("SkunkyArt")
		}
		list.WriteString(`</title>`)

		list.WriteString(`<link rel="alternate" href="`)
		list.WriteString(Host)
		list.WriteString(`"/>`)
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
				list.WriteString(ConvertDeviantArtUrlToSkunkyArt(data.Url))
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
				list.WriteString(ConvertDeviantArtUrlToSkunkyArt(data.Url))
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

/* DESCRIPTION/COMMENT PARSER */
type text struct {
	TXT     string
	TXT_RAW string
	From    int
	To      int
}

func ParseDescription(dscr devianter.Text) string {
	var parseddescription strings.Builder
	TagBuilder := func(content string, tags ...string) string {
		l := len(tags)
		for x := 0; x < l; x++ {
			var htm strings.Builder
			htm.WriteString("<")
			htm.WriteString(tags[x])
			htm.WriteString(">")

			htm.WriteString(content)

			htm.WriteString("</")
			htm.WriteString(tags[x])
			htm.WriteString(">")
			content = htm.String()
		}
		return content
	}
	DeleteTrackingFromUrl := func(url string) string {
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
		try(e)

		entities := make(map[int]devianter.Deviation)
		urls := make(map[int]string)
		for n, x := range descr.EntityMap {
			num, _ := strconv.Atoi(n)
			if x.Data.Url != "" {
				urls[num] = DeleteTrackingFromUrl(x.Data.Url)
			}
			entities[num] = x.Data.Data
		}

		for _, x := range descr.Blocks {
			Styles := make([]text, len(x.InlineStyleRanges))

			if len(x.InlineStyleRanges) != 0 {
				var tags = make(map[int][]string)
				for n, rngs := range x.InlineStyleRanges {
					Styles := &Styles[n]
					switch rngs.Style {
					case "BOLD":
						rngs.Style = "b"
					case "UNDERLINE":
						rngs.Style = "u"
					case "ITALIC":
						rngs.Style = "i"
					}
					Styles.From = rngs.Offset
					Styles.To = rngs.Offset + rngs.Length
					FT := Styles.From * Styles.To
					tags[FT] = append(tags[FT], rngs.Style)
				}
				for n := 0; n < len(Styles); n++ {
					Styles := &Styles[n]
					Styles.TXT_RAW = x.Text[Styles.From:Styles.To]
					Styles.TXT = TagBuilder(Styles.TXT_RAW, tags[Styles.From*Styles.To]...)
				}
			}

			switch x.Type {
			case "atomic":
				d := entities[x.EntityRanges[0].Key]
				parseddescription.WriteString(`<a href="`)
				parseddescription.WriteString(ConvertDeviantArtUrlToSkunkyArt(d.Url))
				parseddescription.WriteString(`"><img width="50%" src="`)
				parseddescription.WriteString(ParseMedia(d.Media))
				parseddescription.WriteString(`" title="`)
				parseddescription.WriteString(d.Author.Username)
				parseddescription.WriteString(" - ")
				parseddescription.WriteString(d.Title)
				parseddescription.WriteString(`"></a>`)
			case "unstyled":
				if l := len(Styles); l != 0 {
					for n, r := range Styles {
						var tag string
						if x.Type == "header-two" {
							tag = "h2"
						}

						parseddescription.WriteString(x.Text[:r.From])
						if len(urls) != 0 && len(x.EntityRanges) != 0 {
							ra := &x.EntityRanges[0]

							parseddescription.WriteString(`<a target="_blank" href="`)
							parseddescription.WriteString(urls[ra.Key])
							parseddescription.WriteString(`">`)
							parseddescription.WriteString(r.TXT)
							parseddescription.WriteString(`</a>`)
						} else if l > n+1 {
							parseddescription.WriteString(r.TXT)
						}
						parseddescription.WriteString(TagBuilder(tag, x.Text[r.To:]))
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
							url := DeleteTrackingFromUrl(a.Val)
							parseddescription.WriteString(`<a target="_blank" href="`)
							parseddescription.WriteString(url)
							parseddescription.WriteString(`">`)
							parseddescription.WriteString(GetValueOfTag(tt))
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
