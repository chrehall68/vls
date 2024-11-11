package vlsp

import (
	"os"
	"runtime"
	"strings"
)

// File represents a file, either on disk or in memory
// It represents an abstraction to deal with the complexity of
// document changes
type File struct {
	Path string

	savedOnDisk bool
	contents    string
}

func NewFile(path string) *File {
	return &File{Path: path, savedOnDisk: true, contents: ""}
}

// GetContents returns the whole contents of the file
func (f File) GetContents() string {
	if !f.savedOnDisk {
		return f.contents
	} else {
		file, _ := os.ReadFile(f.Path)
		return string(file)
	}
}

// SetContents sets the contents of the file;
// assumes that the file is not saved on disk
func (f *File) SetContents(contents string) {
	f.contents = contents
	f.savedOnDisk = false
}

// Save marks the file as saved on disk
func (f *File) Save() {
	f.savedOnDisk = true
	f.contents = ""
}

func URIToPath(uri string) string {
	os := runtime.GOOS
	uri = strings.TrimPrefix(uri, "file://")

	// make sure to handle windows case
	if os == "windows" {
		// it's windows
		uri = strings.ReplaceAll(uri, "%3A", ":")
		// and also strip leading /
		uri = uri[1:]
	}
	return uri
}
func PathToURI(path string) string {
	os := runtime.GOOS
	if os == "windows" {
		return "file:///" + path // need to add the leading / again
	}
	return "file://" + path
}
