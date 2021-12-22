package hybridanalysis

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httputil"

	"go.uber.org/zap"
)

type Client struct {
	sugar   *zap.SugaredLogger
	client  *http.Client
	apiKey  string
	baseURL string
}

const (
	v2APIBaseURL = "https://www.hybrid-analysis.com/api/v2"
	win10x64Env  = "120"
)

func NewClient(sugar *zap.SugaredLogger, apiKey string) *Client {
	client := http.DefaultClient
	baseURL := v2APIBaseURL
	return &Client{sugar, client, apiKey, baseURL}
}

func NewClientForTest(sugar *zap.SugaredLogger, client *http.Client, baseURL string) *Client {
	apiKey := "test"
	return &Client{sugar, client, apiKey, baseURL}
}

func (h *Client) SubmitFile(fileName string, content []byte) ([]byte, error) {
	contentType, reqBody, err := h.createMultiPartRequest(fileName, content)
	if err != nil {
		return nil, err
	}
	respBytes, err := h.submitFile(contentType, reqBody)
	if err != nil {
		return nil, err
	}
	return respBytes, nil
}

func (h *Client) createMultiPartRequest(fileName string, content []byte) (string, []byte, error) {
	var buf bytes.Buffer
	multiWriter := multipart.NewWriter(&buf)
	fieldWriter, err := multiWriter.CreateFormField("environment_id")
	if err != nil {
		return "", nil, fmt.Errorf("failed to create form; %w", err)
	}
	fieldWriter.Write([]byte(win10x64Env))
	formWriter, err := multiWriter.CreateFormFile("file", fileName)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create form; %w", err)
	}
	if _, err := io.Copy(formWriter, bytes.NewReader(content)); err != nil {
		return "", nil, fmt.Errorf("failed to write stream; %w", err)
	}
	multiWriter.Close()
	contentType := multiWriter.FormDataContentType()
	return contentType, buf.Bytes(), nil
}

func (h *Client) submitFile(contentType string, reqBody []byte) ([]byte, error) {
	url := h.baseURL + "/submit/file"
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed make request; %w", err)
	}
	headers := map[string]string{
		"Accept":       "application/json",
		"Content-Type": contentType,
		"User-Agent":   "Falcon Sandbox",
		"api-key":      h.apiKey,
	}
	return h.doHTTPRequest(req, headers)
}

func (h *Client) ReportSummary(jobID string) ([]byte, error) {
	url := fmt.Sprintf("%s/report/%s/summary", v2APIBaseURL, jobID)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to make request; %w", err)
	}
	headers := map[string]string{
		"User-Agent": "Falcon Sandbox",
		"api-key":    h.apiKey,
	}
	return h.doHTTPRequest(req, headers)
}

func (h *Client) doHTTPRequest(req *http.Request, headers map[string]string) ([]byte, error) {
	for k, v := range headers {
		req.Header.Add(k, v)
	}

	// debug
	reqBytes, err := httputil.DumpRequest(req, true)
	if err != nil {
		return nil, fmt.Errorf("failed to dump request; %w", err)
	}
	h.sugar.Debugw("doHTTPRequest", "request", string(reqBytes))

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request; %w", err)
	}
	defer resp.Body.Close()

	// debug
	respBytes, err := httputil.DumpResponse(resp, true)
	if err != nil {
		return nil, fmt.Errorf("failed to dump response; %w", err)
	}
	h.sugar.Debugw("doHTTPRequest", "response", string(respBytes))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("non 2XX HTTP status code; %d; %s", resp.StatusCode, resp.Status)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response ; %w", err)
	}
	return body, nil
}
