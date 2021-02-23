package downloader

import (
	"errors"
	"fmt"
	. "github.com/AlexCollin/goRangeDownloader/internal/errutil"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
)

type worker struct {
	Url       string
	File      *os.File
	Count     int64
	SyncWG    sync.WaitGroup
	TotalSize int64
}

func (w *worker) writeSlice(sliceNum int64, start int64, end int64) {
	var written int64
	body, size, err := w.getSliceData(start, end)
	if err != nil {
		log.Fatalf("Slice %d request error: %s\n", sliceNum, err.Error())
	}
	defer body.Close()
	defer w.SyncWG.Done()

	bufSize := int64(4 * 1024)
	buf := make([]byte, bufSize)
	for {
		nr, er := body.Read(buf)
		if nr > 0 {
			nw, err := w.File.WriteAt(buf[0:nr], start)
			if err != nil {
				log.Fatalf("Slice %d occured error: %s.\n", sliceNum, err.Error())
			}
			if nr != nw {
				log.Fatalf("Slice %d occured error of short writing.\n", sliceNum)
			}

			start = int64(nw) + start
			if nw > 0 {
				written += int64(nw)
			}

			log.Printf("Goroutine %d: writing %d bytes at %v byte of file", sliceNum+1, bufSize, start-bufSize)

		}
		if er != nil {
			if er.Error() == "EOF" {
				if size == written {
				} else {
					HandleError(errors.New(fmt.Sprintf("Slice write %d unfinished.\n", sliceNum)))
				}
				break
			}
			HandleError(errors.New(fmt.Sprintf("Slice %d occured error: %s\n", sliceNum, er.Error())))
		}
	}
}

func (w *worker) getSliceData(start int64, end int64) (io.ReadCloser, int64, error) {
	var client http.Client
	req, err := http.NewRequest("GET", w.Url, nil)
	if err != nil {
		return nil, 0, err
	}

	req.Header.Add("Range", fmt.Sprintf("bytes=%d-%d", start, end))
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	size, err := strconv.ParseInt(resp.Header["Content-Length"][0], 10, 64)
	return resp.Body, size, err
}
