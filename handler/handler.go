package handler

import (
	"fmt"
	"log"
	"math/rand"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"

	"github.com/highsaltlevels/saltbot/giphy"
	"github.com/highsaltlevels/saltbot/jeopardy"
	"github.com/highsaltlevels/saltbot/poll"
	"github.com/highsaltlevels/saltbot/reminder"
	"github.com/highsaltlevels/saltbot/youtube"
)

const help string = ("```Good salty day to you! Here's a list of commands that I understand:\n\n" +
	"!help (!h):     Shows this help message.\n" +
	"!jeopardy (!j): Recieve a category with 5 questions and answers. The answers\n" +
	"                are marked as spolers and are not revealed until you click them\n" +
	"!whipser (!pm): Get a salty DM from SaltBot. This can be used as a playground\n" +
	"                for experiencing all of the salty features.\n" +
	"!gif (!g):      Type !gif followed by keywords to get a cool gif. For example\n" +
	"                \"!gif dog\".\n" +
	"!waifu (!w):    Get a picture of a randomized waifu.\n" +
	"!poll (!p):     Type \"!poll help\" for detailed information\n" +
	"!vote (!v):     Vote in a poll. Type \"!vote <poll id> <poll choice> to vote\n" +
	"!youtube (!y):  Get a youtube search result. Use the \"-i\" parameter to specify\n" +
	"                an index. For example: \"!y dog -i 3\" to get the 3rd query result.\n" +
	"!remind (!r):   Set a reminder. Type \"remind help \" for detailed information\n\n" +
	"Check me out on github: https://github.com/highsaltlevels/saltbot```")

func GetHelpMsg() *discordgo.MessageSend {
	return &discordgo.MessageSend{
		Content: help,
	}
}

func GetWaifu() *discordgo.MessageSend {
	num := rand.Intn(99999)
	url := fmt.Sprintf("https://www.thiswaifudoesnotexist.net/example-%d.jpg", num)
	return &discordgo.MessageSend{
		Content: "Here's a waifu for you!",
		Embeds: []*discordgo.MessageEmbed{
			&discordgo.MessageEmbed{
				URL:         url,
				Type:        discordgo.EmbedTypeImage,
				Title:       "Here's the sauce site",
				Description: "Check out the link above to go to the site that makes this feature possible",
				Image: &discordgo.MessageEmbedImage{
					URL: url,
				},
			},
		},
	}
}

func SendDM(s *discordgo.Session, m *discordgo.MessageCreate) {
	channel, err := s.UserChannelCreate(m.Author.ID)
	if err != nil {
		s.ChannelMessageSendComplex(m.ChannelID, CreateError(err))
		return
	}

	msg := fmt.Sprintf("```Hello %s! You can talk to me here (where no one can hear our mutual salt)```", m.Author.Username)
	s.ChannelMessageSend(channel.ID, msg)
}

func CreateError(err error) *discordgo.MessageSend {
	// Log the exact error but return a generic error message.
	errId := uuid.NewString()
	log.Printf("error with id: %s: %v", errId, err)
	return &discordgo.MessageSend{
		Content: fmt.Sprintf("```Unexpected error with id: %s :(```", errId),
	}
}

func OnMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore messages created by saltbot
	if m.Author.ID == s.State.User.ID {
		return
	}

	// Get the command and retrieve its associated message.
	var message *discordgo.MessageSend
	var err error
	command := strings.Split(m.Content, " ")[0]
	switch command {
	case "!help":
		message = GetHelpMsg()
	case "!h":
		message = GetHelpMsg()
	case "!waifu":
		message = GetWaifu()
	case "!w":
		message = GetWaifu()
	case "!jeopardy":
		message, err = jeopardy.Get()
	case "!j":
		message, err = jeopardy.Get()
	case "!whisper":
		SendDM(s, m)
		return
	case "!pm":
		SendDM(s, m)
		return
	case "!gif":
		message, err = giphy.Get(m.Content)
	case "!g":
		message, err = giphy.Get(m.Content)
	case "!youtube":
		message, err = youtube.Get(m.Content)
	case "!y":
		message, err = youtube.Get(m.Content)
	case "!remind":
		message, err = reminder.Handle(m)
	case "!r":
		message, err = reminder.Handle(m)
	case "!poll":
		message, err = poll.Create(m)
	case "!p":
		message, err = poll.Create(m)
	case "!vote":
		message, err = poll.Vote(m)
	case "!v":
		message, err = poll.Vote(m)
	// If saltbot doesn't know the command, do nothing
	default:
		return
	}

	// If there was an error, send an error message instead.
	if err != nil {
		message = CreateError(err)
	}

	_, err = s.ChannelMessageSendComplex(m.ChannelID, message)
	if err != nil {
		log.Printf("unexpected error sending message: %v\n", err)
	}
}
