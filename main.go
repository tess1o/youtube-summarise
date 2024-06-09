package main

import (
	"embed"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"net/http"
	"os"
	"strconv"
	"time"
	"youtube-summarise/handlers"
)

//go:embed templates/index.html
var content embed.FS

const defaultMaxVideoDuration = "20"

func main() {
	summaryHandler, err := getSummaryHandler()
	if err != nil {
		panic(err)
	}
	indexHandler := handlers.NewIndexHandler

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())

	config := middleware.RateLimiterConfig{
		Skipper: middleware.DefaultSkipper,
		Store: middleware.NewRateLimiterMemoryStoreWithConfig(
			middleware.RateLimiterMemoryStoreConfig{Rate: 2 / 60.0, Burst: 1, ExpiresIn: 1 * time.Minute},
		),
		IdentifierExtractor: func(ctx echo.Context) (string, error) {
			id := ctx.RealIP()
			return id, nil
		},
		ErrorHandler: func(context echo.Context, err error) error {
			return context.String(http.StatusForbidden, "Unexpected server error")
		},
		DenyHandler: func(context echo.Context, identifier string, err error) error {
			return context.String(http.StatusTooManyRequests, "Slow down, cowboy. Too many requests")
		},
	}

	e.Renderer = handlers.NewIndexTemplate(content)

	e.GET("/", indexHandler)
	e.POST("/summary", summaryHandler.SummaryHandler, middleware.RateLimiterWithConfig(config))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	e.Logger.Fatal(e.Start(":" + port))
}

func getSummaryHandler() (*handlers.SummaryHandler, error) {
	youtubeClient := NewYoutubeClient()
	summaryClient := NewSummaryClient(os.Getenv("OPENAI_KEY"))

	maxVideoDuration, err := getMaxVideoDuration()
	if err != nil {
		return nil, err
	}

	summaryHandler := handlers.NewSummaryHandler(youtubeClient, summaryClient, maxVideoDuration)
	return summaryHandler, nil
}

func getMaxVideoDuration() (time.Duration, error) {
	maxDuration, exists := os.LookupEnv("MAX_VIDEO_DURATION_MINUTES")

	if !exists {
		maxDuration = defaultMaxVideoDuration
	}

	maxDurationInt, err := strconv.Atoi(maxDuration)
	if err != nil {
		return time.Minute, err
	}
	return time.Minute * time.Duration(maxDurationInt), nil
}
