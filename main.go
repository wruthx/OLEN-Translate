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

		text := update.Message.Text
		if strings.HasPrefix(text, "/") {
			parts := strings.SplitN(text, " ", 2)
			if len(parts) < 2 {
				continue
			}
			langCode := strings.TrimPrefix(parts[0], "/")
			textToTranslate := parts[1]

			targetLang := getTargetLang(langCode)

			if targetLang == "" {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Unsupported language code!")
				bot.Send(msg)
				continue
			}

			translatedText, err := translateText(textToTranslate, targetLang, deeplApiKey)
			if err != nil {
				log.Println("Translation error:", err)
				continue
			}

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, translatedText)
			bot.Send(msg)
		} else {
			translatedText, err := translateText(text, "UK", deeplApiKey)
			if err != nil {
				log.Println("Translation error:", err)
				continue
			}

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, translatedText)
			bot.Send(msg)
		}
	}
}

func getTargetLang(prefix string) string {
	switch prefix {
	case "en":
		return "EN"
	case "tr":
		return "TR"
	case "de":
		return "DE"
	case "es":
		return "ES"
	case "fr":
		return "FR"
	case "sv":
		return "SV"
	default:
		return ""
	}
}

// translateText translates the input text to the specified target language using the DeepL API
func translateText(text, targetLang, apiKey string) (string, error) {
	deeplURL := "https://api-free.deepl.com/v2/translate"
	reqBody := fmt.Sprintf("auth_key=%s&text=%s&target_lang=%s", apiKey, text, targetLang)
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
