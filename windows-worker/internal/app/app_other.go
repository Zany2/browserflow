//go:build !windows

package app

import "fmt"

func Run() error {
	return fmt.Errorf("BrowserFlow Windows Worker GUI only supports Windows")
}
