package messages

import "github.com/bwmarrin/discordgo"

func GetUsername(m *discordgo.Message) string {
	nameToUse := ""
	// Use the server nickname if it's there
	if m.Member != nil {
		nameToUse = m.Member.Nick

		if nameToUse == "" && m.Member.User != nil {
			nameToUse = m.Member.User.GlobalName
		}
	}

	if nameToUse == "" && m.Author != nil {
		// This is always there
		nameToUse = m.Author.Username
	}

	return nameToUse
}
