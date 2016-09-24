package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
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

	var text string
	t := q.Get("text")
	if strings.Contains(t, "jest fajny") {
		text = "No raczej!"
	} else if strings.Contains(t, "jest głupi") {
		text = "Chyba ty"
	} else {
		text := strings.Replace(q.Get("text"), q.Get("trigger_word"), "", 1)
	}
	sr := SlackResponse{Text: fmt.Sprintf("Text sent: %s by user %s", text, q.Get("user_name"))}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sr)
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
