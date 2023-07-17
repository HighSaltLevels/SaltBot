package youtube

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
var client util.HttpClientInterface

type YoutubeId struct {
	VideoId string `json:"videoId"`
}

type YoutubeVideo struct {
	Id YoutubeId `json:"id"`
}

type YoutubeResponse struct {
	Items []YoutubeVideo `json:"items"`
}

func init() {
	var ok bool
	if token, ok = os.LookupEnv("YOUTUBE_AUTH"); !ok {
		log.Println("failed to get youtube auth from env var")
		log.Println("continuing saltbot startup with partial functionality")
	}

	if client == nil {
		client = &http.Client{}
	}
}

func getYoutubeVideo(query string, idx int) (string, error) {
	url := fmt.Sprintf("https://www.googleapis.com/youtube/v3/search?key=%s&q=%s&maxResult=15&type=video", token, query)
	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to get youtube video: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("received status code %d from youtube", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read youtube resp: %w", err)
	}

	var yt YoutubeResponse
	err = json.Unmarshal(body, &yt)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal youtube response: %w", err)
	}

	if len(yt.Items) < 1 {
		return "```No videos for that query :(```", nil
	}

	return "https://www.youtube.com/watch?v=" + yt.Items[idx].Id.VideoId, nil
}

func Get(content string) (*discordgo.MessageSend, error) {
	args := strings.Split(content, " ")[1:]
	if len(args) == 0 {
		return &discordgo.MessageSend{
			Content: "```Must specify a query like: \"!youtube dog\"```",
		}, nil
	}

	for i, arg := range args {
		if arg == "-i" {
			idx, err := strconv.Atoi(args[i+1])
			if err != nil || idx < 0 || idx > 14 {
				return &discordgo.MessageSend{
					Content: "```Must use a valid number between 0 and 14```",
				}, nil
			}

			yt, err := getYoutubeVideo(util.ParseArgsToQuery(args), idx)
			return &discordgo.MessageSend{
				Content: yt,
			}, err
		}
	}

	// If we get here, the user didn't specify "-i", so just assume first result
	yt, err := getYoutubeVideo(util.ParseArgsToQuery(args), 0)
	return &discordgo.MessageSend{
		Content: yt,
	}, err
}
