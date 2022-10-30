package joke

import (
	"bufio"
	"log"
	"math/rand"
	"os"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func Joke(s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: randomJoke(),
		},
	})
}

func randomJoke() string {
	file, err := os.Open("data/jokes.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var jokes []string

	for scanner.Scan() {
		jokes = append(jokes, scanner.Text())
	}

	joke := jokes[rand.Intn(len(jokes))]
	joke = strings.Replace(joke, "<>", "\n", -1)

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return joke
}
