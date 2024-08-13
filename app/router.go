package app

import (
	"io"
	"net/http"
	u "net/url"
	"skunkyart/static"
	"strconv"
	"strings"
)

var Host, Path string

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

	// функция, что управляет всем
	handle := func(w http.ResponseWriter, r *http.Request) {
		if h := r.Header["X-Forwarded-Proto"]; len(h) != 0 && h[0] == "https" {
			Host = h[0] + "://" + r.Host
		} else {
			Host = "http://" + r.Host
		}

		Path = r.URL.Path
		path := parsepath(Path)
		// структура с функциями
		var skunky skunkyart
		skunky.Writer = w
		skunky.Args = r.URL.Query()
		skunky.BasePath = CFG.URI

		arg := skunky.Args.Get
		skunky.QueryRaw = arg("q")
		skunky.Query = u.QueryEscape(skunky.QueryRaw)

		if t := arg("type"); len(t) > 0 {
			skunky.Type = rune(t[0])
		}
		p, _ := strconv.Atoi(arg("p"))
		skunky.Page = p

		if arg("atom") == "true" {
			skunky.Atom = true
		}

		skunky.Endpoint = path[1]

		// пути
		switch skunky.Endpoint {
		default:
			skunky.ReturnHTTPError(404)
		case "":
			w.Write(open("html/index.htm"))
		case "post":
			skunky.Deviation(path[2], path[3])
		case "search":
			skunky.Search()
		case "dd":
			skunky.DD()
		case "group_user":
			skunky.GRUser()

		case "media":
			switch path[2] {
			case "file":
				if a := arg("filename"); a != "" {
					var filename strings.Builder
					filename.WriteString(`filename="`)
					filename.WriteString(a)
					filename.WriteString(`"`)
					w.Header().Add("Content-Disposition", filename.String())
				}
				skunky.DownloadAndSendMedia(path[3], next(path, 4))
			case "emojitar":
				skunky.Emojitar(path[3])
			}
		case "about":
			skunky.About()
		case "stylesheet":
			w.Header().Add("content-type", "text/css")
			w.Write(open("css/skunky.css"))
		case "favicon.ico":
			w.Write(open("images/logo.png"))
		}
	}

	http.HandleFunc("/", handle)
	println("SkunkyArt is listening on", CFG.Listen)

	tryWithExitStatus(http.ListenAndServe(CFG.Listen, nil), 1)
}
