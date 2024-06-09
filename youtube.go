package main

import (
	"errors"
	"fmt"
	"github.com/kkdai/youtube/v2"
	"log"
	"regexp"
	"strings"
	"time"
)

const preferableLanguage = "en"

var (
	videoIdRegexp            = regexp.MustCompile(`(?:youtube\.com/.*[?&]v=|youtu\.be/|youtube\.com/embed/)([^&/?]+)`)
	wrongUrlError            = errors.New("no video id can be extracted from the given url")
	videoTooLongErrorMessage = "The video is too long. Maximum duration: %d minutes"
)

type YoutubeClient struct {
	client youtube.Client
}

func NewYoutubeClient() *YoutubeClient {
	return &YoutubeClient{client: youtube.Client{}}
}

func (y *YoutubeClient) GetVideoId(videoUrl string) (string, error) {
	match := videoIdRegexp.FindStringSubmatch(videoUrl)
	if len(match) > 1 {
		videoID := match[1]
		return videoID, nil
	} else {
		return "", wrongUrlError
	}
}

func (y *YoutubeClient) GetTranscript(videoUrl string, maxDuration time.Duration) (string, error) {
	videoId := videoUrl
	if strings.Contains(videoUrl, "http") {
		var err error
		videoId, err = y.GetVideoId(videoUrl)
		if err != nil {
			return "", err
		}
	}
	video, err := y.client.GetVideo(videoId)
	if err != nil {
		return "", err
	}
	log.Printf("Video - %+v\n", video)

	if len(video.CaptionTracks) == 0 {
		return "", errors.New("there is no subtitles available, cannot get the summary of the video")
	}

	if video.Duration > maxDuration {
		return "", errors.New(fmt.Sprintf(videoTooLongErrorMessage, int(maxDuration.Minutes())))
	}

	selectedLanguage := determineTranscriptLanguage(video.CaptionTracks)

	transcript, err := y.client.GetTranscript(video, selectedLanguage)
	if err != nil {
		return "", err
	}

	var transcriptFull string
	for _, line := range transcript {
		transcriptFull += line.Text + "\n"
	}

	return transcriptFull, nil
}

// Extracted function
func determineTranscriptLanguage(captionTracks []youtube.CaptionTrack) string {
	selectedLanguage := ""
	for _, c := range captionTracks {
		if c.LanguageCode == preferableLanguage {
			return preferableLanguage
		}
		if selectedLanguage == "" {
			selectedLanguage = c.LanguageCode
		}
	}
	return selectedLanguage
}
