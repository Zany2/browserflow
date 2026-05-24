//go:build !windows

package winui

import "log"

func MessageBox(title string, message string) {
	log.Printf("%s: %s\n", title, message)
}
