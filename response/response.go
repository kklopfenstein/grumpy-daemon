package response

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/bwmarrin/discordgo"
)

type response struct {
	Message   string
	Search    string
	ChannelId string
}

var (
	mu        sync.Mutex
	responses []*response
)

func Load() {
	read()
}

func SetResponse(s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: set(s, i),
		},
	})
}

func ListResponses(s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: list(i),
		},
	})
}

func MessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	if len(m.ChannelID) == 0 {
		return
	}

	mu.Lock()
	for _, r := range responses {
		if r.ChannelId == m.ChannelID && strings.Contains(strings.ToLower(m.Content), strings.ToLower(r.Search)) {
			s.ChannelMessageSendReply(m.ChannelID, r.Message, m.Reference())
		}
	}
	mu.Unlock()
}

func DeleteResponse(s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: delete(s, i),
		},
	})
}

func delete(s *discordgo.Session, i *discordgo.InteractionCreate) string {
	if i.Member == nil || i.Member.User == nil {
		return "Who are you?"
	}

	if len(i.ChannelID) == 0 {
		return "Where is this coming from?"
	}

	options := i.ApplicationCommandData().Options

	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
	for _, opt := range options {
		optionMap[opt.Name] = opt
	}

	var search string

	if option, ok := optionMap["search"]; ok {
		search = option.StringValue()
	} else {
		return "search is required."
	}

	deleted := false
	mu.Lock()
	for index, r := range responses {
		if r.ChannelId == i.ChannelID && r.Search == search {
			responses[index] = responses[len(responses)-1]
			responses = responses[:len(responses)-1]
			deleted = true
		}
	}
	mu.Unlock()

	if deleted {
		return fmt.Sprintf("<@%s> deleted response `%s`.", i.Member.User.ID, search)
	} else {
		return fmt.Sprintf("Could not find response `%s` to delete.", search)
	}
}

func list(i *discordgo.InteractionCreate) string {
	if len(i.ChannelID) == 0 {
		return "Where is this coming from?"
	}

	responsesResp := ""

	mu.Lock()

	for _, r := range responses {
		if r.ChannelId == i.ChannelID {
			responsesResp = fmt.Sprintf("%s\n%s:\t\t%s", responsesResp, r.Search, r.Message)
		}
	}

	mu.Unlock()

	if len(responsesResp) == 0 {
		return "no responses"
	}

	return fmt.Sprintf("```%s```", responsesResp)
}

func set(s *discordgo.Session, i *discordgo.InteractionCreate) string {
	if i.Member == nil || i.Member.User == nil {
		return "Who are you?"
	}

	if len(i.ChannelID) == 0 {
		return "Where is this coming from?"
	}

	options := i.ApplicationCommandData().Options

	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
	for _, opt := range options {
		optionMap[opt.Name] = opt
	}

	var message string
	var search string

	if option, ok := optionMap["message"]; ok {
		message = option.StringValue()
	} else {
		return "Message is required."
	}

	if option, ok := optionMap["search"]; ok {
		search = option.StringValue()
	} else {
		return "Search is required."
	}

	r := response{
		Message:   message,
		Search:    search,
		ChannelId: i.ChannelID,
	}

	mu.Lock()
	responses = append(responses, &r)
	write()
	mu.Unlock()

	return fmt.Sprintf("<@%s> set a response `%s` to `%s`. Use /list_responses to see responses.", i.Member.User.ID, message, search)
}

func write() {
	createDirs()
	homedir := homeDir()
	file, err := json.MarshalIndent(&responses, "", " ")

	if err != nil {
		log.Fatal(err)
	}

	err = os.WriteFile(fmt.Sprintf("%s/.grumpy/responses/responses.json", homedir), file, 0644)

	if err != nil {
		log.Fatal(err)
	}
}

func read() {
	createDirs()
	homedir := homeDir()
	file, err := os.ReadFile(fmt.Sprintf("%s/.grumpy/responses/responses.json", homedir))

	if err != nil {
		log.Printf("Could not open responses: %s", err)
		return
	}

	err = json.Unmarshal(file, &responses)

	log.Printf("loaded %d responses", len(responses))

	if err != nil {
		log.Fatal(err)
	}
}

func homeDir() string {
	homedir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	return homedir
}

func createDirs() {
	homedir := homeDir()

	path := fmt.Sprintf("%s/.grumpy/responses/", homedir)
	err := os.MkdirAll(path, os.ModePerm)

	if err != nil {
		log.Fatal(err)
	}
}
