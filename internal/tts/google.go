package tts

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type GoogleTTS struct {
	httpClient *http.Client
}

const baseUrl = "https://translate.google.com/translate_tts?ie=UTF-8&tl=fr&client=tw-ob&q="

func NewGoogleTTS() TTSProvider {
	return GoogleTTS{httpClient: &http.Client{Timeout: 15 * time.Second}}
}

func (gtts GoogleTTS) Synthesize(text string) ([]byte, error) {
	fullURL := baseUrl + url.QueryEscape(text)
	resp, err := gtts.httpClient.Get(fullURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("not a OK result")
	}

	return io.ReadAll(resp.Body)
}
