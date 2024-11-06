package vlsp

import (
	"os"
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
