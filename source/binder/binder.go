package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"

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
			Description: "Lend an card to another user",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        "user",
					Description: "User to lend the card to",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "card",
					Description: "Name of the card to lend",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "quantity",
					Description: "Number of copies lent",
					Required:    false,
				},
			},
		},
		{
			Name:        "borrow",
			Description: "Borrow an card from another user",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        "user",
					Description: "User to borrow the card from",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "card",
					Description: "Name of the card to borrow",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "quantity",
					Description: "Number of copies borrowed",
					Required:    false,
				},
			},
		},
		{
			Name:        "return",
			Description: "Return a borrowed card",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        "user",
					Description: "User to return the card to",
					Required:    true,
				},
				{
					Type:         discordgo.ApplicationCommandOptionString,
					Name:         "card",
					Description:  "Name of the borrowed card",
					Required:     true,
					Autocomplete: true,
				},
			},
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
	options := i.ApplicationCommandData().Options
	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
	for _, opt := range options {
		optionMap[opt.Name] = opt
	}

	userOpt := optionMap["user"]
	cardOpt := optionMap["card"]

	var quantity int64 = 1
	quantityOpt, ok := optionMap["quantity"]
	if ok {
		quantity = quantityOpt.IntValue()
	}

	borrowerUser := userOpt.UserValue(s)
	cardName := cardOpt.StringValue()

	lenderID, _ := strconv.ParseInt(i.Member.User.ID, 10, 64)
	borrowerID, _ := strconv.ParseInt(borrowerUser.ID, 10, 64)

	tm.Trades <- Trade{
		LenderID: lenderID,
		Borrower: borrowerID,
		CardName: cardName,
		Quantity: quantity,
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("<@%s> has lent <@%s> %d %s", i.Member.User.ID, borrowerUser.ID, quantity, cardName),
		},
	})
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
