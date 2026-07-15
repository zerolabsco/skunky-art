package app

import (
	"io"
	"net/http"
	url "net/url"
	"skunkyart/static"
	"strconv"
	"strings"
)

var Host string

func Router() {
	parsepath := func(path string) map[int]string {
		if l := len(CFG.URI); len(path) > l {
			path = path[l-1:]
		} else {
			path = "/"
		}

		parsedpath := make(map[int]string)
		for x := 0; true; x++ {
			slash := strings.Index(path, "/") + 1
			content := path[:slash]
			path = path[slash:]
			if slash == 0 {
				parsedpath[x] = path
				break
			}
			parsedpath[x] = content[:slash-1]
		}
		return parsedpath
	}

	next := func(path map[int]string, from int) (out string) {
		for x, l := from, len(path)-1; x <= l; x++ {
			out += path[x]
			if x != l {
				out += "/"
			}
		}
		return
	}

	open := func(name string) []byte {
		file, err := static.Templates.Open(name)
		try(err)
		fileReaded, err := io.ReadAll(file)
		try(err)

		return fileReaded
	}

	// the function that drives everything
	handle := func(w http.ResponseWriter, r *http.Request) {
		path := parsepath(r.URL.Path)
		Host = "http://" + r.Host
		if h := r.Header["X-Forwarded-Proto"]; len(h) != 0 && h[0] == "https" {
			Host = "https://" + r.Host
		}

		var skunky = skunkyart{Version: Release.Version}
		skunky._pth = r.URL.Path

		skunky.Args = r.URL.Query()
		arg := skunky.Args.Get
		p, _ := strconv.Atoi(arg("p"))

		skunky.Endpoint = path[1]
		skunky.API.main = &skunky
		skunky.Writer = w
		skunky.BasePath = CFG.URI
		skunky.QueryRaw = arg("q")
		skunky.Query = url.QueryEscape(skunky.QueryRaw)
		skunky.Page = p

		if t := arg("type"); len(t) > 0 {
			skunky.Type = rune(t[0])
		}

		if arg("atom") == "true" {
			skunky.Atom = true
		}

		if CFG.Proxy {
			w.Header().Add("Content-Security-Policy", "default-src 'self'; script-src 'none'; style-src 'self' 'unsafe-inline'")
		} else {
			w.Header().Add("Content-Security-Policy", "default-src 'self'; img-src 'self' *.wixmp.com; script-src 'none'; style-src 'self' 'unsafe-inline'")
		}

		w.Header().Add("X-Frame-Options", "DENY")

		switch skunky.Endpoint {
		// main
		case "":
			skunky.ExecuteTemplate("index.htm", "html", &CFG.URI)
		case "about":
			skunky.Templates.About = About
			skunky.ExecuteTemplate("about.htm", "html", &skunky)
		case "post":
			skunky.Deviation(path[2], path[3])
		case "search":
			skunky.Search()
		case "dd":
			skunky.DD()
		case "group_user":
			skunky.GRUser()

		// media
		case "media":
			switch path[2] {
			case "file":
				if a := arg("filename"); a != "" {
					skunky.SetFilename(a)
				}
				skunky.DownloadAndSendMedia(path[3], next(path, 4))
			case "emojitar":
				skunky.Emojitar(path[3])
			}
		case "stylesheet":
			w.Header().Add("content-type", "text/css")
			w.Write(open("css/skunky.css"))
		case "favicon.ico":
			w.Write(open("images/logo.png"))

		// API
		case "api":
			w.Header().Add("Content-Type", "application/json")
			switch path[2] {
			case "instance":
				skunky.API.Info()
			case "random":
				skunky.API.Random()
			default:
				skunky.API.Error("Not Found", 404)
			}

		// 404
		default:
			skunky.ReturnHTTPError(404)
		}
	}

	http.HandleFunc("/", handle)
	println("SkunkyArt is listening on", CFG.Listen)

	tryWithExitStatus(http.ListenAndServe(CFG.Listen, nil), 1)
}
