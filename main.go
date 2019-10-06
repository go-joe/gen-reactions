package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/kyokomi/emoji"
)

const downloadURL = "https://www.webfx.com/tools/emoji-cheat-sheet/"

type EmojiGroup struct {
	Name   string
	Emojis []*Emoji
}

type Emoji struct {
	Name string
	Code string
}

func main() {
	resp, err := http.Get(downloadURL)
	if err != nil {
		log.Fatalf("Failed to download %q: %v", downloadURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Received bad HTTP status code from %q: %s", downloadURL, resp.Status)
	}

	emojis, err := parse(resp.Body)
	if err != nil {
		log.Fatalf("Failed to parse file: %v", err)
	}

	enrich(emojis)

	err = generateCode(emojis)
	if err != nil {
		log.Fatalf("Failed to generate code: %v", err)
	}
}

func enrich(groups []*EmojiGroup) {
	codeMap := emoji.CodeMap()
	for _, group := range groups {
		for _, emoji := range group.Emojis {
			key := fmt.Sprintf(":%s:", emoji.Name)
			emoji.Code = codeMap[key] // may be empty string if not found
		}
	}
}
