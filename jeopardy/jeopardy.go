package jeopardy

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"

	"github.com/bwmarrin/discordgo"

	"github.com/highsaltlevels/saltbot/util"
)

var client util.HttpClientInterface

type Clue struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

type JeopardyResponse struct {
	Title string `json:"title"`
	Clues []Clue `json:"clues"`
}

func init() {
	if client == nil {
		client = &http.Client{}
	}
}

func Get() (*discordgo.MessageSend, error) {
	num := rand.Intn(18417)
	url := fmt.Sprintf("http://jservice.io/api/category?id=%d", num)

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error getting jeopardy questions: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("got %d status code getting jeopardy question", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read jeopardy response: %v", err)
	}

	var jeopardyResp JeopardyResponse
	err = json.Unmarshal(body, &jeopardyResp)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal jeopardy response: %v", err)
	}

	// Build the message string
	msg := fmt.Sprintf("```The Category is: %s```\n", jeopardyResp.Title)
	for i, clue := range jeopardyResp.Clues {
		msg = fmt.Sprintf("%sQuestion %d: %s\nAnswer: ||%s||\n\n", msg, i+1, clue.Question, clue.Answer)
	}

	return &discordgo.MessageSend{
		Content: msg,
	}, nil
}
