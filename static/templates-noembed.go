//go:build !embed
// +build !embed

package static

import (
	"bytes"
	"io/fs"
	"os"
	"strings"
	"time"
)

var Templates FS

type file struct {
	path    string
	name    string
	content []byte
}

var templateNames = []string{}
var templates = make(map[string][]file)
var StaticPath string

func CopyTemplatesToMemory() {
	baseDir, err := os.ReadDir(StaticPath)
	try(err)

	for _, c := range baseDir {
		if c.IsDir() {
			templateNames = append(templateNames, c.Name())

			var filePath strings.Builder
			filePath.WriteString(StaticPath)
			filePath.WriteString("/")
			filePath.WriteString(c.Name())

			dir, err := os.ReadDir(filePath.String())
			try(err)

			filePath.WriteString("/")
			for _, cd := range dir {
				f, err := os.ReadFile(filePath.String() + cd.Name())
				try(err)
				templates[c.Name()] = append(templates[c.Name()], file{
					content: f,
					name:    cd.Name(),
					path:    c.Name() + "/" + cd.Name(),
				})
			}
		}
	}
}

type FS struct{}

func (FS) Open(name string) (fs.File, error) {
	for i, l := 0, len(templateNames); i < l; i++ {
		for _, x := range templates[templateNames[i]] {
			if x.content != nil && name == x.path {
				return &File{
					name:    x.path,
					content: bytes.NewBuffer(x.content),
				}, nil
			}
		}
	}
	return nil, &fs.PathError{}
}

func (FS) Glob(pattern string) ([]string, error) {
	trimmed := strings.Split(pattern, "/")
	var matches = []string{}
	for x, s := range templates {
		for i, l := 0, len(s); i < l && trimmed[0] == x; i++ {
			s := s[i]
			matches = append(matches, s.path)
		}
	}
	if len(matches) != 0 {
		return matches, nil
	}
	return nil, &fs.PathError{}
}

func try(err error) {
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
}

/* based on https://github.com/psanford/memfs; required for templates.ParseFS to work correctly */
type fileInfo struct {
	name string
}

func (fi fileInfo) Name() string {
	return fi.name
}

func (fi fileInfo) Size() int64 {
	return 4096
}

func (fileInfo) Mode() fs.FileMode {
	return 0
}

func (fileInfo) ModTime() time.Time {
	return time.Time{}
}

func (fileInfo) IsDir() bool {
	return false
}

func (fileInfo) Sys() interface{} {
	return nil
}

type File struct {
	name    string
	content *bytes.Buffer
	closed  bool
}

func (f *File) Stat() (fs.FileInfo, error) {
	return fileInfo{
		name: f.name,
	}, nil
}

func (f *File) Read(b []byte) (int, error) {
	if f.closed {
		return 0, fs.ErrClosed
	}
	return f.content.Read(b)
}

func (f *File) Close() error {
	if f.closed {
		return fs.ErrClosed
	}
	f.closed = true
	return nil
}
