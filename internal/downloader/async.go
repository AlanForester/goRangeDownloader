package downloader

import (
	"log"
	"os"
)

func AsyncDownload(filepath string, url string, size int64, worksCount *int64) (err error) {
	log.Printf("Save to: %s\n", filepath)
	f, err := os.OpenFile(filepath, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return
	}
	defer f.Close()

	var worker = worker{
		Url:       url,
		File:      f,
		Count:     *worksCount,
		TotalSize: size,
	}

	var start, end int64
	var partialSize = size / *worksCount

	for num := int64(0); num < worker.Count; num++ {
		if num == worker.Count {
			end = size
		} else {
			end = start + partialSize
		}

		worker.SyncWG.Add(1)
		go worker.writeSlice(num, start, end-1)
		start = end
	}

	worker.SyncWG.Wait()

	return
}
