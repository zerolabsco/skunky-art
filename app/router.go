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
	// расшифровка эндпоинта из урл
	endpoint := func(url string) (string, string) {
		if CFG.Base_uri != "" {
			url = strings.Replace(url, CFG.Base_uri, "/", 1)
		}

		end := strings.Index(url[1:], "/")
		if end == -1 {
			return url[1:], ""
		}
		return url[1 : end+1], url[end+2:]
	}

	// функция, что управляет всем
	handle := func(w http.ResponseWriter, r *http.Request) {
		e, url := endpoint(r.URL.Path)
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

		// пути
		switch e {
		default:
			skunky.httperr(404)
		case "/", "":
			open_n_send("html/index.htm")
		case "post":
			slash := strings.Index(url, "/")
			skunky.Deviation(url[:slash], url[slash+1:])
		case "search":
			skunky.Search()
		case "dd":
			skunky.DD()
		case "group":
			skunky.GRUser()

		case "media":
			skunky.Emojitar(url)
		case "about":
			open_n_send("html/about.htm")
		case "gui":
			w.Header().Add("content-type", "text/css")
			open_n_send(url)
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
