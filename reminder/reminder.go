package reminder

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/bcampbell/fuzzytime"
	"github.com/bwmarrin/discordgo"
)

type event struct {
	Message   string
	When      string
	Next      time.Time
	ChannelId string
}

var (
	mu     sync.Mutex
	events []*event
)

const format = "2006-01-02T15:04-07:00"

func SetReminder(s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: set(s, i),
		},
	})
}

func ListReminders(s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: list(i),
		},
	})
}

func Poll(s *discordgo.Session) {
	mu.Lock()

	read()

	mu.Unlock()

	for {
		mu.Lock()

		now := time.Now()

		toRemove := []int{}

		for i, event := range events {
			if event.Next.Before(now) {
				s.ChannelMessageSend(event.ChannelId, event.Message)
				toRemove = append(toRemove, i)
			}
		}

		newIndexMap := make(map[int]int)
		for _, removeIndex := range toRemove {
			if newIndex, ok := newIndexMap[removeIndex]; ok {
				events[newIndex] = events[len(events)-1]
				newIndexMap[len(events)-1] = newIndex
				events = events[:len(events)-1]
			} else {
				events[removeIndex] = events[len(events)-1]
				newIndexMap[len(events)-1] = removeIndex
				events = events[:len(events)-1]
			}
		}

		write()

		mu.Unlock()

		time.Sleep(500 * time.Millisecond)
	}
}

func DeleteReminder(s *discordgo.Session, i *discordgo.InteractionCreate) {
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

	var message string

	if option, ok := optionMap["message"]; ok {
		message = option.StringValue()
	} else {
		return "Message is required."
	}

	deleted := false
	mu.Lock()
	for index, r := range events {
		if r.ChannelId == i.ChannelID && r.Message == message {
			events[index] = events[len(events)-1]
			events = events[:len(events)-1]
			deleted = true
		}
	}
	mu.Unlock()

	if deleted {
		return fmt.Sprintf("<@%s> deleted reminder `%s`.", i.Member.User.ID, message)
	} else {
		return fmt.Sprintf("Could not find reminder `%s` to deleted.", message)
	}
}

func list(i *discordgo.InteractionCreate) string {
	if len(i.ChannelID) == 0 {
		return "Where is this coming from?"
	}

	eventsResponse := ""

	mu.Lock()

	for _, event := range events {
		if event.ChannelId == i.ChannelID {
			eventsResponse = fmt.Sprintf("%s\n%s\t\t%s", eventsResponse, event.When, event.Message)
		}
	}

	mu.Unlock()

	if len(eventsResponse) == 0 {
		return "no upcoming events"
	}

	return fmt.Sprintf("```%s```", eventsResponse)
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
	var when string

	if option, ok := optionMap["message"]; ok {
		message = option.StringValue()
	} else {
		return "Message is required."
	}

	if option, ok := optionMap["when"]; ok {
		when = option.StringValue()
	} else {
		return "When is required."
	}

	event, err := buildEvent(message, when, i.ChannelID)

	if err != nil {
		return fmt.Sprintf("I didn't understand that. Example date: %s", format)
	}

	mu.Lock()
	events = append(events, event)
	write()
	mu.Unlock()

	return fmt.Sprintf("<@%s> set a reminder `%s` at `%s`. Use /list_reminders to see reminders.", i.Member.User.ID, message, when)
}

func buildEvent(message string, when string, channelId string) (*event, error) {
	tokens := strings.Split(when, " ")

	if len(tokens) == 0 {
		return nil, fmt.Errorf("invalid token length")
	}

	extractedTime, _, err := fuzzytime.Extract(when)

	if err != nil {
		log.Print(err)
		return nil, fmt.Errorf("could not parse when")
	}

	parsed, err := time.Parse(format, extractedTime.ISOFormat())

	if err != nil {
		log.Print(err)
		return nil, fmt.Errorf("could not parse when")
	}

	event := event{
		Message:   message,
		When:      when,
		Next:      parsed,
		ChannelId: channelId,
	}

	return &event, nil
}

func write() {
	createDirs()
	homedir := homeDir()
	file, err := json.MarshalIndent(&events, "", " ")

	if err != nil {
		log.Fatal(err)
	}

	err = os.WriteFile(fmt.Sprintf("%s/.grumpy/events/events.json", homedir), file, 0644)

	if err != nil {
		log.Fatal(err)
	}
}

func read() {
	createDirs()
	homedir := homeDir()
	file, err := os.ReadFile(fmt.Sprintf("%s/.grumpy/events/events.json", homedir))

	if err != nil {
		log.Printf("Could not open events: %s", err)
		return
	}

	err = json.Unmarshal(file, &events)

	log.Printf("loaded %d events", len(events))

	if err != nil {
		log.Print(err)
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

	path := fmt.Sprintf("%s/.grumpy/events/", homedir)
	err := os.MkdirAll(path, os.ModePerm)

	if err != nil {
		log.Print(err)
	}
}
