package app

import (
	"io"
	"net/http"
	u "net/url"
	"os"
	"strconv"
	"strings"
)

func Router() {
	parsepath := func(path string) map[int]string {
		if l := len(CFG.BasePath); len(path) > l {
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

	// функция, что управляет всем
	handle := func(w http.ResponseWriter, r *http.Request) {
		path := parsepath(r.URL.Path)
		var wr = io.WriteString
		open_n_send := func(name string) {
			f, e := os.ReadFile(name)
			err(e)
			wr(w, string(f))
		}

		// структура с функциями
		var skunky skunkyart
		skunky.Args = r.URL.Query()
		skunky.Writer = w
		skunky.BasePath = CFG.BasePath

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

		// пути
		switch path[1] {
		default:
			skunky.ReturnHTTPError(404)
		case "":
			skunky.ExecuteTemplate("html/index.htm", &CFG.BasePath)
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
				skunky.DownloadAndSendMedia(path[3], next(path, 4))
			case "emojitar":
				skunky.Emojitar(path[3])
			}
		case "about":
			skunky.About()
		case "gui":
			w.Header().Add("content-type", "text/css")
			open_n_send(next(path, 2))
		}
	}

	http.HandleFunc("/", handle)
	http.ListenAndServe(CFG.Listen, nil)
}

func err(e error) {
	if e != nil {
		println(e.Error())
	}
}
