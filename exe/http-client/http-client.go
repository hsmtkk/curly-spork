package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/spf13/cobra"
)

var command = &cobra.Command{
	Use: "http-client",
}

var submitFileCommand = &cobra.Command{
	Use:  "submit-file file",
	Run:  submitFile,
	Args: cobra.ExactArgs(1),
}

var reportSummaryCommand = &cobra.Command{
	Use:  "report-summary sha256",
	Run:  reportSummary,
	Args: cobra.ExactArgs(1),
}

var (
	host string
	port int
)

func init() {
	command.LocalFlags().StringVar(&host, "host", "127.0.0.1", "HTTP host")
	command.LocalFlags().IntVar(&port, "port", 8000, "HTTP port")
	command.AddCommand(submitFileCommand)
	command.AddCommand(reportSummaryCommand)
}

func main() {
	if err := command.Execute(); err != nil {
		log.Fatalf("failed to execute cobra command; %v", err)
	}
}

type submitFileRequest struct {
	EncodedData string `json:"encoded-data"`
}

func submitFile(cmd *cobra.Command, args []string) {
	file := args[0]
	data, err := os.ReadFile(file)
	if err != nil {
		log.Fatalf("failed to read file; %s; %v", file, err)
	}
	encoded := base64.StdEncoding.EncodeToString(data)
	req := submitFileRequest{EncodedData: encoded}
	js, err := json.Marshal(&req)
	if err != nil {
		log.Fatalf("failed to marshal JSON; %v", err)
	}
	url := fmt.Sprintf("http://%s:%d/submit-file", host, port)
	resp, err := http.Post(url, "application/json", bytes.NewReader(js))
	if err != nil {
		log.Fatalf("failed to send POST request; %v", err)
	}
	defer resp.Body.Close()
	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("failed to read response; %v", err)
	}
	fmt.Println(string(respBytes))
}

func reportSummary(cmd *cobra.Command, args []string) {
	sha256 := args[0]
	url := fmt.Sprintf("http://%s:%d/report-summary/%s", host, port, sha256)
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("failed to send GET request; %v", err)
	}
	defer resp.Body.Close()
	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("failed to read response; %v", err)
	}
	fmt.Println(string(respBytes))
}
