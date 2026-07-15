// Package app implements SkunkyArt: a JavaScript-free alternative frontend for
// DeviantArt.
//
// It fetches upstream data through the devianter library, renders it into static
// HTML (or Atom feeds) server-side, and optionally proxies and caches media so
// that no request from the browser reaches DeviantArt directly.
package app
