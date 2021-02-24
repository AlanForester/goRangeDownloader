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
	"sync/atomic"
	"time"
)

type worker struct {
	sync.Mutex

	Url       string
	File      *os.File
	Count     int64
	SyncWG    sync.WaitGroup
	TotalSize int64

	WriteChan chan Chunk

	writeCount *uint32
	writeBytes *uint64
}

func (w *worker) writeChunk(chunk Chunk) {

	nw, err := w.File.WriteAt(chunk.Data, chunk.Start)

	if err != nil {
		log.Fatalf("Slice %d occured error: %s.\n", chunk.FlowNum, err.Error())
	}

	atomic.AddUint64(w.writeBytes, uint64(nw))

	log.Printf("Goroutine №%d: writing %v-%v bytes of file", chunk.FlowNum, chunk.Start, chunk.End)
}

func (w *worker) downloadSlice(sliceNum int64, start int64, end int64, repeat int64, maxRepeats int64) {
	body, size, err := w.getSliceData(start, end)
	if err != nil {
		log.Printf("Slice %d request error: %s\n", sliceNum, err.Error())
		if maxRepeats > repeat {
			repeat += 1
			time.AfterFunc(5*time.Second, func() {
				w.downloadSlice(sliceNum, start, end, repeat, maxRepeats)
			})
			return
		} else {
			log.Fatalf("Error: %v\n", OneOrManyChunksDontLoaded)
		}
	}

	defer body.Close()
	defer w.SyncWG.Done()

	// Write with 4096 bytes block size
	bs := int64(4096)
	log.Printf("Goroutine №%d: Set write block size %v ", sliceNum, bs)
	buf := make([]byte, bs)
	for {
		nr, er := body.Read(buf)
		if nr > 0 {
			start = int64(nr) + start

			log.Printf("Goroutine №%d: send %v-%v bytes on write", sliceNum, start-bs, start)
			w.WriteChan <- Chunk{Start: start - bs, End: start, Data: buf[0:nr], Size: int(size), FlowNum: sliceNum}

			atomic.AddUint32(w.writeCount, 1)

		}
		if er != nil {
			if er.Error() == "EOF" {
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
