package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/go-telegram-bot-api/telegram-bot-api"
)

const (
	dictionaryAPIPrefix = "https://api.dictionaryapi.dev/api/v2/entries/en/"
)

type WordInfo []struct {
	Word      string     `json:"word"`
	Phonetics []struct{} `json:"phonetics"`
	Meanings  []struct {
		PartOfSpeech string `json:"partOfSpeech"`
		Definitions  []struct {
			Definition string        `json:"definition"`
			Example    string        `json:"example"`
			Synonyms   []string      `json:"synonyms"`
			Antonyms   []interface{} `json:"antonyms"`
		} `json:"definitions"`
	} `json:"meanings"`
}

func main() {
	token := os.Getenv("token")

	if token == "" {
		log.Fatal("No token for authorization")
		return
	}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal("Unable to connect to Telegram API", err)
		return
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	config := tgbotapi.UpdateConfig{}
	updates, err := bot.GetUpdatesChan(config)

	for update := range updates {
		receivedWord, err := getWordInfo(update.Message.Text)
		if err != nil {
			log.Println(err)

			errorMsg := "Unable to find the word. Check the spelling."
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, errorMsg)
			_, err = bot.Send(msg)
			if err != nil {
				log.Println("Unable to send message:", err)
				return
			}

			continue
		}

		wordParams := formatMessage(receivedWord)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, wordParams)
		_, err = bot.Send(msg)
		if err != nil {
			log.Println("Unable to send message:", err)
			return
		}
	}
}

func requestWord(url string) (*http.Response, error) {
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	request.Header.Set("Accept", "application/json")

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return response, err
	}

	return response, nil
}

func getWordInfo(word string) (*WordInfo, error) {
	var wordInfo WordInfo

	response, err := requestWord(dictionaryAPIPrefix + url.QueryEscape(word))
	if err != nil {
		return nil, errors.New("Unable to make request:" + err.Error())
	}

	if response.StatusCode == 200 {
		defer response.Body.Close()

		data, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return nil, errors.New("Unable to read response:" + err.Error())
		}

		err = json.Unmarshal(data, &wordInfo)
		if err != nil {
			return nil, errors.New("Unable to unmarshal JSON:" + err.Error())
		}

		return &wordInfo, nil
	}

	return nil, errors.New("Unexpected response")
}

func formatMessage(receivedWord *WordInfo) string {
	var wordParams string

	for _, value := range *receivedWord {
		for _, meaning := range value.Meanings {
			wordParams += "Part of speech: " + meaning.PartOfSpeech + "\n"

			for _, def := range meaning.Definitions {
				wordParams += "Definition: " + def.Definition + "\n"
				wordParams += "Example: " + def.Example + "\n"
			}
		}
	}

	return wordParams
}
