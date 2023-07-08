package giphy

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/highsaltlevels/saltbot/util"
)

var token string

type GiphyData struct {
	// We only care about the bitly_gif_url
	Url string `json:"bitly_gif_url"`
}

type GiphyResponse struct {
	Data []GiphyData `json:"data"`
}

func init() {
	var ok bool
	if token, ok = os.LookupEnv("GIPHY_AUTH"); !ok {
		log.Fatal("failed to get giphy auth from env var")
	}
}

func fetchGif(query string) (*GiphyResponse, error) {
	url := fmt.Sprintf("http://api.giphy.com/v1/gifs/search?q=%s&api_key=%s", query, token)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get giphy gif: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received status code %d from giphy", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read giphy resp: %w", err)
	}

	var gif GiphyResponse
	err = json.Unmarshal(body, &gif)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal giphy response: %w", err)
	}

	return &gif, nil
}

func getAllGifs(query string) (string, error) {
	gif, err := fetchGif(query)
	if err != nil {
		return "", err
	}

	msg := "Here's all the gifs for that query:\n"
	for _, data := range gif.Data {
		msg += data.Url + "\n"
	}

	return msg, nil
}

func getGif(query string, idx int) (string, error) {
	gif, err := fetchGif(query)
	if err != nil {
		return "", err
	}

	return gif.Data[idx].Url, nil
}

func Get(content string) (*discordgo.MessageSend, error) {
	args := strings.Split(content, " ")[1:]
	if len(args) == 0 {
		return &discordgo.MessageSend{
			Content: "```Must specify giphy query like: \"!giphy dog\"```",
		}, nil
	}

	for i, arg := range args {
		if arg == "-i" {
			idx, err := strconv.Atoi(args[i+1])
			if err != nil || idx < 0 || idx > 24 {
				return &discordgo.MessageSend{
					Content: "```Must use a valid number between 0 and 24```",
				}, nil
			}

			gif, err := getGif(util.ParseArgsToQuery(args), idx)
			return &discordgo.MessageSend{
				Content: gif,
			}, err
		}

		if arg == "-a" {
			gif, err := getAllGifs(util.ParseArgsToQuery(args))
			return &discordgo.MessageSend{
				Content: gif,
			}, err
		}
	}

	// If we get here, then -a and -i were not used. Assume first query result
	gif, err := getGif(util.ParseArgsToQuery(args), 0)
	return &discordgo.MessageSend{
		Content: gif,
	}, err
}
