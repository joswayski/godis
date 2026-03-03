package webhooks

import (
	"log/slog"
	"sync"

	"github.com/bwmarrin/discordgo"
)

const (
	godisWebhookName = "Godis Webhook"
)

type webhookCache struct {
	itemsMu sync.RWMutex
	// Key is the channel ID
	items map[string]*discordgo.Webhook

	creatingMu sync.Mutex
	// Key is the guild ID
	creating map[string]*sync.Mutex
}

var wbhkCache = webhookCache{
	items:    make(map[string]*discordgo.Webhook),
	creating: make(map[string]*sync.Mutex),
}

func GetGodisWebhook(s *discordgo.Session, m *discordgo.MessageCreate) (*discordgo.Webhook, error) {
	// Check the cache first
	wbhkCache.itemsMu.RLock()
	webhook := wbhkCache.items[m.ChannelID]

	if webhook != nil {
		wbhkCache.itemsMu.RUnlock()
		return webhook, nil
	}
	wbhkCache.itemsMu.RUnlock()

	// Now we want to lock per guild so that we fetch all of the server's webhooks for it
	wbhkCache.creatingMu.Lock()
	guildLock := wbhkCache.creating[m.GuildID]
	if guildLock == nil {
		// Create the guild mutex
		wbhkCache.creating[m.GuildID] = &sync.Mutex{}
		guildLock = wbhkCache.creating[m.GuildID]
	}
	wbhkCache.creatingMu.Unlock()

	guildLock.Lock()
	// Unlock after we've populated the webhooks for the server
	defer guildLock.Unlock()

	// Check the cache again now that we're locked per guild/server just in case something else was running in the meantime
	wbhkCache.itemsMu.RLock()
	webhook = wbhkCache.items[m.ChannelID]
	if webhook != nil {
		wbhkCache.itemsMu.RUnlock()
		return webhook, nil
	}
	wbhkCache.itemsMu.RUnlock()

	// Now that we have a lock per guild/server, get all of the webhooks for the server that its in
	allWebhooks, err := s.GuildWebhooks(m.GuildID)
	if err != nil {
		slog.Error("Error retrieving all webhooks", "error", err.Error())
		return nil, err
	}

	wbhkCache.itemsMu.Lock()
	for _, wbhk := range allWebhooks {
		// If we have a webhook for a channel, set it
		if wbhk.Name == godisWebhookName {
			wbhkCache.items[wbhk.ChannelID] = wbhk
		}
	}
	godisWebhook := wbhkCache.items[m.ChannelID]
	wbhkCache.itemsMu.Unlock()

	// Check if we set it up above, if so, return it
	if godisWebhook != nil {
		return godisWebhook, nil
	}

	// Create our webhook for the first time
	newWebhook, err := s.WebhookCreate(m.ChannelID, godisWebhookName, "")
	if err != nil {
		slog.Error("Unable to create Godis webhook", "error", err.Error())
		return nil, err
	}

	// Add it to the cache
	wbhkCache.itemsMu.Lock()
	wbhkCache.items[m.ChannelID] = newWebhook
	wbhkCache.itemsMu.Unlock()

	return newWebhook, nil

}
