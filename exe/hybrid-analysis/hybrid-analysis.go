package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/hsmtkk/curly-spork/env"
	"github.com/hsmtkk/curly-spork/hybridanalysis"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var command = &cobra.Command{
	Use: "hybrid-analysis",
}

var submitFileCommand = &cobra.Command{
	Use:  "submit-file file",
	Run:  submitFile,
	Args: cobra.ExactArgs(1),
}

var reportSummaryCommand = &cobra.Command{
	Use:  "report-summary jobID",
	Run:  reportSummary,
	Args: cobra.ExactArgs(1),
}

func init() {
	command.AddCommand(submitFileCommand)
	command.AddCommand(reportSummaryCommand)
}

func main() {
	if err := command.Execute(); err != nil {
		log.Fatalf("cobra command failed; %v", err)
	}
}

func submitFile(cmd *cobra.Command, args []string) {
	filePath := args[0]
	apiKey := env.RequiredString("API_KEY")

	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("failed to read file; %s; %v", filePath, err)
	}

	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("failed to initialize zap logger; %v", err)
	}
	defer logger.Sync()
	sugar := logger.Sugar()

	respBytes, err := hybridanalysis.NewClient(sugar, apiKey).SubmitFile(filepath.Base(filePath), data)
	if err != nil {
		log.Fatalf("failed to submit file; %v", err)
	}
	jobID, err := hybridanalysis.NewParser(sugar).ParseSubmitFile(respBytes)
	if err != nil {
		log.Fatalf("failed to parse; %v", err)
	}
	fmt.Println(jobID)
}

func reportSummary(cmd *cobra.Command, args []string) {
	jobID := args[0]
	apiKey := env.RequiredString("API_KEY")

	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("failed to initialize zap logger; %v", err)
	}
	defer logger.Sync()
	sugar := logger.Sugar()

	respBytes, err := hybridanalysis.NewClient(sugar, apiKey).ReportSummary(jobID)
	if err != nil {
		log.Fatalf("failed to get report; %v", err)
	}
	fmt.Println(string(respBytes))
}
