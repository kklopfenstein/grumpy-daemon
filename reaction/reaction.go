package reaction

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/bwmarrin/discordgo"
	"rawrippers.com/grumpy-daemon/utils"
)

type reaction struct {
	EmojiID   string
	Search    string
	ChannelId string
}

var (
	mu        sync.Mutex
	reactions []*reaction
)

func Load() {
	read()
}

func SetReaction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: set(s, i),
		},
	})
}

func ListReactions(s *discordgo.Session, i *discordgo.InteractionCreate) {
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
	for _, r := range reactions {
		if r.ChannelId == m.ChannelID && utils.ContainsSearch(strings.ToLower(m.Content), strings.ToLower(r.Search)) {
			reg, _ := regexp.Compile("<(:.+:[0-9]+)>")
			customEmojis := reg.FindAllString(r.EmojiID, -1)
			for _, customEmoji := range customEmojis {
				submatches := reg.FindStringSubmatch(customEmoji)
				if len(submatches) == 2 {
					s.MessageReactionAdd(m.ChannelID, m.Reference().MessageID, submatches[1])
				}
			}
			otherEmojis := reg.ReplaceAllString(r.EmojiID, "")
			s.MessageReactionAdd(m.ChannelID, m.Reference().MessageID, otherEmojis)
		}
	}
	mu.Unlock()
}

func DeleteReaction(s *discordgo.Session, i *discordgo.InteractionCreate) {
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
	for index, r := range reactions {
		if r.ChannelId == i.ChannelID && r.Search == search {
			reactions[index] = reactions[len(reactions)-1]
			reactions = reactions[:len(reactions)-1]
			deleted = true
		}
	}
	mu.Unlock()

	if deleted {
		return fmt.Sprintf("<@%s> deleted reaction `%s`.", i.Member.User.ID, search)
	} else {
		return fmt.Sprintf("Could not find reaction `%s` to delete.", search)
	}
}

func list(i *discordgo.InteractionCreate) string {
	if len(i.ChannelID) == 0 {
		return "Where is this coming from?"
	}

	reactionsResp := ""

	mu.Lock()

	for _, r := range reactions {
		if r.ChannelId == i.ChannelID {
			reactionsResp = fmt.Sprintf("%s\n%s:\t\t%s", reactionsResp, r.Search, r.EmojiID)
		}
	}

	mu.Unlock()

	if len(reactionsResp) == 0 {
		return "no reactions"
	}

	return fmt.Sprintf("```%s```", reactionsResp)
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

	var emojiID string
	var search string

	if option, ok := optionMap["emoji"]; ok {
		emojiID = option.StringValue()
	} else {
		return "Message is required."
	}

	if option, ok := optionMap["search"]; ok {
		search = option.StringValue()
	} else {
		return "Search is required."
	}

	r := reaction{
		EmojiID:   emojiID,
		Search:    search,
		ChannelId: i.ChannelID,
	}

	mu.Lock()
	reactions = append(reactions, &r)
	write()
	mu.Unlock()

	return fmt.Sprintf("<@%s> set a reaction `%s` to `%s`. Use /list_reactions to see reactions.", i.Member.User.ID, emojiID, search)
}

func write() {
	createDirs()
	homedir := homeDir()
	file, err := json.MarshalIndent(&reactions, "", " ")

	if err != nil {
		log.Fatal(err)
	}

	err = os.WriteFile(fmt.Sprintf("%s/.grumpy/reactions/reactions.json", homedir), file, 0644)

	if err != nil {
		log.Fatal(err)
	}
}

func read() {
	createDirs()
	homedir := homeDir()
	file, err := os.ReadFile(fmt.Sprintf("%s/.grumpy/reactions/reactions.json", homedir))

	if err != nil {
		log.Printf("Could not open reactions: %s", err)
		return
	}

	err = json.Unmarshal(file, &reactions)

	log.Printf("loaded %d reactions", len(reactions))

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

	path := fmt.Sprintf("%s/.grumpy/reactions/", homedir)
	err := os.MkdirAll(path, os.ModePerm)

	if err != nil {
		log.Fatal(err)
	}
}
