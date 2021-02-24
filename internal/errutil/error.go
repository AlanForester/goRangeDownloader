package errutil

import (
	"errors"
	"log"
	"os"
)

var FileSizeIsNotEqualWrittenBytesError = errors.New("File size is not equal written bytes ")

func HandleError(err error) {
	if err != nil {
		log.Println("Error:", err)
		os.Exit(1)
	}
}
