package messages

import "github.com/bwmarrin/discordgo"

// Returns true if already typiing
func StartTyping(alreadyTyping bool, s *discordgo.Session, channelId string) bool {
	if alreadyTyping {
		return true
	}

	s.ChannelTyping(channelId)

	return true
}
