package urlutil

import (
	"errors"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
)

func GetSizeAndCheckRangeSupport(url string) (size int64, err error) {
	client := &http.Client{}
	req, err := http.NewRequest("HEAD", url, nil)
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
	log.Printf("Response header range: %v\n", supported)
	if !supported {
		err = errors.New("Doesn't support header 'Accept-Ranges'. ")
	} else if supported && acceptRanges[0] != "bytes" {
		err = errors.New("Exists header 'Accept-Ranges', but value is not 'bytes'. ")
	}
	if _, ok := header["Content-Length"]; !ok {
		err = errors.New("Header 'Content-Length' is empty. ")
	} else {
		size, err = strconv.ParseInt(header["Content-Length"][0], 10, 64)
	}
	return
}

func GetFileName(downloadUrl string) (string, error) {
	urlStruct, err := url.Parse(downloadUrl)
	if err != nil {
		return "", err
	}
	return filepath.Base(urlStruct.Path), nil
}
