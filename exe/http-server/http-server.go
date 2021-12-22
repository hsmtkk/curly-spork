package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"

	"github.com/hsmtkk/curly-spork/env"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

func main() {
	httpPort := env.RequiredInt("HTTP_PORT")

	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("failed to initialize zap logger; %v", err)
	}
	defer logger.Sync()
	sugar := logger.Sugar()

	hdl := &handler{sugar}

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.POST("/submit-file", hdl.SubmitFile)
	e.GET("/report-summary/:sha256", hdl.ReportSummary)
	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", httpPort)))
}

type handler struct {
	sugar *zap.SugaredLogger
}

type submitFileRequest struct {
	EncodedData string `json:"encoded-data"`
}

func (h *handler) SubmitFile(ctx echo.Context) error {
	h.sugar.Info("submit-file")
	req := new(submitFileRequest)
	if err := ctx.Bind(req); err != nil {
		return fmt.Errorf("failed to bind; %w", err)
	}
	data, err := base64.StdEncoding.DecodeString(req.EncodedData)
	if err != nil {
		return fmt.Errorf("failed to decode with Base64; %w", err)
	}
	sha256 := calcSHA256(data)
	// TODO: write content to the database
	// TODO: write sha256 to the queue
	return ctx.String(http.StatusOK, sha256)
}

func (h *handler) ReportSummary(ctx echo.Context) error {
	sha256 := ctx.Param("sha256")
	h.sugar.Infow("report-summary", "sha256", sha256)
	// TODO: query database
	return ctx.String(http.StatusOK, sha256)
}

func calcSHA256(data []byte) string {
	hash := sha256.Sum256(data)
	bs := hash[:]
	return hex.EncodeToString(bs)
}
