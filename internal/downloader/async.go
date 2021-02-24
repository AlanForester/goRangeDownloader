package downloader

import (
	"github.com/AlexCollin/goRangeDownloader/internal/cmd"
	"github.com/AlexCollin/goRangeDownloader/internal/errutil"
	"log"
	"os"
	"sync/atomic"
)

func AsyncDownload(filepath string, url string, size int64, worksCount *int64, maxRepeats *int64) (err error) {
	log.Printf("Save to: %s\n", filepath)
	f, err := os.OpenFile(filepath, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return
	}
	defer f.Close()

	wc, wb := uint32(0), uint64(0)
	var worker = worker{
		Url:       url,
		File:      f,
		Count:     *worksCount,
		TotalSize: size,

		writeCount: &wc,
		writeBytes: &wb,
	}

	var start, end int64
	var partialSize = size / *worksCount
	log.Println("Part size on goroutine:", partialSize)
	for num := int64(0); num < worker.Count; num++ {
		start = num * partialSize
		end += partialSize
		if end >= size {
			end = size
		} else if start > 0 {
			start += 1
		}

		// If surplus exists and division by worksCount != 0
		if num == (worker.Count - 1) {
			end += size % *worksCount
		}

		log.Printf("Goroutine №%d: Start download from '%d' to '%d' bytes of file", num, start, end)

		worker.SyncWG.Add(1)
		go worker.writeSlice(num, start, end, 0, *maxRepeats)
		start = end
	}

	worker.SyncWG.Wait()

	writesCount := atomic.LoadUint32(worker.writeCount)
	log.Println("Writes count:", writesCount)

	writesByte := atomic.LoadUint64(worker.writeBytes)
	log.Println("Check write total bytes:", writesByte)

	if uint64(size) != writesByte {
		log.Printf("Error: %v", errutil.FileSizeIsNotEqualWrittenBytesError.Error())
		if cmd.DownloadAgainQuestion() {
			return AsyncDownload(filepath, url, size, worksCount, maxRepeats)
		}
	}

	return
}
