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
	if t := s.Args.Get("token"); t != "" {
		url.WriteString("?token=")
		url.WriteString(t)
	}

	var response []byte

	switch {
	case CFG.Cache.Enabled:
		fileName := sha1.Sum([]byte(subdomain + path))
		filePath := CFG.Cache.Path + "/" + hex.EncodeToString(fileName[:])

		c := func() {
			file, err := os.Open(filePath)
			if err != nil {
				if dwnld := Download(url.String()); dwnld.Status == 200 && dwnld.Headers["Content-Type"][0][:5] == "image" {
					response = dwnld.Body
					try(os.WriteFile(filePath, response, 0700))
				} else {
					s.ReturnHTTPError(dwnld.Status)
					return
				}
			} else {
				file, e := io.ReadAll(file)
				try(e)
				response = file
			}
		}

		if CFG.Cache.MemCache {
			mx.Lock()
			if tempFS[fileName] == nil {
				tempFS[fileName] = &file{}
			}
			mx.Unlock()

			if tempFS[fileName].Content != nil {
				response = tempFS[fileName].Content
				tempFS[fileName].Score += 2
				break
			} else {
				c()
				go func() {
					defer restore()

					mx.RLock()
					tempFS[fileName].Content = response
					mx.RUnlock()

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
		} else {
			c()
		}
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
	for {
		dir, err := os.ReadDir(c.Path)
		if err != nil {
			if os.IsNotExist(err) {
				os.Mkdir(c.Path, 0700)
				continue
			}
			println(err.Error())
		}

		var total int64
		for _, file := range dir {
			fileName := c.Path + "/" + file.Name()
			fileInfo, err := file.Info()
			try(err)

			if c.Lifetime != "" {
				now := time.Now().UnixMilli()

				stat := fileInfo.Sys().(*syscall.Stat_t)
				time := statTime(stat)

				if time+lifetimeParsed <= now {
					try(os.RemoveAll(fileName))
				}
			}

			total += fileInfo.Size()
			// if c.MaxSize != 0 && fileInfo.Size() > c.MaxSize {
			// 	try(os.RemoveAll(fileName))
			// }
		}

		if c.MaxSize != 0 && total > c.MaxSize {
			try(os.RemoveAll(c.Path))
			os.Mkdir(c.Path, 0700)
		}

		time.Sleep(time.Second * time.Duration(c.UpdateInterval))
	}
}
