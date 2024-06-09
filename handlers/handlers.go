package handlers

import (
	"github.com/labstack/echo/v4"
	"github.com/patrickmn/go-cache"
	"log"
	"net/http"
	"strings"
	"time"
)

type youtubeService interface {
	GetVideoId(videoUrl string) (string, error)
	GetTranscript(videoId string, maxDuration time.Duration) (string, error)
}

type summariseService interface {
	GetSummary(text string) (string, error)
}

type SummaryHandler struct {
	youtube          youtubeService
	summarise        summariseService
	maxVideoDuration time.Duration
	c                *cache.Cache
}

func NewSummaryHandler(y youtubeService, s summariseService, maxVideoDuration time.Duration, cacheExpiration time.Duration) *SummaryHandler {
	return &SummaryHandler{
		youtube:          y,
		summarise:        s,
		maxVideoDuration: maxVideoDuration,
		c:                cache.New(cacheExpiration, cacheExpiration*2),
	}
}

func (h *SummaryHandler) SummaryHandler(c echo.Context) error {
	input := c.FormValue("input")
	if input == "" {
		return c.String(http.StatusBadRequest, "Input cannot be empty")
	}

	log.Printf("Received input from user: %s\n", input)

	videoId := input

	if strings.Contains(input, "http") {
		var err error
		videoId, err = h.youtube.GetVideoId(input)
		if err != nil {
			return c.String(http.StatusBadRequest, "Cannot identify YouTube video id")
		}
	}

	summaryFromCache, found := h.c.Get(videoId)
	if found {
		log.Printf("Found results in cache for id %s, returning\n", videoId)
		return c.String(http.StatusOK, summaryFromCache.(string))
	}

	transcript, err := h.youtube.GetTranscript(videoId, h.maxVideoDuration)

	if err != nil {
		log.Printf("Error from youtube:%+v", err)
		return c.String(http.StatusInternalServerError, "Unexpected error from YouTube, sorry")
	}

	summary, err := h.summarise.GetSummary(transcript)

	if err != nil {
		log.Printf("Error from openai:%+v", err)
		return c.String(http.StatusInternalServerError, "Unable to generate summary for this video, please try later")
	}

	log.Printf("Response from openai: %s\n", summary)
	h.c.Set(videoId, summary, cache.DefaultExpiration)
	return c.String(http.StatusOK, summary)
}
