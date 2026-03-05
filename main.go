package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/joswayski/godis/internal/config"
	"github.com/joswayski/godis/internal/godis"
	"github.com/joswayski/godis/internal/logger"
)

func main() {
	logger.Init()
	slog.Info("Godis is starting...")

	bot := &godis.Godis{
		Config: config.GetConfig(),
	}

	discord, err := discordgo.New("Bot " + bot.Config.DiscordToken)
	if err != nil {
		slog.Error("Error ocurred loading discord", "error", err.Error())
		os.Exit(1)
	}

	discord.AddHandler(bot.HandleMessage)

	discord.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentMessageContent

	err = discord.Open()
	defer discord.Close()

	if err != nil {
		slog.Error("Error opening websocket", "error", err.Error())
		os.Exit(1)
	}

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, os.Interrupt)
	<-sc

	slog.Info("Shutting down Godis!")

}
