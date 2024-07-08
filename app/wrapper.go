package app

import (
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
var Templates = make(map[string]string)

type skunkyart struct {
	Writer          http.ResponseWriter
	Args            url.Values
	BasePath        string
	Type            rune
	Query, QueryRaw string
	Page            int
	Atom            bool
	Templates       struct {
		About struct {
			Proxy bool
			Nsfw  bool
		}

		SomeList  string
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
		g := group.GR
		s.Atom = false
		for _, x := range g.Gruser.Page.Modules {
			switch x.Name {
			case "about", "group_about":
				switch g.Owner.Group {
				case true:
					var about = &x.ModuleData.GroupAbout
					group.Group = true
					group.CreationDate = x.ModuleData.GroupAbout.FoundatedAt.UTC().String()
					group.About.DescriptionFormatted = ParseDescription(about.Description)
				case false:
					group.About.A = x.ModuleData.About
					var about = &group.About.A
					group.CreationDate = time.Unix(time.Now().Unix()-x.ModuleData.About.RegDate, 0).UTC().String()

					group.About.DescriptionFormatted = ParseDescription(about.Description)

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
				}
				group.About.Comments = s.ParseComments(devianter.CommentsFunc(
					strconv.Itoa(group.GR.Gruser.ID),
					"",
					s.Page,
					4,
				))

			case "cover_deviation":
				group.About.BGMeta = x.ModuleData.CoverDeviation.Deviation
				group.About.BGMeta.Url = s.ConvertDeviantArtUrlToSkunkyArt(group.About.BGMeta.Url)
				group.About.BG = s.ParseMedia(group.About.BGMeta.Media)
			case "group_admins":
				var htm strings.Builder
				for _, z := range x.ModuleData.GroupAdmins.Results {
					htm.WriteString(`<div class="admin"><img src="`)
					htm.WriteString(UrlBuilder("media", "emojitar", z.User.Username, "?type=a"))
					htm.WriteString(`"><a href="`)
					htm.WriteString(UrlBuilder("group_user", "?type=about&q=", z.User.Username))
					htm.WriteString(`">`)
					htm.WriteString(z.User.Username)
					htm.WriteString(`</a></div>`)
				}
				group.Admins += htm.String()
			}

		}
	case 'g':
		folderid, _ := strconv.Atoi(s.Args.Get("folder"))
		if s.Page == 0 {
			s.Page++
		}

		gallery := g.Gallery(s.Page, folderid)
		if folderid > 0 {
			group.Gallery.List = s.DeviationList(gallery.Content.Results, dlist{
				More: gallery.Content.HasMore,
			})
		} else {
			for _, x := range gallery.Content.Gruser.Page.Modules {
				if l := len(x.ModuleData.Folders.Results); l != 0 {
					var folders strings.Builder
					folders.WriteString(`<h1 id="folders"><a href="#folder">#</a> Folders</h1><div class="folders"><br>`)
					for _, x := range x.ModuleData.Folders.Results {
						folders.WriteString(`<div class="block folder-item">`)

						folders.WriteString(`<a href="`)
						folders.WriteString(s.ConvertDeviantArtUrlToSkunkyArt(x.Thumb.Url))
						folders.WriteString(`"><img loading="lazy" src="`)
						folders.WriteString(s.ParseMedia(x.Thumb.Media))
						folders.WriteString(`" title="`)
						folders.WriteString(x.Thumb.Title)
						folders.WriteString(`"></a><br>`)

						folders.WriteString(`<a href="?folder=`)
						folders.WriteString(strconv.Itoa(x.FolderId))
						folders.WriteString("&q=")
						folders.WriteString(s.Query)
						folders.WriteString("&type=")
						folders.WriteString(string(s.Type))
						folders.WriteString(`">`)
						folders.WriteString(x.Name)
						folders.WriteString(`</a>`)

						folders.WriteString("</div>")
					}
					folders.WriteString(`</div><h1 id="content"><a href="#content">#</a> Content</h1>`)
					group.Gallery.Folders = folders.String()
				}

				if x.Name == "folder_deviations" {
					group.Gallery.List = s.DeviationList(x.ModuleData.Folder.Deviations, dlist{
						Pages: x.ModuleData.Folder.Pages,
						More:  x.ModuleData.Folder.HasMore,
					})
				}
			}
		}
	default:
		s.ReturnHTTPError(400)
	}

	if !s.Atom {
		s.ExecuteTemplate("html/gruser.htm", &s)
	}
}

// посты
func (s skunkyart) Deviation(author, postname string) {
	id_search := regexp.MustCompile("[0-9]+").FindAllString(postname, -1)
	if len(id_search) >= 1 {
		post := &s.Templates.Deviation

		id := id_search[len(id_search)-1]
		post.Post = devianter.DeviationFunc(id, author)

		post.Post.Description = ParseDescription(post.Post.Deviation.TextContent)
		// время публикации
		post.StringTime = post.Post.Deviation.PublishedTime.UTC().String()
		post.Post.IMG = s.ParseMedia(post.Post.Deviation.Media)
		for _, x := range post.Post.Deviation.Extended.RelatedContent {
			if len(x.Deviations) != 0 {
				post.Related = s.DeviationList(x.Deviations)
			}
		}

		// хештэги
		for _, x := range post.Post.Deviation.Extended.Tags {
			var tag strings.Builder
			tag.WriteString(` <a href="`)
			tag.WriteString(UrlBuilder("search", "?q=", x.Name, "&type=tag"))
			tag.WriteString(`">#`)
			tag.WriteString(x.Name)
			tag.WriteString("</a>")

			post.Tags += tag.String()
		}

		if post.Post.Comments.Total <= 50 {
			post.Post.Comments.Cursor = ""
		}

		post.Comments = s.ParseComments(devianter.CommentsFunc(id, post.Post.Comments.Cursor, s.Page, 1))

		s.ExecuteTemplate("html/deviantion.htm", &s)
	} else {
		s.ReturnHTTPError(400)
	}
}

func (s skunkyart) DD() {
	dd := devianter.DailyDeviationsFunc(s.Page)
	s.Templates.SomeList = s.DeviationList(dd.Deviations, dlist{
		Pages: 0,
		More:  dd.HasMore,
	})
	if !s.Atom {
		s.ExecuteTemplate("html/list.htm", &s)
	}
}

func (s skunkyart) Search() {
	s.Atom = false
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
		}
		wr(s.Writer, ae)
	} else {
		s.ReturnHTTPError(400)
	}
}

func (s skunkyart) About() {
	s.Templates.About.Nsfw = CFG.Nsfw
	s.Templates.About.Proxy = CFG.Proxy
	s.ExecuteTemplate("html/about.htm", &s)
}
