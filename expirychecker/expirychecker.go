package expirychecker

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
	c "github.com/highsaltlevels/saltbot/cache"
)

type SessionInterface interface {
	ChannelMessageSend(channelID string, content string, options ...discordgo.RequestOption) (*discordgo.Message, error)
}

type Poller struct {
	session SessionInterface
	ctx     context.Context
}

func NewPoller(s SessionInterface, ctx context.Context) *Poller {
	return &Poller{
		session: s,
		ctx:     ctx,
	}
}

func (p *Poller) Loop() {
	for {
		select {
		case <-p.ctx.Done():
			fmt.Println("poller stopped")
			return

		case <-time.After(1 * time.Second):
			polls, reminders := p.getExpired()
			for _, poll := range polls {
				log.Printf("sending poll %s to %s\n", poll.Id, poll.Channel)
				err := p.sendPoll(&poll)
				if err != nil {
					log.Printf("error sending poll: %v\n", err)
					log.Printf("will retry again on the next iteration")
				} else {
					c.Cache.Delete(fmt.Sprintf("poll-%s", poll.Id))
				}
			}

			for _, reminder := range reminders {
				log.Printf("sending reminder %s to %s\n", reminder.Id, reminder.Channel)
				err := p.sendReminder(&reminder)
				if err != nil {
					log.Printf("error sending reminder: %v\n", err)
					log.Printf("will retry again on the next iteration")
				} else {
					c.Cache.Delete(fmt.Sprintf("reminder-%s", reminder.Id))
				}
			}
		}
	}
}

func (p *Poller) getExpired() (polls []c.Poll, reminders []c.Reminder) {
	for _, poll := range c.Cache.ListPolls() {
		expiry := time.Unix(poll.Expiry, 0)
		if time.Now().Sub(expiry) >= 0.0 {
			polls = append(polls, poll)
		}
	}

	for _, reminder := range c.Cache.ListReminders() {
		expiry := time.Unix(reminder.Expiry, 0)
		if time.Now().Sub(expiry) >= 0.0 {
			reminders = append(reminders, reminder)
		}
	}

	return polls, reminders
}

func (p *Poller) sendPoll(poll *c.Poll) error {
	var msg string
	totalVotes := 0
	results := make([]int, len(poll.Choices))
	for choiceNum := range poll.Choices {
		total := len(poll.Votes[strconv.Itoa(choiceNum)])
		results[choiceNum] = total
		totalVotes += total
	}

	if totalVotes == 0 {
		msg = "```No one voted on this poll :("
	} else {
		msg = fmt.Sprintf("```Results for prompt \"%s\" (Total votes: %d):\n\n", poll.Prompt, totalVotes)
		for idx := range results {
			choice := poll.Choices[idx]
			votesForChoice := len(poll.Votes[strconv.Itoa(idx)])

			result := float64(votesForChoice) / float64(totalVotes) * 100.0
			msg += fmt.Sprintf("\t%s -> %.0f%%\n", choice, result)
		}
	}

	return p.sendMessage(poll.Channel, fmt.Sprintf("%s```", msg))
}

func (p *Poller) sendReminder(r *c.Reminder) error {
	msg := fmt.Sprintf("```%s```", r.Message)
	return p.sendMessage(r.Channel, msg)
}

func (p *Poller) sendMessage(channel string, msg string) error {
	_, err := p.session.ChannelMessageSend(channel, msg)
	return err
}
