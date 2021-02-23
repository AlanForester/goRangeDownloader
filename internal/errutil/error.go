package errutil

import (
	"log"
	"os"
)

func HandleError(err error) {
	if err != nil {
		log.Println("Error:", err)
		os.Exit(1)
	}
}
