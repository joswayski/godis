package files

import (
	"io"
	"log/slog"
	"net/http"
	"sync"

	"github.com/bwmarrin/discordgo"
)

func GetFilesInMessage(m *discordgo.MessageCreate) (files []*discordgo.File, closers []io.Closer) {

	var fileMu sync.Mutex
	var fileWg sync.WaitGroup

	for _, attachment := range m.Attachments {
		fileWg.Add(1)
		go func(atch *discordgo.MessageAttachment) {
			defer fileWg.Done()
			response, err := http.Get(atch.URL)
			if err != nil {
				slog.Error("Error downloading file", "attachment", atch)
				return
			}

			if response.StatusCode != http.StatusOK {
				response.Body.Close()
				slog.Error("Non 200 status code downloading file", "status", response.Status, "attachment", atch)
				return
			}

			fileMu.Lock()
			closers = append(closers, response.Body)
			files = append(files, &discordgo.File{
				Name:        atch.Filename,
				ContentType: atch.ContentType,
				Reader:      response.Body,
			})
			fileMu.Unlock()
		}(attachment)
	}

	fileWg.Wait()

	return files, closers
}
