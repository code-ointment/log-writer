package log_file

import (
	"compress/gzip"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	logwriter "github.com/code-ointment/log-writer"
)

const (
	logPermissions os.FileMode = 0644
)

type zippedFile struct {
	FileName string
	Modtime  int64
}

type LogFileWriter struct {
	FileName     string
	Generations  int
	Size         int
	fd           *os.File
	bytesWritten int
	zipRequests  chan string
	zippedFiles  []*zippedFile
	zipWait      sync.WaitGroup
	gzipper      *gzip.Writer
}

/*
* fileName - file to write to.
* generations - number of compressed logs to retain.
* sz - number of bytes written before compressing.
 */
func NewLogFileWriter(fileName string, generations int, sz int) *LogFileWriter {

	absFname, err := filepath.Abs(fileName)
	if err != nil {
		fmt.Printf("invalid path %s : %v\n", fileName, err)
		return nil
	}

	lw := LogFileWriter{
		FileName:    absFname,
		Generations: generations,
		Size:        sz,
		// Caller will block if the zipper is really behind.
		zipRequests: make(chan string, generations*10),
		gzipper:     gzip.NewWriter(nil),
	}

	lw.findZipped()
	go lw.zipper()
	logwriter.Register(&lw)

	return &lw
}

func (lw *LogFileWriter) Close() {
	lw.zipWait.Wait()
}

func (lw *LogFileWriter) Write(p []byte) (int, error) {

	var err error

	if lw.fd == nil {
		lw.fd, err = os.OpenFile(lw.FileName,
			os.O_CREATE|os.O_APPEND|os.O_WRONLY, logPermissions)
		if err != nil {
			fmt.Printf("failed opening %s\n", lw.FileName)
			return 0, err
		}
	}

	n, err := lw.fd.Write(p)
	if err != nil {
		fmt.Printf("failed writing %s\n", lw.FileName)
	}
	lw.bytesWritten += n

	// We will exceed lw.Size by the length of the buffer. I mean to do this.
	if lw.bytesWritten >= lw.Size {

		lw.fd.Close()
		lw.bytesWritten = 0
		lw.fd = nil
		lw.archive()
	}
	return n, err
}

/*
* Create a tracking entry.
 */
func (lw *LogFileWriter) zippedFileFromName(fname string) *zippedFile {

	st, err := os.Lstat(fname)
	if err != nil {
		fmt.Printf("failed creating entry for %s : %v\n", fname, err)
		return nil
	}

	cf := zippedFile{
		FileName: fname,
		Modtime:  st.ModTime().UnixNano(),
	}
	return &cf
}

/*
* Read the directory looking for files we've gzipped
 */
func (lw *LogFileWriter) findZipped() {
	pat := lw.FileName + ".*.gz"
	matches, err := filepath.Glob(pat)

	if err != nil {
		fmt.Printf("failed globbing: %v\n", err)
	}

	mlen := len(matches)
	if mlen == 0 {
		return
	}

	for _, fname := range matches {

		cf := lw.zippedFileFromName(fname)
		if cf == nil {
			continue
		}
		lw.zippedFiles = append(lw.zippedFiles, cf)
	}
}

/*
* If we haven't created all generations, return the next open id.
* Otherwise -1.
 */
func (lw *LogFileWriter) getNextNewId() int {

	mlen := len(lw.zippedFiles)
	if mlen == 0 {
		return 1
	}

	if mlen < lw.Generations {
		return mlen + 1
	}
	return -1
}

/*
* Find the oldest file and update it's time stamp.
 */
func (lw *LogFileWriter) oldestZippedFile() string {

	var index int

	least := lw.zippedFiles[0].Modtime
	for i, zf := range lw.zippedFiles {
		if least > zf.Modtime {
			index = i
			least = zf.Modtime
		}
	}

	lw.zippedFiles[index].Modtime = time.Now().UnixNano()
	return lw.zippedFiles[index].FileName
}

/*
* New entry in zippedFile tracking.
 */
func (lw *LogFileWriter) addZippedFile(fname string) {

	zf := zippedFile{
		FileName: fname,
		Modtime:  time.Now().UnixNano(),
	}

	lw.zippedFiles = append(lw.zippedFiles, &zf)
}

/*
* Get a suitable zip file name to compress 'fname' to.
 */
func (lw *LogFileWriter) getZipFileName(fname string) string {

	fields := strings.Split(fname, ".")
	flen := len(fields)

	if flen < 2 {
		fmt.Printf("bad file name %s\n", fname)
		return fname
	}

	// Drop trailing random number
	fields = fields[:flen-1]

	id := lw.getNextNewId()
	if id != -1 {
		fields = append(fields, strconv.Itoa(id), "gz")
		fname := strings.Join(fields, ".")
		lw.addZippedFile(fname)
		return fname
	}

	return lw.oldestZippedFile()
}

/*
* Generate a random file name and request a compression.
 */
func (lw *LogFileWriter) archive() {

	tmpId := rand.Int()
	fname := fmt.Sprintf("%s.%d", lw.FileName, tmpId)

	if _, err := os.Stat(fname); err == nil {
		os.Remove(fname) // unlikely but just in case.
	}
	os.Rename(lw.FileName, fname)
	lw.zipWait.Add(1)
	lw.zipRequests <- fname
}

/*
* Pull the request an zip it
 */
func (lw *LogFileWriter) zipper() {

	for {
		fileName := <-lw.zipRequests
		outFile := lw.getZipFileName(fileName)
		fmt.Printf("src: %s target: %s\n", fileName, outFile)

		lw.doZip(fileName, outFile)
		lw.zipWait.Done()

		os.Remove(fileName)
	}
}

func (lw *LogFileWriter) doZip(src string, dest string) {

	in, err := os.Open(src)
	if err != nil {
		fmt.Printf("error opening input %s : %v\n", src, err)
		return
	}
	defer in.Close()

	out, err := os.OpenFile(dest, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0666)
	if err != nil {
		fmt.Printf("error opening output %s : %v\n", src, err)
		return
	}
	defer out.Close()

	lw.gzipper.Reset(out)
	io.Copy(lw.gzipper, in)
	defer lw.gzipper.Close()
}
