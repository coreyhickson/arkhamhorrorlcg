package bot

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"testing"
)

func TestNewDiscordBot(t *testing.T) {
	t.Run("NewDiscordBot only returns 1 single bot, no matter how much it's called", func(t *testing.T) {
		// A for loop that asynchronously calls NewDiscordBot 100 times
		firstBot := make(chan *discordBot)
		lastBot := make(chan *discordBot)

		// A go func that calls NewDiscordBot and puts it into firstBot
		go func() {
			firstBot <- NewDiscordBot()
		}()
		for i := 0; i < 100; i++ {
			go NewDiscordBot()
		}
		go func() {
			lastBot <- NewDiscordBot()
		}()

		f := <-firstBot
		l := <-lastBot
		if f != l {
			t.Errorf("NewDiscordBot() = %v, want %v", f, l)
		}
	})
}

func Test_discordBot_Run(t *testing.T) {
	// parse config.json file, get the token
	f, err := os.Open("../config.json")
	if err != nil {
		fmt.Println(err)
		t.Fatalf("Test_discordBot_Run() error = %v", err)
	}
	defer f.Close()

	bs, err := io.ReadAll(f)
	if err != nil {
		fmt.Println(err)
		t.Fatalf("Test_discordBot_Run() error = %v", err)
	}

	type token struct {
		Value string `json:"token"`
	}
	var testToken token
	err = json.Unmarshal(bs, &testToken)
	if err != nil {
		fmt.Println(err)
		t.Fatalf("Test_discordBot_Run() error = %v", err)
	}

	tests := []struct {
		name    string
		token   string
		wantErr bool
	}{
		{
			name:    "Run() returns an error if the token is empty",
			token:   "",
			wantErr: true,
		},
		{
			name:    "Run() returns no error if valid token and shuts down properly",
			token:   testToken.Value,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &discordBot{}

			serverDown, err := b.Run(tt.token)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			go func() {
				<-serverDown
				if (err != nil) != tt.wantErr {
					t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
				}
			}()
			go func() {
				b.Close()
			}()
		})
	}
}
