package app

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"git.macaw.me/skunky/devianter"
	"golang.org/x/net/html"
)

var wr = io.WriteString
var Templates = make(map[string]string)

type skunkyart struct {
	Writer http.ResponseWriter

	Args            url.Values
	BasePath        string
	Type            rune
	Query, QueryRaw string
	Page            int
	Atom            bool

	Templates struct {
		About struct {
			Proxy     bool
			Nsfw      bool
			Instances []struct {
				Title   string
				Country string
				Urls    []struct {
					I2P      string `json:"i2p"`
					Ygg      string
					Tor      string
					Clearnet string
				}
				Settings struct {
					Nsfw  bool
					Proxy bool
				}
			}
		}

		SomeList  string
		DDStrips  string
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
	var err error
	s.Templates.GroupUser.GR, err = g.GetGroup()
	try(err)

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
				group.About.Comments = s.ParseComments(devianter.GetComments(
					strconv.Itoa(group.GR.Gruser.ID),
					"",
					s.Page,
					4,
				))

			case "cover_deviation":
				group.About.BGMeta = x.ModuleData.CoverDeviation.Deviation
				group.About.BGMeta.Url = ConvertDeviantArtUrlToSkunkyArt(group.About.BGMeta.Url)
				group.About.BG = ParseMedia(group.About.BGMeta.Media)
			case "group_admins":
				var htm strings.Builder
				for _, z := range x.ModuleData.GroupAdmins.Results {
					htm.WriteString(BuildUserPlate(z.User.Username))
				}
				group.Admins += htm.String()
			}

		}
	case 'g':
		folderid, _ := strconv.Atoi(s.Args.Get("folder"))
		if s.Page == 0 {
			s.Page++
		}

		gallery, err := g.GetGallery(s.Page, folderid)
		try(err)

		if folderid > 0 {
			group.Gallery.List = s.DeviationList(gallery.Content.Results, true, DeviationList{
				More: gallery.Content.HasMore,
			})
		} else {
			for _, x := range gallery.Content.Gruser.Page.Modules {
				if l := len(x.ModuleData.Folders.Results); l != 0 {
					var folders strings.Builder
					folders.WriteString(`<h1 id="folders"><a href="#folder">#</a> Folders</h1><div class="folders"><br>`)
					for _, x := range x.ModuleData.Folders.Results {
						folders.WriteString(`<div class="block folder-item">`)

						if !(x.Thumb.NSFW && !CFG.Nsfw) {
							folders.WriteString(`<a href="`)
							folders.WriteString(ConvertDeviantArtUrlToSkunkyArt(x.Thumb.Url))
							folders.WriteString(`"><img loading="lazy" src="`)
							folders.WriteString(ParseMedia(x.Thumb.Media))
							folders.WriteString(`" title="`)
							folders.WriteString(x.Thumb.Title)
							folders.WriteString(`"></a>`)
						} else {
							folders.WriteString(`<h1>[ <span class="nsfw">NSFW</span> ]</h1>`)
						}
						folders.WriteString("<br>")

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
					group.Gallery.List = s.DeviationList(x.ModuleData.Folder.Deviations, true, DeviationList{
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
		s.ExecuteTemplate("gruser.htm", &s)
	}
}

// посты
func (s skunkyart) Deviation(author, postname string) {
	id_search := regexp.MustCompile("[0-9]+").FindAllString(postname, -1)
	if len(id_search) >= 1 {
		post := &s.Templates.Deviation

		id := id_search[len(id_search)-1]
		post.Post = devianter.GetDeviation(id, author)

		if post.Post.Deviation.TextContent.Excerpt != "" {
			post.Post.Description = ParseDescription(post.Post.Deviation.TextContent)
		} else {
			post.Post.Description = ParseDescription(post.Post.Deviation.Extended.DescriptionText)
		}
		// время публикации
		post.StringTime = post.Post.Deviation.PublishedTime.UTC().String()
		post.Post.IMG = ParseMedia(post.Post.Deviation.Media)
		for _, x := range post.Post.Deviation.Extended.RelatedContent {
			if len(x.Deviations) != 0 {
				post.Related += s.DeviationList(x.Deviations, false)
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

		post.Comments = s.ParseComments(devianter.GetComments(id, post.Post.Comments.Cursor, s.Page, 1))

		s.ExecuteTemplate("deviantion.htm", &s)
	} else {
		s.ReturnHTTPError(400)
	}
}

func (s skunkyart) DD() {
	dd := devianter.GetDailyDeviations(s.Page)
	var strips strings.Builder
	for _, x := range dd.Strips {
		strips.WriteString(`<h3 class="`)
		strips.WriteString(x.Codename)
		strips.WriteString(`"> <a href="#`)
		strips.WriteString(x.Codename)
		strips.WriteString(`"># </a>`)
		strips.WriteString(x.Title)
		strips.WriteString(`</h3>`)

		strips.WriteString(s.DeviationList(x.Deviations, false))
	}
	s.Templates.DDStrips = strips.String()
	s.Templates.SomeList = s.DeviationList(dd.Deviations, true, DeviationList{
		Pages: 0,
		More:  dd.HasMore,
	})
	if !s.Atom {
		s.ExecuteTemplate("daily.htm", &s)
	}
}

func (s skunkyart) Search() {
	s.Atom = false
	var err error
	ss := &s.Templates.Search
	switch s.Type {
	case 'a', 't':
		ss.Content, err = devianter.PerformSearch(s.Query, s.Page, s.Type)
	case 'g':
		ss.Content, err = devianter.PerformSearch(s.Query, s.Page, s.Type, s.Args.Get("usr"))
	case 'r': // скраппер, поскольку девиантартовцы зажопили гостевое API для поиска групп
		var (
			usernames = make(map[int]string)
			url       strings.Builder
			num       int
		)

		s.Page++

		url.WriteString("https://www.deviantart.com/groups/?q=")
		url.WriteString(s.Query)
		if s.Page > 1 {
			url.WriteString("&offset=")
			url.WriteString(strconv.Itoa(10 * s.Page))
		}

		dwnld := Download(url.String())

		for z := html.NewTokenizer(strings.NewReader(string(dwnld.Body))); ; {
			if n, token := z.Next(), z.Token(); n == html.StartTagToken && token.Data == "a" {
				for _, x := range token.Attr {
					if x.Key == "class" && x.Val == "u regular username" {
						usernames[num] = GetValueOfTag(z)
						num++
					}
				}
			} else if n == 0 {
				break
			} else {
				continue
			}
		}

		if l := len(usernames); l != 0 {
			ss.List += `<div class="content plates">`
			for x := 0; x < len(usernames); x++ {
				ss.List += BuildUserPlate(usernames[x])
			}
			ss.List += `</div>`
			ss.List += s.NavBase(DeviationList{
				More: true,
			})
		}
	default:
		s.ReturnHTTPError(400)
	}
	try(err)

	if s.Type != 'r' {
		ss.List = s.DeviationList(ss.Content.Results, false, DeviationList{
			Pages: ss.Content.Pages,
			More:  ss.Content.HasMore,
		})
	}

	s.ExecuteTemplate("search.htm", &s)
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
	try(json.Unmarshal([]byte(Templates["instances.json"]), &s.Templates.About))
	s.ExecuteTemplate("about.htm", &s)
}
