package app

// TODO: implement JSON caching and clean up the code.

import (
	"crypto/sha1" //nolint:gosec // G505: SHA-1 is a cache-key hash here, not a security primitive
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

// DownloadAndSendMedia proxies one image from DeviantArt's wixmp CDN to the
// client, serving it from the on-disk or in-memory cache when enabled. It
// responds 403 when proxying is turned off for this instance.
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
		fileName := sha1.Sum([]byte(subdomain + path)) //nolint:gosec // G401: cache-key hash, not a security primitive
		filePath := CFG.Cache.Path + "/" + hex.EncodeToString(fileName[:])

		c := func() {
			// filePath is built from a SHA-1 of the request, not from user input,
			// so it cannot escape the cache directory.
			file, err := os.Open(filePath) //nolint:gosec // G304: path is a hash, not user-controlled
			if err != nil {
				dwnld := Download(url.String())
				if dwnld.Status == 200 && strings.HasPrefix(dwnld.Headers.Get("Content-Type"), "image") {
					response = dwnld.Body
					try(os.WriteFile(filePath, response, 0600))
				} else {
					s.ReturnHTTPError(dwnld.Status)
					return
				}
			} else {
				defer func() { try(file.Close()) }()
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

	_, _ = s.Writer.Write(response)
}

// InitCacheSystem runs the cache rotation loop forever, evicting files past
// their lifetime and emptying the cache when it outgrows max-size. Run it in its
// own goroutine.
func InitCacheSystem() {
	c := &CFG.Cache
	for {
		dir, err := os.ReadDir(c.Path)
		if err != nil {
			if os.IsNotExist(err) {
				try(os.Mkdir(c.Path, 0700))
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

				// Sys() is platform-specific and only documented to be a
				// *syscall.Stat_t on unix; skip rotation rather than panic
				// if the filesystem reports something else.
				if stat, ok := fileInfo.Sys().(*syscall.Stat_t); ok {
					if statTime(stat)+lifetimeParsed <= now {
						try(os.RemoveAll(fileName))
					}
				}
			}

			total += fileInfo.Size()
			// if c.MaxSize != 0 && fileInfo.Size() > c.MaxSize {
			// 	try(os.RemoveAll(fileName))
			// }
		}

		if c.MaxSize != 0 && total > c.MaxSize {
			try(os.RemoveAll(c.Path))
			try(os.Mkdir(c.Path, 0700))
		}

		time.Sleep(time.Second * time.Duration(c.UpdateInterval))
	}
}
