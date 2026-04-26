package util

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var allowedImageExt = map[string]struct{}{
	".jpg":  {},
	".jpeg": {},
	".png":  {},
	".gif":  {},
	".webp": {},
}

var allowedImageMIME = map[string]struct{}{
	"image/jpeg": {},
	"image/png":  {},
	"image/gif":  {},
	"image/webp": {},
}

func SaveUploadedImage(fileHeader *multipart.FileHeader, baseDir string, maxUploadSize int64, pathParts ...string) (string, int64, error) {
	if fileHeader == nil {
		return "", 0, fmt.Errorf("文件不能为空")
	}

	if fileHeader.Size > maxUploadSize {
		return "", 0, fmt.Errorf("文件大小不能超过10MB")
	}

	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if _, ok := allowedImageExt[ext]; !ok {
		return "", 0, fmt.Errorf("不支持的文件扩展名")
	}

	src, err := fileHeader.Open()
	if err != nil {
		return "", 0, fmt.Errorf("打开上传文件失败: %w", err)
	}
	defer src.Close()

	header := make([]byte, 512)
	n, err := src.Read(header)
	if err != nil && err != io.EOF {
		return "", 0, fmt.Errorf("读取上传文件失败: %w", err)
	}

	mimeType := http.DetectContentType(header[:n])
	if _, ok := allowedImageMIME[mimeType]; !ok {
		return "", 0, fmt.Errorf("不支持的文件MIME类型")
	}

	if _, err := src.Seek(0, io.SeekStart); err != nil {
		return "", 0, fmt.Errorf("重置文件读取位置失败: %w", err)
	}

	targetDir := filepath.Join(append([]string{baseDir}, pathParts...)...)
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return "", 0, fmt.Errorf("创建目录失败: %w", err)
	}

	fileName, err := uniqueFileName(ext)
	if err != nil {
		return "", 0, fmt.Errorf("生成文件名失败: %w", err)
	}

	targetPath := filepath.Join(targetDir, fileName)
	dst, err := os.Create(targetPath)
	if err != nil {
		return "", 0, fmt.Errorf("创建目标文件失败: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return "", 0, fmt.Errorf("保存文件失败: %w", err)
	}

	relativePath := "/" + filepath.ToSlash(targetPath)
	return relativePath, fileHeader.Size, nil
}

func uniqueFileName(ext string) (string, error) {
	randomBytes := make([]byte, 8)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	return fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), hex.EncodeToString(randomBytes), ext), nil
}
