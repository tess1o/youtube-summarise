package main

import (
	"context"
	"fmt"
	"github.com/sashabaranov/go-openai"
)

const prompt = `I have a video transcript from YouTube.
Please generate a concise summary of the video based on the provided transcript. 
The answer should be in english. In your response do not mention that it's a video transcript. 
The summary should be detailed enough so the user can understand what the video is about.
Here is the transcript:%s`

type SummaryClient struct {
	client *openai.Client
}

func NewSummaryClient(key string) *SummaryClient {
	return &SummaryClient{client: openai.NewClient(key)}
}

func (c *SummaryClient) GetSummary(transcript string) (string, error) {
	resp, err := c.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: fmt.Sprintf(prompt, transcript),
				},
			},
		},
	)

	if err != nil {
		return "", err
	}

	return resp.Choices[0].Message.Content, nil
}
