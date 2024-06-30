package app

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"git.macaw.me/skunky/devianter"
)

var wr = io.WriteString

type skunkyart struct {
	Writer    http.ResponseWriter
	Args      url.Values
	Type      rune
	Query     string
	Page      int
	Templates struct {
		GroupUser struct {
			GR           devianter.GRuser
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
				Pages int
				List  string
			}
		}
		Search struct {
			Content devianter.Search
			List    string
		}
	}
}

func (s skunkyart) GRUser() {
	if len(s.Query) < 1 {
		s.ReturnHTTPError(400)
		return
	}

	var g devianter.Group
	g.Name = s.Query
	s.Templates.GroupUser.GR = g.GroupFunc()
	group := &s.Templates.GroupUser

	switch s.Type {
	case 'a':
		if g := group.GR; !g.Owner.Group {
			for _, x := range g.Gruser.Page.Modules {
				switch x.Name {
				case "about":
					group.About.A = x.ModuleData.About
					var about = group.About.A
					group.About.DescriptionFormatted = ParseDescription(about.Description)
					group.About.Comments = s.ParseComments(devianter.CommentsFunc(
						strconv.Itoa(group.GR.Gruser.ID),
						"",
						s.Page,
						4,
					))

					for _, val := range x.ModuleData.About.SocialLinks {
						var social strings.Builder
						social.WriteString(`<a target="_blank" href="`)
						social.WriteString(val.Value)
						social.WriteString(`">`)
						social.WriteString(val.Value)
						social.WriteString("</a><br>")
						group.About.Social += social.String()
					}

					for _, val := range x.ModuleData.About.Interests {
						var interest strings.Builder
						interest.WriteString(val.Label)
						interest.WriteString(": <b>")
						interest.WriteString(val.Value)
						interest.WriteString("</b><br>")
						group.About.Interests += interest.String()
					}

					if rd := x.ModuleData.About.RegDate; rd != 0 {
						group.CreationDate = time.Unix(time.Now().Unix()-rd, 0).UTC().String()
					}
				case "cover_deviation":
					group.About.BGMeta = x.ModuleData.CoverDeviation.Deviation
					group.About.BG = devianter.UrlFromMedia(group.About.BGMeta.Media)
				}
			}
		} else {

		}
	case 'g':
		gallery := g.Gallery(s.Page)
		fmt.Println(gallery)
		for _, x := range gallery.Content.Gruser.Page.Modules {
			group.Gallery.List = s.DeviationList(x.ModuleData.Folder.Deviations, dlist{
				Pages: x.ModuleData.Folder.Pages,
			})
		}
	default:
		s.ReturnHTTPError(400)
	}

	s.ExecuteTemplate("html/gruser.htm", &s)
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

		post.Comments = s.ParseComments(devianter.CommentsFunc(id, post.Post.Comments.Cursor, s.Page, 1))

		s.ExecuteTemplate("html/deviantion.htm", &post)
	} else {
		s.ReturnHTTPError(400)
	}
}

func (s skunkyart) DD() {
	dd := devianter.DailyDeviationsFunc(s.Page)
	s.ExecuteTemplate("html/list.htm", s.DeviationList(dd.Deviations, dlist{
		Pages: 0,
		More:  dd.HasMore,
	}))
}

func (s skunkyart) Search() {
	var e error
	ss := &s.Templates.Search
	switch s.Type {
	case 'a', 't':
		ss.Content, e = devianter.SearchFunc(s.Query, s.Page, s.Type)
	case 'g':
		ss.Content, e = devianter.SearchFunc(s.Query, s.Page, s.Type, s.Args.Get("usr"))
	default:
		s.ReturnHTTPError(400)
	}
	err(e)

	ss.List = s.DeviationList(ss.Content.Results, dlist{
		Pages: ss.Content.Pages,
		More:  ss.Content.HasMore,
	})

	s.ExecuteTemplate("html/search.htm", &s)
}

func (s skunkyart) Emojitar(name string) {
	if name != "" && (s.Type == 'a' || s.Type == 'e') {
		ae, e := devianter.AEmedia(name, s.Type)
		if e != nil {
			s.ReturnHTTPError(404)
			println(e.Error())
		}
		wr(s.Writer, ae)
	} else {
		s.ReturnHTTPError(400)
	}
}
