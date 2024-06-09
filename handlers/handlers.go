package handlers

import (
	"github.com/labstack/echo/v4"
	"log"
	"net/http"
	"time"
)

type youtubeService interface {
	GetTranscript(videoUrl string, maxDuration time.Duration) (string, error)
}

type summariseService interface {
	GetSummary(text string) (string, error)
}

type SummaryHandler struct {
	youtube          youtubeService
	summarise        summariseService
	maxVideoDuration time.Duration
}

func NewSummaryHandler(y youtubeService, s summariseService, maxVideoDuration time.Duration) *SummaryHandler {
	return &SummaryHandler{
		youtube:          y,
		summarise:        s,
		maxVideoDuration: maxVideoDuration,
	}
}

func (h *SummaryHandler) SummaryHandler(c echo.Context) error {
	videoUrl := c.FormValue("input")
	if videoUrl == "" {
		return c.String(http.StatusBadRequest, "Input cannot be empty")
	}

	transcript, err := h.youtube.GetTranscript(videoUrl, h.maxVideoDuration)

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
	return c.String(http.StatusOK, summary)
}
