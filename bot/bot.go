package bot

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"

	"github.com/bwmarrin/discordgo"
)

// Bot defines the simple interface required to implement a bot
type Bot interface {
	// Run starts the bot with the given token
	// Returns a channel for if the server is closed, or an error if the bot fails to start
	Run(token string) (chan struct{}, error)
}

// discordBot is a simple implementation of the Bot interface, for Discord
type discordBot struct {
	// closer is a channel for listening if we've received a signal to shut down the bot
	closer chan os.Signal
}

// A lock and pointer for the singleton pattern
var lock = &sync.Mutex{}
var b *discordBot

// NewDiscordBot returns a new bot instance
func NewDiscordBot() *discordBot {
	// Uses a singleton pattern, as we only ever need one instance of the bot
	if b == nil {
		lock.Lock()
		defer lock.Unlock()
		if b == nil {
			b = &discordBot{}
		}
	}

	return b
}

func (b *discordBot) Run(token string) (chan struct{}, error) {
	// Set up a new Discord session and add our handler to it
	discord, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}

	discord.AddHandler(handleMessage)

	err = discord.Open()
	if err != nil {
		return nil, err
	}

	// Run the bot until we receive an interrupt
	fmt.Println("Bot running...")

	serverDown := make(chan struct{})
	b.closer = make(chan os.Signal, 1)
	signal.Notify(b.closer, os.Interrupt)

	// Wait for the closer to be called, then shut everything down
	go func() {
		<-b.closer
		discord.Close()
		close(serverDown)
	}()

	return serverDown, nil
}

func (b *discordBot) Close() {
	b.closer <- os.Interrupt
}

// card a struct for ArkhamDB card data
type card struct {
	// The URL of the card's image
	URL string `json:"url"`
}

// handleMessage parses and responds to incoming Discord messages
func handleMessage(discord *discordgo.Session, message *discordgo.MessageCreate) {
	// TODO - Could refactor this so the Discord & HTTP calls are their own clients, allowing for more testing
	fmt.Println("Message received:", message.Content)

	// Check if the message is from the bot
	if message.Author.ID == discord.State.User.ID {
		return
	}

	switch {
	case strings.Contains(message.Content, "!ping"):
		_, err := discord.ChannelMessageSend(message.ChannelID, "Pong! 🏓")
		if err != nil {
			fmt.Println("Error sending message:", err)
		}
	case strings.Contains(message.Content, "!card"):
		// Get the card ID from the message, and call the ArkhamDB API
		cardID := strings.Split(message.Content, " ")[1]

		resp, err := http.Get(fmt.Sprintf("https://arkhamdb.com/api/public/card/%s", cardID))
		if err != nil {
			log.Fatalln(err)
		}
		defer resp.Body.Close()

		// Read the body into the response struct
		var c card
		err = json.NewDecoder(resp.Body).Decode(&c)

		_, err = discord.ChannelMessageSend(message.ChannelID, c.URL)
		if err != nil {
			fmt.Println("Error sending message:", err)
		}
	}
}
