package game

import (
	"fmt"
	"log"
	"sync"

	"github.com/bwmarrin/discordgo"
)

var (
	mu        sync.Mutex
	adventure *GameProc
)

func Adventure(s *discordgo.Session, i *discordgo.InteractionCreate) {
	mu.Lock()
	if adventure == nil || !adventure.started {
		log.Print("Starting a new game")
		adventure = New("adventure")
	} else {
		log.Print("Game is already started")
	}
	mu.Unlock()

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: adventureExecute(s, i),
		},
	})
}

func adventureExecuteAndRespond(s *discordgo.Session, username string, channelId string, command string) {
	mu.Lock()
	response := adventure.Execute(command)
	mu.Unlock()

	s.ChannelMessageSend(channelId, response)
}

func adventureExecute(s *discordgo.Session, i *discordgo.InteractionCreate) string {
	if i.Member == nil || i.Member.User == nil {
		return "Who are you?"
	}

	if len(i.ChannelID) == 0 {
		return "Where is this coming from?"
	}

	username := fmt.Sprintf("<@%s>", i.Member.User.ID)
	channelId := i.ChannelID

	options := i.ApplicationCommandData().Options

	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
	for _, opt := range options {
		optionMap[opt.Name] = opt
	}

	if option, ok := optionMap["command"]; ok {
		command := option.StringValue()
		go adventureExecuteAndRespond(s, username, channelId, command)
		return fmt.Sprintf("%s sent '%s'", username, command)
	} else {
		return "You broke it."
	}
}

func Stop() {
	mu.Lock()
	if adventure != nil && adventure.started {
		adventure.Stop()
	}
	mu.Unlock()
}
