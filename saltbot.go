package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/highsaltlevels/saltbot/cache"
	"github.com/highsaltlevels/saltbot/expirychecker"
	"github.com/highsaltlevels/saltbot/handler"
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
	var err error
	time.Local, err = time.LoadLocation("US/Eastern")
	if err != nil {
		log.Fatalf("failed to load locale: %w", err)
	}

	session, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalf("failed to initialize saltbot: %w", err)
	}

	err = session.Open()
	if err != nil {
		log.Fatalf("failed to open discord socket: %w", err)
	}
	defer session.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Println("initializing messenger")
	checker := expirychecker.NewPoller(session, cache.Cache, ctx)
	go checker.Loop()

	log.Println("registering message handlers")
	session.AddHandler(handler.OnMessageCreate)

	log.Println("salbot initialized and logged in")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	log.Println("closing down gracefully")
}
