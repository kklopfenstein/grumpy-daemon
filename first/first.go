package first

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

func FirstOfTheMonth(s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: postFirst(i),
		},
	})
}

func postFirst(i *discordgo.InteractionCreate) string {
	if i.Member != nil && i.Member.User != nil {
		username := fmt.Sprintf("<@%s>", i.Member.User.ID)
		return fmt.Sprintf("%s says wake up!\nhttps://www.youtube.com/watch?v=4j_cOsgRY7w", username)
	} else {
		return "Wake up!\nhttps://www.youtube.com/watch?v=4j_cOsgRY7w"
	}
}
