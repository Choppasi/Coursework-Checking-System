package services

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type FileService struct {
	UploadDir string
	MaxSize   int64
}

func NewFileService(uploadDir string) *FileService {
	return &FileService{
		UploadDir: uploadDir,
		MaxSize:   10 * 1024 * 1024, // 10 MB
	}
}

func (s *FileService) SaveFile(r *http.Request, fieldName string) (string, string, error) {
	if err := r.ParseMultipartForm(s.MaxSize); err != nil {
		return "", "", fmt.Errorf("файл слишком большой (макс. 10 МБ)")
	}

	file, header, err := r.FormFile(fieldName)
	if err != nil {
		return "", "", nil // нет файла — не ошибка
	}
	defer file.Close()

	if header.Size > s.MaxSize {
		return "", "", fmt.Errorf("файл слишком большой")
	}

	if err := os.MkdirAll(s.UploadDir, 0755); err != nil {
		return "", "", err
	}

	ext := filepath.Ext(header.Filename)
	if ext == "" {
		ext = ".bin"
	}
	// разрешенные расширения
	allowed := map[string]bool{
		".pdf": true, ".doc": true, ".docx": true, ".txt": true,
		".zip": true, ".rar": true, ".7z": true,
		".jpg": true, ".jpeg": true, ".png": true, ".gif": true,
	}
	if !allowed[strings.ToLower(ext)] {
		return "", "", fmt.Errorf("недопустимый тип файла")
	}

	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", "", fmt.Errorf("file name generation failed")
	}
	filename := fmt.Sprintf("%d_%s%s", time.Now().Unix(), hex.EncodeToString(b), ext)
	path := filepath.Join(s.UploadDir, filename)

	out, err := os.Create(path)
	if err != nil {
		return "", "", fmt.Errorf("could not save file")
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		return "", "", fmt.Errorf("could not save file")
	}

	return "/uploads/" + filename, header.Filename, nil
}
