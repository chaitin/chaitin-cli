package monkeyscan

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func prepareScanArchive(opts scanOptions) (string, func(), error) {
	if strings.TrimSpace(opts.File) != "" {
		if err := validateArchive(opts.File); err != nil {
			return "", func() {}, err
		}
		if err := validateArchiveFileBodySize(opts.File); err != nil {
			return "", func() {}, err
		}
		return opts.File, func() {}, nil
	}
	source := strings.TrimSpace(opts.Path)
	if source == "" {
		source = "."
	}
	info, err := os.Stat(source)
	if err != nil {
		return "", func() {}, err
	}
	if !info.IsDir() {
		return "", func() {}, fmt.Errorf("--path 必须是目录: %s", source)
	}
	tmp, err := os.CreateTemp("", "monkeyscan-*.zip")
	if err != nil {
		return "", func() {}, err
	}
	tmpPath := tmp.Name()
	if err := tmp.Close(); err != nil {
		return "", func() {}, err
	}
	cleanup := func() { _ = os.Remove(tmpPath) }
	if err := zipDirectory(source, tmpPath); err != nil {
		cleanup()
		return "", func() {}, err
	}
	if err := validateArchive(tmpPath); err != nil {
		cleanup()
		return "", func() {}, err
	}
	if err := validateArchiveFileBodySize(tmpPath); err != nil {
		cleanup()
		return "", func() {}, err
	}
	return tmpPath, cleanup, nil
}

func zipDirectory(sourceDir, outputPath string) error {
	out, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer out.Close()
	zw := zip.NewWriter(out)
	defer zw.Close()
	root, err := filepath.Abs(sourceDir)
	if err != nil {
		return err
	}
	return filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		name := entry.Name()
		if entry.IsDir() && (name == ".git" || name == ".monkeyscan") {
			return filepath.SkipDir
		}
		if entry.IsDir() {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = rel
		header.Method = zip.Deflate
		writer, err := zw.CreateHeader(header)
		if err != nil {
			return err
		}
		in, err := os.Open(path)
		if err != nil {
			return err
		}
		_, copyErr := io.Copy(writer, in)
		closeErr := in.Close()
		if copyErr != nil {
			return copyErr
		}
		return closeErr
	})
}

func validateArchive(filePath string) error {
	archiveType, err := archiveTypeFromPath(filePath)
	if err != nil {
		return err
	}
	if archiveType == "zip" {
		reader, err := zip.OpenReader(filePath)
		if err != nil {
			return fmt.Errorf("无效的 zip 文件: %w", err)
		}
		defer reader.Close()
		return validateZipArchive(&reader.Reader)
	}
	return validateTarGzArchive(filePath)
}

func validateArchiveFileBodySize(filePath string) error {
	info, err := os.Stat(filePath)
	if err != nil {
		return err
	}
	if info.Size() > maxArchiveFileBody {
		return fmt.Errorf("压缩包当前大小是 %.1f M，超过 100M 限制", float64(info.Size())/(1024*1024))
	}
	return nil
}

func archiveTypeFromPath(filePath string) (string, error) {
	lower := strings.ToLower(filePath)
	switch {
	case strings.HasSuffix(lower, ".tar.gz"), strings.HasSuffix(lower, ".tgz"):
		return "targz", nil
	case filepath.Ext(lower) == ".zip":
		return "zip", nil
	default:
		return "", errors.New("仅支持 .zip、.tar.gz、.tgz 压缩包")
	}
}

func validateZipArchive(reader *zip.Reader) error {
	var total int64
	for i, file := range reader.File {
		if i+1 > maxArchiveEntries {
			return fmt.Errorf("压缩包条目数量超过限制: %d > %d", i+1, maxArchiveEntries)
		}
		if err := validateArchiveEntry(file.Name, int64(file.UncompressedSize64), file.FileInfo().IsDir()); err != nil {
			return err
		}
		if !file.FileInfo().IsDir() {
			total += int64(file.UncompressedSize64)
			if total > maxArchiveSize {
				return fmt.Errorf("压缩包总解压体积超过限制: %d > %d", total, maxArchiveSize)
			}
		}
	}
	return nil
}

func validateTarGzArchive(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	gzr, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("无效的 gzip 文件: %w", err)
	}
	defer gzr.Close()
	tr := tar.NewReader(gzr)
	var total int64
	count := 0
	for {
		header, err := tr.Next()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("读取 tar 条目失败: %w", err)
		}
		count++
		if count > maxArchiveEntries {
			return fmt.Errorf("压缩包条目数量超过限制: %d > %d", count, maxArchiveEntries)
		}
		isDir := header.Typeflag == tar.TypeDir
		if err := validateArchiveEntry(header.Name, header.Size, isDir); err != nil {
			return err
		}
		if header.Typeflag == tar.TypeReg {
			total += header.Size
			if total > maxArchiveSize {
				return fmt.Errorf("压缩包总解压体积超过限制: %d > %d", total, maxArchiveSize)
			}
		}
	}
}

func validateArchiveEntry(name string, size int64, isDir bool) error {
	clean := filepath.Clean(name)
	if filepath.IsAbs(clean) {
		return fmt.Errorf("不允许使用绝对路径: %s", name)
	}
	if strings.HasPrefix(clean, "..") || strings.Contains(clean, "../") || strings.Contains(clean, "..\\") {
		return fmt.Errorf("检测到路径穿越攻击: %s", name)
	}
	depth := strings.Count(name, "/") + strings.Count(name, "\\")
	if depth > maxArchiveDepth {
		return fmt.Errorf("文件路径深度超过限制: %s (深度: %d > %d)", name, depth, maxArchiveDepth)
	}
	if !isDir && size > maxArchiveFileSize {
		return fmt.Errorf("单个文件大小超过限制: %s (%d > %d)", name, size, maxArchiveFileSize)
	}
	return nil
}
