package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type (
	// SlackResponse is a struct used for responses
	SlackResponse struct {
		Text string `json:"text"`
	}
	// SlackHandler is a request handler
	SlackHandler struct {
	}

	SynoResponse struct {
		Word     string   `json:"word"`
		Synonyms []string `json:"synonyms"`
	}
)

func (h *SlackHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var buf bytes.Buffer
	buf.ReadFrom(r.Body)
	q, err := url.ParseQuery(buf.String())
	if err != nil {
		log.Printf("Error parsing query (%v): %s\n", r.Body, err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if q.Get("token") != os.Getenv("BOT_TOKEN") {
		log.Printf("Invalid token received: %s\n", q.Get("token"))
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	text := strings.Replace(strings.Replace(q.Get("text"), "<@"+q.Get("user_id")+">", "", 1), "googlebot: ", "", 1)
	var respText string
	if text == "" {
		respText = "Y U so quiet?"
	} else {
		words := strings.Split(text, " ")
		var synonyms []string
		for _, word := range words {
			if synonym, err := h.GetSynonim(word); err != nil {
				log.Printf("Failed to get synonym for %s (%s)\n", word, err.Error())
			} else {
				synonyms = append(synonyms, synonym)
			}
		}
		respText = "So you're saying that " + strings.Join(synonyms, " ")
	}
	sr := SlackResponse{Text: respText}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sr)
}

func (h *SlackHandler) GetSynonim(word string) (string, error) {
	r, err := http.Get("http://workshop.x7467.com:1080/" + word)
	defer r.Body.Close()
	if err != nil {
		return "", err
	}

	if r.StatusCode == http.StatusNotFound {
		return "[" + word + "]", nil
	}

	s := &SynoResponse{}
	var buf bytes.Buffer
	buf.ReadFrom(r.Body)
	if err := json.Unmarshal(buf.Bytes(), s); err != nil {
		return "", err
	}

	return s.Synonyms[rand.Intn(len(s.Synonyms))], nil
}

func main() {
	var botAddr string
	if botAddr = os.Getenv("BOT_ADDR"); botAddr == "" {
		botAddr = "localhost:8080"
	}

	if os.Getenv("BOT_TOKEN") == "" {
		fmt.Println("BOT_TOKEN environment variable is not set! Exiting...")
		return
	}

	log.Fatal(http.ListenAndServe(botAddr, &SlackHandler{}))
}
