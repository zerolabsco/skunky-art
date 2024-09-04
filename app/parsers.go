package app

import (
	"encoding/json"
	"strconv"
	"strings"

	"git.macaw.me/skunky/devianter"
	"golang.org/x/net/html"
)

func (s skunkyart) ParseComments(c devianter.Comments, daError devianter.Error) string {
	if daError.RAW != nil {
		return "Failed to fetch comments :("
	}

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
			cmmts.WriteString(` In reply to <a href="`)
			cmmts.WriteString(Path)
			cmmts.WriteString("#")
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

func (s skunkyart) DeviationList(devs []devianter.Deviation, allowAtom bool, content ...DeviationList) string {
	if s.Atom && s.Page > 1 {
		s.ReturnHTTPError(400)
		return ""
	}

	var list, listContent strings.Builder

	for i, l := 0, len(devs); i < l; i++ {
		data := &devs[i]
		if preview, fullview := ParseMedia(data.Media, 320), ParseMedia(data.Media); !(data.NSFW && !CFG.Nsfw) {
			if allowAtom && s.Atom {
				s.Writer.Header().Add("Content-type", "application/atom+xml")
				id := strconv.Itoa(data.ID)
				listContent.WriteString(`<entry><author><name>`)
				listContent.WriteString(data.Author.Username)
				listContent.WriteString(`</name></author><title>`)
				listContent.WriteString(data.Title)
				listContent.WriteString(`</title><link rel="alternate" type="text/html" href="`)
				listContent.WriteString(UrlBuilder("post", data.Author.Username, "atom-"+id))
				listContent.WriteString(`"/><id>`)
				listContent.WriteString(id)
				listContent.WriteString(`</id><published>`)
				listContent.WriteString(data.PublishedTime.UTC().Format("Mon, 02 Jan 2006 15:04:05 -0700"))
				listContent.WriteString(`</published>`)
				listContent.WriteString(`<media:group><media:title>`)
				listContent.WriteString(data.Title)
				listContent.WriteString(`</media:title><media:thumbinal url="`)
				listContent.WriteString(preview)
				listContent.WriteString(`"/></media:group><content type="xhtml"><div xmlns="http://www.w3.org/1999/xhtml"><a href="`)
				listContent.WriteString(ConvertDeviantArtUrlToSkunkyArt(data.Url))
				listContent.WriteString(`"><img src="`)
				listContent.WriteString(fullview)
				listContent.WriteString(`"/></a><p>`)
				listContent.WriteString(ParseDescription(data.TextContent))
				listContent.WriteString(`</p></div></content></entry>`)
			} else {
				listContent.WriteString(`<div class="block">`)
				if fullview != "" && preview != "" {
					listContent.WriteString(`<a title="open/download" href="`)
					listContent.WriteString(fullview)
					listContent.WriteString(`"><img loading="lazy" src="`)
					listContent.WriteString(preview)
					listContent.WriteString(`" width="15%"></a>`)
				} else {
					listContent.WriteString(`<h1>[ TEXT ]</h1>`)
				}
				listContent.WriteString(`<br><a href="`)
				listContent.WriteString(ConvertDeviantArtUrlToSkunkyArt(data.Url))
				listContent.WriteString(`">`)
				listContent.WriteString(data.Author.Username)
				listContent.WriteString(" - ")
				listContent.WriteString(data.Title)

				if data.NSFW {
					listContent.WriteString(` [<span class="nsfw">NSFW</span>]`)
				}
				if data.AI {
					listContent.WriteString(" [🤖]")
				}
				if data.DD {
					listContent.WriteString(` [<span class="dd">DD</span>]`)
				}

				listContent.WriteString("</a></div>")
			}
		}
	}

	if allowAtom && s.Atom {
		list.WriteString(`<?xml version="1.0" encoding="UTF-8"?><feed xmlns:media="http://search.yahoo.com/mrss/" xmlns="http://www.w3.org/2005/Atom">`)

		list.WriteString(`<title>`)
		if s.Type == 0 {
			list.WriteString("Daily Deviations")
		} else if s.Type == 'g' && len(devs) != 0 {
			list.WriteString(devs[0].Author.Username)
		} else {
			list.WriteString("SkunkyArt")
		}
		list.WriteString(`</title>`)

		list.WriteString(`<link rel="alternate" href="`)
		list.WriteString(Host)
		list.WriteString(`"/>`)

		list.WriteString(listContent.String())

		list.WriteString("</feed>")
		wr(s.Writer, list.String())
	} else {
		list.WriteString(`<div class="content">`)

		list.WriteString(listContent.String())

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

// переписать весь этот пиздец нахуй
func ParseDescription(dscr devianter.Text) string {
	var parsedDescription strings.Builder
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
				if len(x.EntityRanges) != 0 {
					d := entities[x.EntityRanges[0].Key]
					parsedDescription.WriteString(`<a href="`)
					parsedDescription.WriteString(ConvertDeviantArtUrlToSkunkyArt(d.Url))
					parsedDescription.WriteString(`"><img width="50%" src="`)
					parsedDescription.WriteString(ParseMedia(d.Media))
					parsedDescription.WriteString(`" title="`)
					parsedDescription.WriteString(d.Author.Username)
					parsedDescription.WriteString(" - ")
					parsedDescription.WriteString(d.Title)
					parsedDescription.WriteString(`"></a>`)
				}
			case "unstyled":
				if l := len(Styles); l != 0 {
					for n, r := range Styles {
						var tag string
						if x.Type == "header-two" {
							tag = "h2"
						}

						parsedDescription.WriteString(x.Text[:r.From])
						if len(urls) != 0 && len(x.EntityRanges) != 0 {
							ra := &x.EntityRanges[0]

							parsedDescription.WriteString(`<a target="_blank" href="`)
							parsedDescription.WriteString(urls[ra.Key])
							parsedDescription.WriteString(`">`)
							parsedDescription.WriteString(r.TXT)
							parsedDescription.WriteString(`</a>`)
						} else if l > n+1 {
							parsedDescription.WriteString(r.TXT)
						}
						parsedDescription.WriteString(TagBuilder(tag, x.Text[r.To:]))
					}
				} else {
					parsedDescription.WriteString(x.Text)
				}
			}
			parsedDescription.WriteString("<br>")
		}
	} else if dl != 0 {
		for tt := html.NewTokenizer(strings.NewReader(dscr.Html.Markup)); ; {
			switch tt.Next() {
			case html.ErrorToken:
				return parsedDescription.String()
			case html.StartTagToken, html.EndTagToken, html.SelfClosingTagToken:
				token := tt.Token()
				switch token.Data {
				case "a":
					for _, a := range token.Attr {
						if a.Key == "href" {
							url := DeleteTrackingFromUrl(a.Val)
							parsedDescription.WriteString(`<a target="_blank" href="`)
							parsedDescription.WriteString(url)
							parsedDescription.WriteString(`">`)
							parsedDescription.WriteString(GetValueOfTag(tt))
							parsedDescription.WriteString("</a> ")
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
								parsedDescription.WriteString(`<img src="`)
								parsedDescription.WriteString(uri)
								parsedDescription.WriteString(`" title="`)
								parsedDescription.WriteString(title)
								parsedDescription.WriteString(`">`)
							}
						}
					}
				case "br", "li", "ul", "p", "b":
					parsedDescription.WriteString(token.String())
				case "div":
					parsedDescription.WriteString("<p> ")
				}
			case html.TextToken:
				parsedDescription.Write(tt.Text())
			}
		}
	}

	return parsedDescription.String()
}
