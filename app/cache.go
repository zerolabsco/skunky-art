package app

// TODO: implement JSON caching and clean up the code.

import (
	"crypto/sha1" //nolint:gosec // G505: SHA-1 is a cache-key hash here, not a security primitive
	"encoding/hex"
	"io"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"
)

type file struct {
	Score   int
	Content []byte
}

// tempFS is the in-memory media cache, guarded by mx. A plain Mutex rather than
// an RWMutex on purpose: every operation here mutates something (a read bumps
// Score), and the previous code took an RLock to write, which is not exclusive.
var tempFS = make(map[[20]byte]*file)
var mx sync.Mutex

// memGet returns the cached body for key and raises its score so that popular
// entries outlive the janitor, or nil when the entry is absent or still empty.
func memGet(key [20]byte) []byte {
	mx.Lock()
	defer mx.Unlock()

	f := tempFS[key]
	if f == nil || f.Content == nil {
		return nil
	}
	f.Score += 2
	return f.Content
}

// memPut caches body under key. An empty body is not cached, so a failed fetch
// cannot poison the cache with a zero-length image.
func memPut(key [20]byte, body []byte) {
	if len(body) == 0 {
		return
	}

	mx.Lock()
	defer mx.Unlock()
	tempFS[key] = &file{Content: body}
}

// InitMemCacheJanitor ages the in-memory cache forever, dropping entries whose
// score has run out. Run it in its own goroutine, once, and only when memcache
// is enabled.
//
// One loop ages the whole map. The previous design started a goroutine per
// cached file, each looping until its own entry was evicted, and each touching
// the map without holding mx — a concurrent map read and write, which the Go
// runtime treats as a fatal error that recover cannot catch.
func InitMemCacheJanitor() {
	for {
		time.Sleep(1 * time.Minute)
		ageMemCache()
	}
}

// ageMemCache runs one round of aging: every entry loses a point, and entries
// that are already out of points are dropped. An entry starts at zero, so a body
// nothing asks for again is gone within a round.
func ageMemCache() {
	mx.Lock()
	defer mx.Unlock()

	for k, f := range tempFS {
		if f.Score <= 0 {
			delete(tempFS, k)
			continue
		}
		f.Score--
	}
}

// mediaSubdomain matches the one hostname label wixmp media URLs vary: a hex
// string, sometimes with dashes. Anything outside that set is rejected rather
// than escaped, because this label is what selects the host to fetch from.
var mediaSubdomain = regexp.MustCompile(`^[a-zA-Z0-9-]+$`)

// buildMediaURL returns the wixmp CDN URL for one media item, reporting false
// when subdomain is not a bare hostname label.
//
// subdomain and path arrive already percent-decoded from the request path, so
// they can carry the characters that end a host. Concatenated into a URL string,
// a subdomain of "x@attacker.example#" reparses as host attacker.example, with
// "images-wixmp-x" demoted to userinfo and the intended host to a fragment —
// pointing the fetch at whatever the caller names, including addresses reachable
// only from the instance itself.
func buildMediaURL(subdomain, path, token string) (string, bool) {
	if !mediaSubdomain.MatchString(subdomain) {
		return "", false
	}

	// Fields rather than concatenation: String escapes the path, so a decoded
	// "#" or "?" in it stays part of the path instead of ending it. The host is
	// checked above rather than escaped, because url.URL passes it through
	// verbatim.
	u := url.URL{
		Scheme: "https",
		Host:   "images-wixmp-" + subdomain + ".wixmp.com",
		Path:   "/" + path,
	}
	if token != "" {
		u.RawQuery = url.Values{"token": {token}}.Encode()
	}
	return u.String(), true
}

// DownloadAndSendMedia proxies one image from DeviantArt's wixmp CDN to the
// client, serving it from the on-disk or in-memory cache when enabled. It
// responds 403 when proxying is turned off for this instance.
func (s skunkyart) DownloadAndSendMedia(subdomain, path string) {
	mediaURL, ok := buildMediaURL(subdomain, path, s.Args.Get("token"))
	if !ok {
		s.ReturnHTTPError(400)
		return
	}

	var response []byte

	switch {
	case CFG.Cache.Enabled:
		key := sha1.Sum([]byte(subdomain + path)) //nolint:gosec // G401: cache-key hash, not a security primitive
		filePath := CFG.Cache.Path + "/" + hex.EncodeToString(key[:])

		if CFG.Cache.MemCache {
			if cached := memGet(key); cached != nil {
				response = cached
				break
			}
		}

		body, ok := s.loadOrFetchMedia(filePath, mediaURL)
		if !ok {
			// loadOrFetchMedia has already written the error response.
			return
		}
		response = body

		if CFG.Cache.MemCache {
			memPut(key, response)
		}
	case CFG.Proxy:
		dwnld := Download(mediaURL)
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

// loadOrFetchMedia returns the media body for filePath, preferring the on-disk
// cache and falling back to fetching mediaURL, which it then writes back to the
// cache. It reports false when it has already written an error response, so the
// caller must not write anything further.
func (s skunkyart) loadOrFetchMedia(filePath, mediaURL string) ([]byte, bool) {
	// filePath is built from a SHA-1 of the request, not from user input, so it
	// cannot escape the cache directory.
	if f, err := os.Open(filePath); err == nil { //nolint:gosec // G304: path is a hash, not user-controlled
		defer func() { try(f.Close()) }()

		if body, err := io.ReadAll(f); err == nil {
			return body, true
		} else {
			// An unreadable cache entry is not fatal; re-fetch it instead.
			try(err)
		}
	}

	dwnld := Download(mediaURL)
	if dwnld.Status != 200 || !strings.HasPrefix(dwnld.Headers.Get("Content-Type"), "image") {
		s.ReturnHTTPError(dwnld.Status)
		return nil, false
	}

	try(os.WriteFile(filePath, dwnld.Body, 0600))
	return dwnld.Body, true
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
