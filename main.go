package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type DeepLResponse struct {
	Translations []struct {
		Text string `json:"text"`
	} `json:"translations"`
}

func main() {
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	deeplApiKey := os.Getenv("DEEPL_API_KEY")

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		translatedText, err := translateToUkrainian(update.Message.Text, deeplApiKey)
		if err != nil {
			log.Println("Translation error:", err)
			continue
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, translatedText)
		bot.Send(msg)
	}
}

func translateToUkrainian(text, apiKey string) (string, error) {
	deeplURL := "https://api-free.deepl.com/v2/translate"
	reqBody := fmt.Sprintf("auth_key=%s&text=%s&target_lang=UK", apiKey, text)
	reqBodyReader := strings.NewReader(reqBody)

	req, err := http.NewRequest("POST", deeplURL, reqBodyReader)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var deeplResp DeepLResponse
	if err := json.NewDecoder(resp.Body).Decode(&deeplResp); err != nil {
		return "", err
	}

	if len(deeplResp.Translations) > 0 {
		return deeplResp.Translations[0].Text, nil
	}
	return "", fmt.Errorf("no translation found")
}
