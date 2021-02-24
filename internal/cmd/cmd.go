package cmd

import (
	"fmt"
)

func DownloadAgainQuestion() bool {
	var answer string
	fmt.Print("Try download file again?[yn] ")
	_, _ = fmt.Scan(&answer)
	if answer == "y" {
		return true
	}
	return false
}
