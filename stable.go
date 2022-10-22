package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func Stable(s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: StableGet(s, i),
		},
	})
}

func StableRun(s *discordgo.Session, username string, channelId string, prompt string) {
	message := ""
	match, _ := regexp.MatchString("^[a-zA-Z0-9\\s]+$", prompt)
	if !match {
		message = "Stop sending me garbage."
	} else {
		args := strings.Split(prompt, " ")
		var out bytes.Buffer
		command := exec.Command("stable", args...)
		command.Stdout = &out
		err := command.Run()
		if err != nil {
			message = "Oops! Maybe not so stable!"
		} else {
			message = out.String()
		}
	}

	s.ChannelMessageSend(channelId, message)
}

func StableGet(s *discordgo.Session, i *discordgo.InteractionCreate) string {
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

	if option, ok := optionMap["prompt"]; ok {
		prompt := option.StringValue()
		go StableRun(s, username, channelId, prompt)
		return fmt.Sprintf("Buildin' an image for \"%s\"", prompt)
	} else {
		return "You broke it."
	}
}
