package main

import (
	"embed"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"
	"net/http"
	"os"
	"strconv"
	"time"
	"youtube-summarise/handlers"
)

//go:embed templates/index.html
var content embed.FS

const defaultMaxVideoDuration = 20
const defaultCacheExpiration = 12
const defaultRequestsPerMinuteLimit = 5

func main() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())
	e.Renderer = handlers.NewIndexTemplate(content)

	config := rateLimitConfig()
	summaryHandler := getSummaryHandler()
	indexHandler := handlers.NewIndexHandler

	e.GET("/", indexHandler)
	e.POST("/summary", summaryHandler.SummaryHandler, middleware.RateLimiterWithConfig(config))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	e.Logger.Fatal(e.Start(":" + port))
}

func rateLimitConfig() middleware.RateLimiterConfig {
	requestsPerMinute := getIntFromEnvOrDefault("REQUEST_PER_MINUTE_RATE", defaultRequestsPerMinuteLimit)
	config := middleware.RateLimiterConfig{
		Skipper: middleware.DefaultSkipper,
		Store: middleware.NewRateLimiterMemoryStoreWithConfig(
			middleware.RateLimiterMemoryStoreConfig{Rate: rate.Limit(requestsPerMinute) / 60.0, Burst: 5, ExpiresIn: 10 * time.Minute},
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
	return config
}

func getSummaryHandler() *handlers.SummaryHandler {
	youtubeClient := NewYoutubeClient()
	summaryClient := NewSummaryClient(os.Getenv("OPENAI_KEY"))

	maxVideoDuration := getIntFromEnvOrDefault("MAX_VIDEO_DURATION_MINUTES", defaultMaxVideoDuration)
	cacheExpiration := getIntFromEnvOrDefault("CACHE_EXPIRATION_HOURS", defaultCacheExpiration)

	summaryHandler := handlers.NewSummaryHandler(
		youtubeClient,
		summaryClient,
		time.Duration(maxVideoDuration)*time.Minute,
		time.Duration(cacheExpiration)*time.Hour,
	)
	return summaryHandler
}

func getIntFromEnvOrDefault(key string, defaultValue int) int {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return intValue
}
