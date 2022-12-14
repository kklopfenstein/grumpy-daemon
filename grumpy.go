package main

import (
	"flag"
	"log"
	"os"
	"os/signal"

	"github.com/bwmarrin/discordgo"
	"rawrippers.com/grumpy-daemon/first"
	"rawrippers.com/grumpy-daemon/game"
	"rawrippers.com/grumpy-daemon/joke"
	"rawrippers.com/grumpy-daemon/reaction"
	"rawrippers.com/grumpy-daemon/reminder"
	"rawrippers.com/grumpy-daemon/response"
	"rawrippers.com/grumpy-daemon/stable"
)

var (
	GuildID        = flag.String("guild", "", "Test guild ID. If not passed - bot registers commands globally")
	BotToken       = flag.String("token", "", "Bot access token")
	RemoveCommands = flag.Bool("rmcmd", true, "Remove all commands after shutdowning or not")
)

var s *discordgo.Session

func init() {
	flag.Parse()
}

func init() {
	var err error
	s, err = discordgo.New("Bot " + *BotToken)
	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
	}
}

var (
	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "joke",
			Description: "Tell a joke",
		},
		{
			Name:        "first",
			Description: "First of the month",
		},
		{
			Name:        "stable",
			Description: "Stable diffusion",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "prompt",
					Description: "Stable diffusion prompt",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "width",
					Description: "width",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "height",
					Description: "height",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "num_outputs",
					Description: "number of images",
					Required:    false,
					MaxValue:    4,
				},
				{
					Type:        discordgo.ApplicationCommandOptionNumber,
					Name:        "guidance_scale",
					Description: "guidance scale",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionNumber,
					Name:        "prompt_strength",
					Description: "prompt strength",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "num_inference_steps",
					Description: "number of inference steps",
					Required:    false,
				},
			},
		},
		{
			Name:        "adventure",
			Description: "play the Adventure text based game",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "command",
					Description: "command to send to Adventure",
					Required:    true,
				},
			},
		},
		{
			Name:        "reminder",
			Description: "set a channel reminder",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "message",
					Description: "message to post",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "when",
					Description: "e.g. 2022-10-29 08:43 -0400",
					Required:    true,
				},
			},
		},
		{
			Name:        "response",
			Description: "set a channel response",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "message",
					Description: "message to post",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "search",
					Description: "e.g. search that will trigger response",
					Required:    true,
				},
			},
		},
		{
			Name:        "list_reminders",
			Description: "list all channel reminders",
		},
		{
			Name:        "list_responses",
			Description: "list all channel responses",
		},
		{
			Name:        "delete_reminder",
			Description: "delete a reminder",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "message",
					Description: "message to delete",
					Required:    true,
				},
			},
		},
		{
			Name:        "delete_response",
			Description: "delete a response",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "search",
					Description: "message to delete (by search)",
					Required:    true,
				},
			},
		},
		{
			Name:        "reaction",
			Description: "set a channel message reaction",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "emoji",
					Description: "emoji reaction",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "search",
					Description: "e.g. search that will trigger reaction",
					Required:    true,
				},
			},
		},
		{
			Name:        "list_reactions",
			Description: "list all channel reactions",
		},
		{
			Name:        "delete_reaction",
			Description: "delete a reaction",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "search",
					Description: "reaction to delete (by search)",
					Required:    true,
				},
			},
		},
	}

	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"joke": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			joke.Joke(s, i)
		},
		"first": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			first.FirstOfTheMonth(s, i)
		},
		"stable": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			stable.Stable(s, i)
		},
		"adventure": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			game.Adventure(s, i)
		},
		"reminder": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			reminder.SetReminder(s, i)
		},
		"list_reminders": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			reminder.ListReminders(s, i)
		},
		"response": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			response.SetResponse(s, i)
		},
		"list_responses": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			response.ListResponses(s, i)
		},
		"delete_response": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			response.DeleteResponse(s, i)
		},
		"delete_reminder": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			reminder.DeleteReminder(s, i)
		},
		"reaction": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			reaction.SetReaction(s, i)
		},
		"list_reactions": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			reaction.ListReactions(s, i)
		},
		"delete_reaction": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			reaction.DeleteReaction(s, i)
		},
	}
)

func init() {
	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})
}

func main() {
	go reminder.Poll(s)
	go response.Load()
	go reaction.Load()

	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})
	s.AddHandler(response.MessageCreate)
	s.AddHandler(reaction.MessageCreate)
	err := s.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}

	log.Println("Adding commands...")

	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))
	for i, v := range commands {
		cmd, err := s.ApplicationCommandCreate(s.State.User.ID, *GuildID, v)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
		}
		registeredCommands[i] = cmd
	}

	defer s.Close()
	defer game.Stop()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop

	if *RemoveCommands {
		log.Println("Removing commands..")
		for _, v := range registeredCommands {
			err := s.ApplicationCommandDelete(s.State.User.ID, *GuildID, v.ID)
			if err != nil {
				log.Panicf("Cannot delete '%v' command: %v", v.Name, err)
			}
		}
	}

	log.Println("Gracefully shutting down.")
}
