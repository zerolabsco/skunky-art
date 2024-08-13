// TODO: реализовать кеширование JSON и почистить код
package app

import (
	"crypto/sha1"
	"encoding/hex"
	"io"
	"os"
	"strings"
	"sync"
	"syscall"
	"time"
)

type file struct {
	Score   int
	Content []byte
}

var tempFS = make(map[[20]byte]*file)
var mx = &sync.RWMutex{}

func (s skunkyart) DownloadAndSendMedia(subdomain, path string) {
	var url strings.Builder
	url.WriteString("https://images-wixmp-")
	url.WriteString(subdomain)
	url.WriteString(".wixmp.com/")
	url.WriteString(path)
	url.WriteString("?token=")
	url.WriteString(s.Args.Get("token"))

	var response []byte

	switch {
	case CFG.Cache.Enabled:
		fileName := sha1.Sum([]byte(subdomain + path))
		filePath := CFG.Cache.Path + "/" + hex.EncodeToString(fileName[:])

		mx.Lock()
		if tempFS[fileName] == nil {
			tempFS[fileName] = &file{}
		}
		f := *tempFS[fileName]
		mx.Unlock()

		if f.Content != nil {
			f.Score += 2
		} else {
			file, err := os.Open(filePath)
			if err != nil {
				if dwnld := Download(url.String()); dwnld.Status == 200 && dwnld.Headers["Content-Type"][0][:5] == "image" {
					f.Content = dwnld.Body
					try(os.WriteFile(filePath, f.Content, 0700))
				} else {
					s.ReturnHTTPError(dwnld.Status)
					return
				}
			} else {
				file, e := io.ReadAll(file)
				try(e)
				f.Content = file
			}

			go func() {
				defer restore()
				for {
					time.Sleep(1 * time.Minute)

					mx.Lock()
					if tempFS[fileName].Score <= 0 {
						delete(tempFS, fileName)
						mx.Unlock()
						return
					}
					tempFS[fileName].Score--
					mx.Unlock()
				}
			}()
		}

		mx.Lock()
		tempFS[fileName] = &f
		mx.Unlock()
		response = f.Content
	case CFG.Proxy:
		dwnld := Download(url.String())
		if dwnld.Status != 200 {
			s.ReturnHTTPError(dwnld.Status)
			return
		}
		response = dwnld.Body
	default:
		s.Writer.WriteHeader(403)
		response = []byte("Sorry, butt proxy on this instance are disabled.")
	}

	s.Writer.Write(response)
}

func InitCacheSystem() {
	c := &CFG.Cache
	os.Mkdir(c.Path, 0700)
	for {
		dir, e := os.Open(c.Path)
		try(e)
		stat, e := dir.Stat()
		try(e)

		dirnames, e := dir.Readdirnames(-1)
		try(e)
		for _, a := range dirnames {
			a = c.Path + "/" + a
			if c.Lifetime != "" {
				now := time.Now().UnixMilli()

				f, _ := os.Stat(a)
				stat := f.Sys().(*syscall.Stat_t)
				time := statTime(stat)

				if time+lifetimeParsed <= now {
					try(os.RemoveAll(a))
				}
			}
			if c.MaxSize != 0 && stat.Size() > c.MaxSize {
				try(os.RemoveAll(a))
			}
		}

		dir.Close()
		time.Sleep(time.Second * time.Duration(c.UpdateInterval))
	}
}
