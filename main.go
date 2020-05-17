//go:generate statik -f -src /tmp/files-statik/
package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rakyll/statik/fs"

	_ "tesk/statik" // TODO: Replace with the absolute import path
)

type neuteredFileSystem struct {
	sfs http.FileSystem
}

type nfsFile struct {
	io.Closer
	io.Reader
	io.Seeker
	contentReader io.ReadSeeker
	fileInfo      nfsFileInfo
}

func (nf *nfsFile) Readdir(count int) ([]os.FileInfo, error) {
	var x []os.FileInfo
	return x, fmt.Errorf("no dir")
}
func (nf *nfsFile) Stat() (os.FileInfo, error)       { return &nf.fileInfo, nil }
func (nf *nfsFile) Close() error                     { return nil }
func (nf *nfsFile) Read(p []byte) (n int, err error) { return nf.contentReader.Read(p) }
func (nf *nfsFile) Seek(offset int64, whence int) (int64, error) {
	return nf.contentReader.Seek(offset, whence)
}

type nfsFileInfo struct {
	os.FileInfo
	name string
	sz   int64
}

func (nfi *nfsFileInfo) Name() string       { return nfi.name }
func (nfi *nfsFileInfo) Size() int64        { return nfi.sz }
func (nfi *nfsFileInfo) Mode() os.FileMode  { return 0755 }
func (nfi *nfsFileInfo) ModTime() time.Time { return time.Time{} }
func (nfi *nfsFileInfo) IsDir() bool        { return false }

func (nfi *nfsFileInfo) Sys() interface{} { return nil }

func (nfs *neuteredFileSystem) RootIndexAs(path string) (http.File, error) {
	nfile := &nfsFile{}
	index := "/index.html"
	fmt.Println("...serving", index, "in place of", path)
	f, err := nfs.sfs.Open(index)
	if err == nil {
		defer f.Close()
		contents, _ := ioutil.ReadAll(f)
		nfile.contentReader = bytes.NewReader(contents)
		nfile.fileInfo.name = path
		nfile.fileInfo.sz = int64(len(contents))
		fmt.Println("...file info", nfile.fileInfo)
		return nfile, nil
	}
	fmt.Println("...read of index failed", err)
	return nil, err
}

func (nfs neuteredFileSystem) Open(path string) (http.File, error) {
	// we want to detect all sub-dirs here and just serve root /index.html
	path = filepath.Clean(path)
	if len(path) > 1 && strings.LastIndex(path, "/") == len(path)-1 {
		return nfs.RootIndexAs(path)
	}
	fmt.Println("opening", path)
	f, err := nfs.sfs.Open(path)
	if err != nil {
		fmt.Println("...bad open", err)
		// just return /index.html for anything we do not understand here and let
		// the frontend deal with the issue
		return nfs.RootIndexAs(path)
	}

	return f, err
}

func main() {
	// ...
	var err error
	nfs := &neuteredFileSystem{}
	nfs.sfs, err = fs.New()
	if err != nil {
		log.Fatal(err)
	}

	// Serve the contents over HTTP.
	http.Handle("/", http.StripPrefix("/", http.FileServer(nfs)))
	// http.Handle("/", http.StripPrefix("/", http.FileServer(statikFS)))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
