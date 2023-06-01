package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/highsaltlevels/saltbot/cache"
	"github.com/highsaltlevels/saltbot/poll"
)

// Bot token
var token string

func init() {
	var ok bool
	if token, ok = os.LookupEnv("BOT_TOKEN"); !ok {
		log.Fatal("failed to get bot token from env var")
	}
}

func main() {
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalf("failed to initialize saltbot: %v", err)
	}

	err = session.Open()
	if err != nil {
		log.Fatalf("failed to open discord socket: %v", err)
	}
	defer session.Close()

	log.Println("salbot messenger initialized and logged in")

	c := cache.NewConfigMapCache()
	ctx, pollCancel := context.WithCancel(context.Background())
	poller := poll.NewPoller(session, c, ctx)
	go poller.Loop()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	log.Println("closing down gracefully")
	pollCancel()
}
