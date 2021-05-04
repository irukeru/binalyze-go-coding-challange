// Package magicbytes Some file formats are intended to be read different than text-based files because they have special
// formats. In order to determine a file's format/type, magic bytes/magical numbers are used to mark files with special
// signatures, located at the beginning of the file, mostly. You are assigned to create a Go package to find files
// recursively in a target directory with the following API for given file meta information.

package magicbytes

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"runtime"
)

// MaxMetaArrayLength holds Maximum length for meta type array
const MaxMetaArrayLength = 1000

// ErrMetaArrayLengthExceeded holds meta array length exceeded max value error
var ErrMetaArrayLengthExceeded = errors.New("Meta array length exceeded max value")

// Meta holds the name, magical bytes, and offset of the magical bytes to be searched.
type Meta struct {
	Type   string // name of the file/meta type.
	Bytes  []byte // magical bytes.
	Offset int64  // offset of the magical bytes from the file start position.
}

// OnMatchFunc represents a function to be called when Search function finds a match.
// Returning false must immediately stop Search process.
type OnMatchFunc func(path, metaType string) bool

// Search searches the given target directory to find files recursively using meta information.
// For every match, onMatch callback is called concurrently.
func Search(ctx context.Context, targetDir string, metas []*Meta, onMatch OnMatchFunc) error {

	if len(metas) > MaxMetaArrayLength {
		return ErrMetaArrayLengthExceeded
	}

	// no need to search files if meta array is empty
	if len(metas) == 0 {
		return nil
	}

	PathChannel := make(chan string, runtime.NumCPU())
	defer close(PathChannel)

	for i := 0; i < runtime.NumCPU(); i++ {
		go findMatchWorker(PathChannel, onMatch, metas)
	}

	// need to fix file path for running os
	p := filepath.FromSlash(targetDir)
	err := readDir(ctx, p, PathChannel)
	if err != nil {
		log.Println("readDir error: ", err)
		return err
	}

	return nil
}

// findMatchWorker receives the path messages from channel and runs the findMath function
func findMatchWorker(pathChannel <-chan string, onMatch OnMatchFunc, metas []*Meta) {

	defer func() {
		if recover() != nil {
			return
		}
	}()

	for path := range pathChannel {
		metaType, status := findMatch(path, metas)
		if status {
			if !onMatch(path, metaType) {
				return
			}
		}
	}
}

// readDir gets the file list and sends it via channel by using filePath.WalkDir method
func readDir(ctx context.Context, root string, pathChannel chan<- string) error {

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Println("Unable to read directory: ", err)
		} else if d.Type().IsRegular() {
			pathChannel <- path
		}
		return nil
	})
	if err != nil {
		log.Println("filepath walk error: ", err)
	}

	return nil
}

// findMatch mission is find the file with given meta data
func findMatch(path string, meta []*Meta) (string, bool) {

	for i := 0; i < len(meta); i++ {
		if checkMetaData(path, *meta[i]) {
			return meta[i].Type, true
		}
	}

	return "", false
}

// checkMetaData checks file's initial bytes and return true if given meta data is matched with file
func checkMetaData(filename string, meta Meta) bool {

	file, err := os.Open(filename)
	defer file.Close()

	if err != nil {
		log.Println("Unable to open file: ", err)
		return false
	}

	fi, err := file.Stat()
	if err != nil {
		log.Println("file stat error: ", err)
		return false
	}

	var size int64 = meta.Offset + int64(len(meta.Bytes))
	if size > int64(fi.Size()) {
		log.Println("file size is not enough", filename, meta.Type)
		return false
	}

	bufr := bufio.NewReader(file)

	bytesToRead := make([]byte, size)
	_, err = bufr.Read(bytesToRead)
	if err != nil {
		log.Println("Unable to read file: ", err, filename)
		return false
	}

	return bytes.Equal(meta.Bytes, bytesToRead[meta.Offset:])
}
