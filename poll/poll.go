package poll
/*

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"

	"github.com/highsaltlevels/saltbot/cache"
	"github.com/highsaltlevels/saltbot/util"
)

const helpMessage string = ("```How to set a poll:\n" +
	"Type the \"!poll\" command followed by the question, answers, and\n" +
	"the time separated by semicolons. For Example:\n\n" +
	"!poll Would you rather eat poop flavored curry or curry flavored poop? ;" +
	" poop flavored curry ; curry flavord poop ; neither ; ends in 2 hours\n\n" +
	"The poll expiry must be in the format \"ends in X Y\" where X is any\n" +
	"positive integer and Y is one of (hours, hour, minutes, minute, seconds,\n" +
	"second)```")

const voteHelpMessage string = ("```To vote on a poll, use \"!vote <poll id> " +
	"<choice num>\". For example: \"!vote dd32251a 1\"'''")

func parsePoll(args []string, m *discordgo.MessageCreate) (*cache.Poll, error) {
	prompt := strings.Replace(strings.Replace(args[0], "!poll", "", 1), "!p", "", 1)
	units := strings.Split(args[len(args)-1], " ")
	unit := units[len(units)-1]
	duration := units[len(units)-2]
	expiry, err := util.ParseExpiry(unit, duration)
	if err != nil {
		return nil, fmt.Errorf("```Error parsing expiry: %w```%s\n", err, helpMessage)
	}

	choices := make([]string, len(args)-2)
	for idx, choice := range args[1 : len(args)-1] {
		choices[idx] = strings.TrimSpace(choice)
	}

	id := strings.Split(uuid.NewString(), "-")[0]
	return &cache.Poll{
		Author:  m.Author.ID,
		Channel: m.ChannelID,
		Prompt:  strings.TrimSpace(prompt),
		Choices: choices,
		Expiry:  expiry,
		Id:      id,
		Votes:   map[string][]string{},
	}, nil
}

func Create(m *discordgo.MessageCreate) (*discordgo.MessageSend, error) {
	args := strings.Split(m.Content, " ")[1:]
	if len(args) < 4 || args[0] == "help" {
		return &discordgo.MessageSend{
			Content: helpMessage,
		}, nil
	}

	// Verify it says "ends in X Y"
	if (args[len(args)-4]) != "ends" && (args[len(args)-4]) != "in" {
		return &discordgo.MessageSend{
			Content: helpMessage,
		}, nil
	}

	// Now split on semicolon, perform checking there, and begin parsing
	args = strings.Split(m.Content, ";")
	if len(args) < 4 {
		return &discordgo.MessageSend{
			Content: helpMessage,
		}, nil
	}

	poll, err := parsePoll(args, m)
	if err != nil {
		return &discordgo.MessageSend{
			Content: err.Error(),
		}, nil
	}

	err = cache.Cache.AddPoll(poll)
	if err != nil {
		return nil, fmt.Errorf("error adding poll to k8s: %w", err)
	}

	// Create the poll string
	msg := fmt.Sprintf("```%s\n\n", poll.Prompt)
	for idx, choice := range poll.Choices {
		msg += fmt.Sprintf("%d. %s\n", idx+1, choice)
	}
	msg += fmt.Sprintf("Type of DM me \"!vote %s <choice number>\" to vote```", poll.Id)

	return &discordgo.MessageSend{
		Content: msg,
	}, nil
}

func Vote(m *discordgo.MessageCreate) (*discordgo.MessageSend, error) {
	args := strings.Split(m.Content, " ")[1:]
	if len(args) < 2 {
		return &discordgo.MessageSend{
			Content: voteHelpMessage,
		}, nil
	}

	choiceNum, err := strconv.Atoi(args[1])
	if err != nil {
		return &discordgo.MessageSend{
			Content: voteHelpMessage,
		}, nil
	}

	poll := cache.Cache.GetPoll(args[0], m.Author.ID)
	if poll == nil {
		return &discordgo.MessageSend{
			Content: fmt.Sprintf("```Poll %s does not exist!```", args[0]),
		}, nil
	}

	if choiceNum > len(poll.Choices) || choiceNum < 1 {
		return &discordgo.MessageSend{
			Content: fmt.Sprintf("```No such choice number: %d```", choiceNum),
		}, nil
	}

	// Strip the users current vote so that they can't double vote.
	votes := make(map[string][]string, len(poll.Votes))
	for vote, _ := range poll.Votes {
		if !util.Contains(poll.Votes[vote], m.Author.Username) {
			votes[vote] = poll.Votes[vote]
		}
	}

	choiceStr := fmt.Sprintf("%d", choiceNum-1)
	votes[choiceStr] = append(votes[choiceStr], m.Author.Username)

	updatedPoll := cache.Poll{
		Author:  poll.Author,
		Channel: poll.Channel,
		Prompt:  poll.Prompt,
		Choices: poll.Choices,
		Expiry:  poll.Expiry,
		Id:      poll.Id,
		Votes:   votes,
	}
	err = cache.Cache.UpdatePoll(&updatedPoll)
	if err != nil {
		return nil, fmt.Errorf("failed to update poll: %w", err)
	}

	return &discordgo.MessageSend{
		Content: fmt.Sprintf("```You have voted for %s```", poll.Choices[choiceNum-1]),
	}, nil
}
*/
