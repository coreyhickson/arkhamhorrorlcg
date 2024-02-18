package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"arkhamhhorrorlcg/bot"
)

func main() {
	fmt.Println("Initializing the Arkham Horror LCG bot...")

	// parse config.json file, get the token
	f, err := os.Open("config.json")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer f.Close()

	bs, err := io.ReadAll(f)
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}

	type token struct {
		Value string `json:"token"`
	}
	var t token
	err = json.Unmarshal(bs, &t)
	if err != nil {
		fmt.Println(err)
		os.Exit(3)
	}

	fmt.Println("Successfully parsed the token from config.json file.")

	// Start the bot
	fmt.Println("Starting the bot...")

	b := bot.NewDiscordBot()
	serverDown, err := b.Run(t.Value)
	if err != nil {
		fmt.Println(err)
		os.Exit(4)
	}
	defer b.Close()

	<-serverDown

	fmt.Println("Shutting down the bot...")
}
