//go:build !windows

package browser

func selectBrowserBinPath() (string, bool) {
	return "", false
}

func selectUserDataDir() (string, bool) {
	return "", false
}
