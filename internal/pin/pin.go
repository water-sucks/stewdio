package pin_utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"stewdio/internal/config"
	"stewdio/internal/refs"
)

func Push(path string, remote config.Remote, version string) error {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	metadataBytes, err := json.Marshal(map[string]string{
		"version": version,
	})
	if err != nil {
		return fmt.Errorf("failed to encode metadata: %w", err)
	}

	if err := writer.WriteField("meta", string(metadataBytes)); err != nil {
		return fmt.Errorf("failed to write meta field: %w", err)
	}

	filePath := filepath.Join(path, ".stew", "objects", version, refs.ObjectTarName)
	if _, err := writer.CreateFormFile("file", filepath.Base(filePath)); err != nil {
		return fmt.Errorf("failed to create form file: %w", err)
	}

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer func() { _ = file.Close() }()

	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err := io.Copy(part, file); err != nil {
		return fmt.Errorf("failed to copy file data: %w", err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close multipart writer: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/projects/%s/pins", remote.Server, remote.Project)

	req, err := http.NewRequest("POST", url, &buf)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("upload failed: %s\n%s", res.Status, string(body))
	}

	return nil
}
