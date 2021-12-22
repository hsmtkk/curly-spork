package main

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"

	"github.com/go-redis/redis/v8"
	"github.com/hsmtkk/curly-spork/env"
	"github.com/hsmtkk/curly-spork/filerepo"
	"github.com/hsmtkk/curly-spork/reportrepo"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

func main() {
	httpPort := env.RequiredInt("HTTP_PORT")
	redisHost := env.RequiredString("REDIS_HOST")
	redisPort := env.RequiredInt("REDIS_PORT")

	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("failed to initialize zap logger; %v", err)
	}
	defer logger.Sync()
	sugar := logger.Sugar()

	redisAddr := fmt.Sprintf("%s:%d", redisHost, redisPort)
	fileClient := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: "",
		DB:       0,
	})
	reportClient := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: "",
		DB:       1,
	})
	fileRepo := filerepo.New(sugar, fileClient)
	reportRepo := reportrepo.New(sugar, reportClient)

	hdl := &handler{sugar, fileRepo, reportRepo}

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.POST("/submit-file", hdl.SubmitFile)
	e.GET("/report-summary/:sha256", hdl.ReportSummary)
	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", httpPort)))
}

type handler struct {
	sugar      *zap.SugaredLogger
	fileRepo   fileRepo
	reportRepo reportRepo
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
	if err := h.fileRepo.PutFile(ctx.Request().Context(), sha256, data); err != nil {
		return fmt.Errorf("failed to put record; %w", err)
	}
	// TODO: write sha256 to the queue
	return ctx.String(http.StatusOK, sha256)
}

func (h *handler) ReportSummary(ctx echo.Context) error {
	sha256 := ctx.Param("sha256")
	h.sugar.Infow("report-summary", "sha256", sha256)
	report, err := h.reportRepo.GetReport(ctx.Request().Context(), sha256)
	if err != nil {
		return fmt.Errorf("failed to get record; %w", err)
	}
	return ctx.String(http.StatusOK, report)
}

func calcSHA256(data []byte) string {
	hash := sha256.Sum256(data)
	bs := hash[:]
	return hex.EncodeToString(bs)
}

type fileRepo interface {
	PutFile(ctx context.Context, sha256 string, data []byte) error
	GetFile(ctx context.Context, sha256 string) ([]byte, error)
}

type reportRepo interface {
	PutReport(ctx context.Context, sha256, report string) error
	GetReport(ctx context.Context, sha256 string) (string, error)
}
