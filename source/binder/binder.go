package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"

	"github.com/bwmarrin/discordgo"
)

const BotTokenEnv = "DISCORD_BOT_TOKEN"

var tradeFile = flag.String("trades", "", "path to trades CSV file")

func run(tm *TradeManager) error {
	dg, err := discordgo.New("Bot " + os.Getenv(BotTokenEnv))
	if err != nil {
		return err
	}

	dg.AddHandler(interactionHandler(dg, tm))

	dg.Identify.Intents |= discordgo.IntentsAllWithoutPrivileged
	if err := dg.Open(); err != nil {
		return err
	}

	_, err = dg.ApplicationCommandBulkOverwrite(dg.State.User.ID, "", commands(dg))
	if err != nil {
		return err
	}

	return nil
}

func commands(_ *discordgo.Session) []*discordgo.ApplicationCommand {
	return []*discordgo.ApplicationCommand{
		{
			Name:        "lend",
			Description: "Lend an item to another user",
		},
		{
			Name:        "borrow",
			Description: "Borrow an item from another user",
		},
		{
			Name:        "return",
			Description: "Return a borrowed item",
		},
	}
}

func interactionHandler(s *discordgo.Session, tm *TradeManager) func(*discordgo.Session, *discordgo.InteractionCreate) {
	return func(_ *discordgo.Session, i *discordgo.InteractionCreate) {
		data := i.ApplicationCommandData()
		switch data.Name {
		case "lend":
			handleLend(s, i, tm)
		case "borrow":
			handleBorrow(s, i, tm)
		case "return":
			handleReturn(s, i, tm)
		}
	}
}

func handleLend(s *discordgo.Session, i *discordgo.InteractionCreate, tm *TradeManager) {
	_ = tm // implement /lend logic here
}

func handleBorrow(s *discordgo.Session, i *discordgo.InteractionCreate, tm *TradeManager) {
	_ = tm // implement /borrow logic here
}

func handleReturn(s *discordgo.Session, i *discordgo.InteractionCreate, tm *TradeManager) {
	_ = tm // implement /return logic here
}

func main() {
	flag.Parse()

	tm, err := NewTradeManager(*tradeFile)
	if err != nil {
		log.Fatal(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	go func() {
		if err := run(tm); err != nil {
			log.Fatal(err)
		}
	}()

	<-ctx.Done()
	stop()
}
