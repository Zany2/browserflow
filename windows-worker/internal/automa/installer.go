package automa

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func InstallFromZip(zipPath, targetDir string) error {
	zipPath = strings.TrimSpace(zipPath)
	targetDir = strings.TrimSpace(targetDir)
	if zipPath == "" {
		return errors.New("请选择 Automa 安装包")
	}
	if targetDir == "" {
		return errors.New("Automa 安装目录为空")
	}

	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("打开 Automa 安装包失败：%w", err)
	}
	defer reader.Close()

	parentDir := filepath.Dir(targetDir)
	tmpDir := targetDir + ".tmp"
	if err = os.MkdirAll(parentDir, 0o755); err != nil {
		return err
	}
	_ = os.RemoveAll(tmpDir)
	if err = os.MkdirAll(tmpDir, 0o755); err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	for _, file := range reader.File {
		if err = extractFile(file, tmpDir); err != nil {
			return err
		}
	}

	if _, err = os.Stat(filepath.Join(tmpDir, "manifest.json")); err != nil {
		return errors.New("安装包格式不正确，未找到 manifest.json")
	}

	if err = os.RemoveAll(targetDir); err != nil {
		return err
	}
	if err = os.Rename(tmpDir, targetDir); err != nil {
		return err
	}
	return nil
}

func extractFile(file *zip.File, targetDir string) error {
	cleanName := filepath.Clean(file.Name)
	if cleanName == "." || strings.HasPrefix(cleanName, ".."+string(os.PathSeparator)) || filepath.IsAbs(cleanName) {
		return fmt.Errorf("安装包包含非法路径：%s", file.Name)
	}

	targetPath := filepath.Join(targetDir, cleanName)
	cleanTargetDir, err := filepath.Abs(targetDir)
	if err != nil {
		return err
	}
	cleanTargetPath, err := filepath.Abs(targetPath)
	if err != nil {
		return err
	}
	if cleanTargetPath != cleanTargetDir && !strings.HasPrefix(cleanTargetPath, cleanTargetDir+string(os.PathSeparator)) {
		return fmt.Errorf("安装包包含非法路径：%s", file.Name)
	}

	if file.FileInfo().IsDir() {
		return os.MkdirAll(targetPath, 0o755)
	}
	if err = os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return err
	}

	source, err := file.Open()
	if err != nil {
		return err
	}
	defer source.Close()

	target, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.FileInfo().Mode())
	if err != nil {
		return err
	}
	defer target.Close()

	_, err = io.Copy(target, source)
	return err
}
