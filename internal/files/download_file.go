package files

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

var downloadClient = &http.Client{
	Timeout: 30 * time.Second,
}

func DownloadFile(url string) ([]byte, error) {

	resp, err := downloadClient.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("non 200 status: %s", resp.Status)
	}

	return io.ReadAll(resp.Body)
}
