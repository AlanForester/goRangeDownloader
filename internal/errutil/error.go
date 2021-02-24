package errutil

import (
	"errors"
	"log"
	"os"
)

var FileSizeIsNotEqualWrittenBytesError = errors.New("File size is not equal written bytes ")
var OneOrManyChunksDontLoaded = errors.New("One or many chunks dont loaded ")

func HandleError(err error) {
	if err != nil {
		log.Println("Error:", err)
		os.Exit(1)
	}
}
