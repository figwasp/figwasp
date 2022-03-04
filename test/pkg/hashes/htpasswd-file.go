package hashes

import (
	"fmt"
	"io/ioutil"
	"os"

	"golang.org/x/crypto/bcrypt"
)

type HtpasswdFile struct {
	path string
}

func NewHtpasswdFile(username, password string) (f *HtpasswdFile, e error) {
	const (
		tempFileDirectory = ""
		tempFilePattern   = "*"

		cost = 5 // htpasswd default

		htpasswdFormat = "%s:%s"
	)

	var (
		file *os.File
		hash []byte
	)

	file, e = ioutil.TempFile(tempFileDirectory, tempFilePattern)
	if e != nil {
		return
	}

	hash, e = bcrypt.GenerateFromPassword(
		[]byte(password),
		cost,
	)
	if e != nil {
		return
	}

	_, e = fmt.Fprintf(file, htpasswdFormat,
		username,
		hash,
	)
	if e != nil {
		return
	}

	file.Close()

	f = &HtpasswdFile{
		path: file.Name(),
	}

	return
}

func (f *HtpasswdFile) Path() string {
	return f.path
}

func (f *HtpasswdFile) Destroy() (e error) {
	e = os.Remove(f.path)
	if e != nil {
		return
	}

	return
}
