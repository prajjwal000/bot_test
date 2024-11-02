package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/joho/godotenv"
	"github.com/bwmarrin/discordgo"
)

func main() {
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatal(err)
	}

	token := os.Getenv("DISCORD_TOKEN")

	discord, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatal(err)
	}

	discord.Identify.Intents |= discordgo.IntentsMessageContent
	discord.Identify.Intents |= discordgo.IntentsGuildMessages

	discord.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as %s", r.User.String())
	})

	discord.AddHandler(func( s *discordgo.Session, r *discordgo.MessageCreate){
		if r.Author.ID == s.State.User.ID {
			return
		}

		if r.Content == "!hello" {
			s.ChannelMessageSend(r.ChannelID, "Hello!!!")
		}
	})

	err = discord.Open()
	if err != nil {
		log.Print("Failure Starting Bot\n")
	}



	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, os.Interrupt)
	<-sigch

	err = discord.Close()
	if err != nil {
		log.Printf("Not graceful...")
	}
}
