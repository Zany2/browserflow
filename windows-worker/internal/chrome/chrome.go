package chrome

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func Detect() (string, error) {
	for _, candidate := range chromeCandidates() {
		if candidate == "" {
			continue
		}
		if err := ValidatePath(candidate); err == nil {
			return candidate, nil
		}
	}
	return "", errors.New("未检测到 Chrome")
}

func Validate(path string) error {
	return ValidatePath(path)
}

func ValidatePath(path string) error {
	path = strings.TrimSpace(path)
	if path == "" {
		return errors.New("Chrome path is empty")
	}
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return fmt.Errorf("%s is a directory, not an executable file", path)
	}
	if runtime.GOOS == "windows" && !strings.EqualFold(filepath.Ext(path), ".exe") {
		return fmt.Errorf("%s is not a .exe executable file", path)
	}
	return nil
}

func chromeCandidates() []string {
	var candidates []string
	if runtime.GOOS == "windows" {
		programFiles := os.Getenv("ProgramFiles")
		programFilesX86 := os.Getenv("ProgramFiles(x86)")
		localAppData := os.Getenv("LOCALAPPDATA")
		candidates = append(candidates,
			filepath.Join(programFiles, "Google", "Chrome", "Application", "chrome.exe"),
			filepath.Join(programFilesX86, "Google", "Chrome", "Application", "chrome.exe"),
			filepath.Join(localAppData, "Google", "Chrome", "Application", "chrome.exe"),
		)
	}
	candidates = append(candidates, "chrome", "google-chrome", "google-chrome-stable")
	return candidates
}
