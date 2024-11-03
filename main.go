package main

import (
	"log"
	"io"
	"os"
	"os/signal"
	"encoding/binary"
	"time"

	"github.com/joho/godotenv"
	"github.com/bwmarrin/discordgo"
)

var buffer = make([][]byte,0)

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

	discord.Identify.Intents = discordgo.IntentsMessageContent | discordgo.IntentsGuildMessages | discordgo.IntentsGuildVoiceStates | discordgo.IntentsGuilds

	err = loadSound()
	if err != nil {
		log.Println("Error: ",err)
		return
	}

	discord.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as %s", r.User.String())
	})

	discord.AddHandler(guildCreate)

	discord.AddHandler(func( s *discordgo.Session, r *discordgo.MessageCreate){
		if r.Author.ID == s.State.User.ID {
			return
		}

		if r.Content == "!sing" {

			c, err := s.State.Channel(r.ChannelID)
			if err != nil {
				log.Println("get channel error...")
				return
			}

			g, err := s.State.Guild(c.GuildID)
			if err != nil {
				log.Println("Get Guild Error")
				return
			}

			for _,vs := range g.VoiceStates {
				if vs.UserID == r.Author.ID {
					err = playSound(s, g.ID, vs.ChannelID)
					if err != nil {
						log.Println("Error playing jingles...")
					}

					return
				}
			}

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
		log.Fatal("Not graceful...")
	}

	log.Print("Gracefully closing")
}

func guildCreate(s *discordgo.Session, event *discordgo.GuildCreate) {

	if event.Guild.Unavailable {
		return
	}

	for _, channel := range event.Guild.Channels {
		if channel.ID == event.Guild.ID {
			_, _ = s.ChannelMessageSend(channel.ID, "Airhorn is ready! Type !airhorn while in a voice channel to play a sound.")
			return
		}
	}
}

func loadSound() error {

	file, err := os.Open("jingle.dca")
	if err != nil {
		log.Println("Error opening dca file :", err)
		return err
	}

	var opuslen int16

	for {
		// Read opus frame length from dca file.
		err = binary.Read(file, binary.LittleEndian, &opuslen)

		// If this is the end of the file, just return.
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			err := file.Close()
			if err != nil {
				return err
			}
			return nil
		}

		if err != nil {
			log.Println("Error reading from dca file :", err)
			return err
		}

		// Read encoded pcm from dca file.
		InBuf := make([]byte, opuslen)
		err = binary.Read(file, binary.LittleEndian, &InBuf)

		// Should not be any end of file errors
		if err != nil {
			log.Println("Error reading from dca file :", err)
			return err
		}

		// Append encoded pcm data to the buffer.
		buffer = append(buffer, InBuf)
	}
}

func playSound(s *discordgo.Session, guildID, channelID string) (err error) {

	// Join the provided voice channel.
	vc, err := s.ChannelVoiceJoin(guildID, channelID, false, true)
	if err != nil {
		return err
	}

	// Sleep for a specified amount of time before playing the sound
	time.Sleep(250 * time.Millisecond)

	// Start speaking.
	err = vc.Speaking(true)
	if err != nil {
		log.Println("Fucked up speaking...")
	}

	// Send the buffer data.
	for _, buff := range buffer {
		vc.OpusSend <- buff
	}

	// Stop speaking
	vc.Speaking(false)

	// Sleep for a specificed amount of time before ending.
	time.Sleep(250 * time.Millisecond)

	// Disconnect from the provided voice channel.
	vc.Disconnect()

	return nil
}
