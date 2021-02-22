package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/cheggaaa/pb"
)

type Worker struct {
	Url       string
	File      *os.File
	Count     int64
	SyncWG    sync.WaitGroup
	TotalSize int64
	State
}

type State struct {
	Pool *pb.Pool
	Bars []*pb.ProgressBar
}

func main() {

	var workerCount = flag.Int64("c", 5, "goroutines count")
	flag.Parse()

	now := time.Now().UTC()

	downloadUrl  := "https://ivbb.ru/"
	log.Println("Url:", downloadUrl)
	fileSize, err := getSizeAndCheckRangeSupport(downloadUrl)
	if fileSize == 0 {
		handleError(err)
	}

	log.Printf("Size: %d bytes\n", fileSize)
	fName := getFileName(downloadUrl)
	if fName == "" {
		fName = "default.null"
	}
	var filePath = fmt.Sprintf("./downloads/%s", fName)

	if err == nil {
		log.Printf("Save to: %s\n", filePath)
		f, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDWR, 0666)
		handleError(err)
		defer f.Close()

		var worker = Worker{
			Url:       downloadUrl,
			File:      f,
			Count:     *workerCount,
			TotalSize: fileSize,
		}

		var start, end int64
		var partialSize = int64(fileSize / *workerCount)

		for num := int64(0); num < worker.Count; num++ {
			bar := pb.New(0).Prefix(fmt.Sprintf("Slice %d  0%% ", num))
			bar.ShowSpeed = true
			bar.SetMaxWidth(100)
			bar.SetUnits(pb.U_BYTES_DEC)
			bar.SetRefreshRate(time.Second)
			bar.ShowPercent = true
			worker.State.Bars = append(worker.State.Bars, bar)

			if num == worker.Count {
				end = fileSize
			} else {
				end = start + partialSize
			}

			worker.SyncWG.Add(1)
			go worker.writeSlice(num, start, end-1)
			start = end
		}
		worker.State.Pool, err = pb.StartPool(worker.State.Bars...)
		handleError(err)
		worker.SyncWG.Wait()
		worker.State.Pool.Stop()
	} else {
		err = downloadFile(filePath, downloadUrl)
	}

	handleError(err)


	log.Println("Elapsed time:", time.Since(now))
	log.Println("Finish!")
}

func downloadFile(filepath string, url string) error {

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func (w *Worker) writeSlice(sliceNum int64, start int64, end int64) {
	var written int64
	body, size, err := w.getSliceData(start, end)
	if err != nil {
		log.Fatalf("Slice %d request error: %s\n", sliceNum, err.Error())
	}
	defer body.Close()
	defer w.Bars[sliceNum].Finish()
	defer w.SyncWG.Done()

	w.Bars[sliceNum].Total = size

	percentFlag := map[int64]bool{}

	buf := make([]byte, 4*1024)
	for {
		nr, er := body.Read(buf)
		if nr > 0 {
			nw, err := w.File.WriteAt(buf[0:nr], start)
			if err != nil {
				log.Fatalf("Slice %d occured error: %s.\n", sliceNum, err.Error())
			}
			if nr != nw {
				log.Fatalf("Slice %d occured error of short writiing.\n", sliceNum)
			}

			start = int64(nw) + start
			if nw > 0 {
				written += int64(nw)
			}

			w.Bars[int(sliceNum)].Set64(written)

			p := int64(float32(written) / float32(size) * 100)
			_, flagged := percentFlag[p]
			if !flagged {
				percentFlag[p] = true
				w.Bars[int(sliceNum)].Prefix(fmt.Sprintf("Slice %d  %d%% ", sliceNum, p))
			}
		}
		if er != nil {
			if er.Error() == "EOF" {
				if size == written {
				} else {
					handleError(errors.New(fmt.Sprintf("Slice %d unfinished.\n", sliceNum)))
				}
				break
			}
			handleError(errors.New(fmt.Sprintf("Slice %d occured error: %s\n", sliceNum, er.Error())))
		}
	}
}

func (w *Worker) getSliceData(start int64, end int64) (io.ReadCloser, int64, error) {
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

func getSizeAndCheckRangeSupport(url string) (size int64, err error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	res, err := client.Do(req)
	if err != nil {
		return
	}

	header := res.Header
	log.Printf("%v+", header)
	acceptRanges, supported := header["Accept-Ranges"]
	log.Printf("Response header range: %v\n",supported)
	if !supported {
		err = errors.New("Doesn't support header 'Accept-Ranges'. ")
	} else if supported && acceptRanges[0] != "bytes" {
		err = errors.New("Exists header 'Accept-Ranges', but value is not 'bytes'. ")
	}
	if _, ok := header["Content-Length"];!ok {
		err = errors.New("Header 'Content-Length' is empty. Download is nothing ")
		return
	}
	size, err = strconv.ParseInt(header["Content-Length"][0], 10, 64)
	return
}

func getFileName(downloadUrl string) string {
	urlStruct, err := url.Parse(downloadUrl)
	handleError(err)
	return filepath.Base(urlStruct.Path)
}

func handleError(err error) {
	if err != nil {
		log.Println("Error:", err)
		os.Exit(1)
	}
}
