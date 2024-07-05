package app

import (
	"io"
	"net/http"
	u "net/url"
	"os"
	"strconv"
	"strings"
)

const addr string = "0.0.0.0:3003"

// роутер
func Router() {
	parsepath := func(path string) map[int]string {
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

		arg := skunky.Args.Get
		skunky.Query = u.QueryEscape(arg("q"))
		if t := arg("type"); len(t) > 0 {
			skunky.Type = rune(t[0])
		}
		p, _ := strconv.Atoi(arg("p"))
		skunky.Page = p
		if arg("atom") == "true" {
			skunky.atom = true
		}

		// пути
		switch path[1] {
		default:
			skunky.ReturnHTTPError(404)
		case "":
			open_n_send("html/index.htm")
		case "post":
			skunky.Deviation(path[2], path[3])
		case "search":
			skunky.Search()
		case "dd":
			skunky.DD()
		case "group_user":
			skunky.GRUser()

		case "media":
			skunky.Emojitar(path[2])
		case "about":
			open_n_send("html/about.htm")
		case "gui":
			w.Header().Add("content-type", "text/css")
			open_n_send(next(path, 2))
		}
	}

	http.HandleFunc("/", handle)
	http.ListenAndServe(addr, nil)
}

func err(e error) {
	if e != nil {
		println(e.Error())
	}
}
