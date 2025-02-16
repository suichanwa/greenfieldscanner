package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

type Client struct {
	BaseURL string
	Token   string
}

type AuthResponse struct {
	Token string `json:"token"`
}

type FileInfo struct {
	ID           uint   `json:"id"`
	Name         string `json:"name"`
	Size         int64  `json:"size"`
	LastModified string `json:"last_modified"`
	Hash         string `json:"hash"`
}

func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL: baseURL,
	}
}

func (c *Client) sendRequest(method, path string, payload interface{}, result interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	url := c.BaseURL + path
	req, err := http.NewRequest(method, url, bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("request failed: %s", body)
	}

	return json.NewDecoder(resp.Body).Decode(result)
}

func (c *Client) Login(username, password string) error {
	payload := map[string]string{
		"username": username,
		"password": password,
	}

	var response AuthResponse
	err := c.sendRequest("POST", "/api/v1/login", payload, &response)
	if err != nil {
		return err
	}

	c.Token = response.Token
	return nil
}

func (c *Client) UploadFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return err
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return err
	}
	writer.Close()

	req, err := http.NewRequest("POST", c.BaseURL+"/api/v1/upload", body)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+c.Token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("upload failed: %d", resp.StatusCode)
	}

	return nil
}

func (c *Client) ListFiles() ([]FileInfo, error) {
	var resp struct {
		Files []FileInfo `json:"files"`
	}
	err := c.sendRequest("GET", "/api/v1/files", nil, &resp)
	if err != nil {
		return nil, err
	}
	return resp.Files, nil
}

func (c *Client) Sync() error {
	var resp map[string]interface{}
	return c.sendRequest("POST", "/api/v1/sync", nil, &resp)
}

// In desktop/api/client.go

func (c *Client) DownloadFile(fileID string, fileName string) error {
	req, err := http.NewRequest("GET", c.BaseURL+"/api/v1/files/"+fileID+"/download", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: %d", resp.StatusCode)
	}

	outFile, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, resp.Body)
	return err
}

func (c *Client) DeleteFile(fileID string) error {
	return c.sendRequest("DELETE", "/api/v1/files/"+fileID, nil, nil)
}
