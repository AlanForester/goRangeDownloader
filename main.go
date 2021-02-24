package main

import (
	"flag"
	"fmt"
	"github.com/AlexCollin/goRangeDownloader/internal/downloader"
	. "github.com/AlexCollin/goRangeDownloader/internal/errutil"
	"github.com/AlexCollin/goRangeDownloader/internal/urlutil"
	"log"
	"time"
)

func main() {

	var workerCount = flag.Int64("c", 5, "goroutines count")
	flag.Parse()

	now := time.Now().UTC()

	downloadUrl := "http://i.imgur.com/z4d4kWk.jpg"
	log.Println("Url:", downloadUrl)
	fileSize, err := urlutil.GetSizeAndCheckRangeSupport(downloadUrl)

	log.Printf("Size: %d bytes\n", fileSize)
	fName, err := urlutil.GetFileName(downloadUrl)
	if err != nil {
		log.Printf("Could not get file name. Using default")
		fName = "default"
	}
	var filePath = fmt.Sprintf("./downloads/%s", fName)

	if err == nil {
		HandleError(downloader.AsyncDownload(filePath, downloadUrl, fileSize, workerCount))
	} else {
		// TODO: check on write bytes with repeat download as in AsyncDownload
		HandleError(downloader.SyncDownload(filePath, downloadUrl))
	}

	log.Println("Elapsed time:", time.Since(now))
	log.Println("Finish!")
}
