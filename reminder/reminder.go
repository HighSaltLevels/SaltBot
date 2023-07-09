package reminder

/*
import (
	"errors"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"

	"github.com/highsaltlevels/saltbot/cache"
	"github.com/highsaltlevels/saltbot/util"
)

const helpMessage string = ("```Set a reminder, show reminders or delete a reminder.\n\n To set one:\n" +
	"\"!remind set finish fixing saltbot bugs in 4 hours\"\n\nTo show all " +
	"reminders:\n\"!remind list\"\n\nTo delete a reminder:\n\"!remind delete" +
	"<ID>\" where <ID> is the id of the reminder given by \"!remind list\"```")

func parseReminder(args []string, m *discordgo.MessageCreate) (*cache.Reminder, error) {
	if args[len(args)-3] != "in" {
		return nil, errors.New(helpMessage)
	}

	unit := args[len(args)-1]
	duration := args[len(args)-2]
	expiry, err := util.ParseExpiry(unit, duration)
	if err != nil {
		return nil, fmt.Errorf("```Error parsing expiry: %w```%s\n", err, helpMessage)
	}

	id := strings.Split(uuid.NewString(), "-")[0]
	reminder := cache.Reminder{
		Author:  m.Author.ID,
		Channel: m.ChannelID,
		Expiry:  expiry,
		Message: strings.Join(args[:len(args)-3], " "),
		Id:      id,
	}

	return &reminder, nil
}

func Handle(m *discordgo.MessageCreate) (*discordgo.MessageSend, error) {
	args := strings.Split(m.Content, " ")[1:]
	if len(args) == 0 || args[0] == "help" {
		return &discordgo.MessageSend{
			Content: helpMessage,
		}, nil
	}

	switch args[0] {
	case "set":
		reminder, err := parseReminder(args[1:], m)
		if err != nil {
			return &discordgo.MessageSend{
				Content: err.Error(),
			}, nil
		}

		err = cache.Cache.AddReminder(reminder, m.Author.Username)
		if err != nil {
			return nil, fmt.Errorf("error adding reminder to k8s: %w", err)
		}

		return &discordgo.MessageSend{
			Content: fmt.Sprintf("```Created reminder with id: %s```", reminder.Id),
		}, nil

	case "list":
		msg := "```Reminders:\n"
		allReminders := cache.Cache.ListReminders()
		for _, reminder := range allReminders {
			if reminder.Author == m.Author.ID {
				expiry := util.TimeFromExpiry(reminder.Expiry)
				msg += fmt.Sprintf("%s: %s on %s\n", reminder.Id, reminder.Message, expiry)
			}
		}

		return &discordgo.MessageSend{
			Content: msg + "```",
		}, nil

	case "delete":
		if len(args) == 1 {
			return &discordgo.MessageSend{
				Content: "```To delete a reminder, you must specify the id. Use \"!remind list\" to see all your reminders```",
			}, nil
		}

		reminder := cache.Cache.GetReminder(args[1], m.Author.ID)
		if reminder == nil {
			return &discordgo.MessageSend{
				Content: "```Either that reminder doesn't exist or you don't have access to it```",
			}, nil
		}

		cache.Cache.Delete("reminder-" + reminder.Id)
		return &discordgo.MessageSend{
			Content: fmt.Sprintf("```Deleted reminder %s```", reminder.Id),
		}, nil
	}

	return &discordgo.MessageSend{
		Content: helpMessage,
	}, nil
}
*/
