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
	"sort"
	"strings"
	"time"
)

type (
	// SlackResponse is a struct used for responses
	SlackResponse struct {
		Text string `json:"text"`
	}
	// SlackHandler is a request handler
	SlackHandler struct {
	}
	// SynoResponse is a response from the synonym service
	SynoResponse struct {
		Word     string   `json:"word"`
		Synonyms []string `json:"synonyms"`
	}
	synonym struct {
		Id   int
		Word string
	}
	errorS struct {
		Id  int
		Err error
	}
	byId []synonym
)

func (h *SlackHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer log.Printf("Processing took %v\n", time.Since(start))
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

	text := h.getCleanText(q)
	respText := h.getTranslated(text)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(SlackResponse{Text: respText})
}

func (h *SlackHandler) getCleanText(q url.Values) string {
	replacer := strings.NewReplacer("<@"+q.Get("user_id")+"> ", "", "googlebot: ", "")
	return replacer.Replace(q.Get("text"))
}

func (h *SlackHandler) getTranslated(text string) string {
	if text == "" {
		return "Y U so quiet?"
	}

	synC := make(chan synonym)
	errC := make(chan errorS)
	words := strings.Split(text, " ")
	var synonyms byId
	for i := 0; i < len(words); i++ {
		go h.getSynonym(words[i], i, synC, errC)
	}

	for i := 0; i < len(words); i++ {
		select {
		case synonym := <-synC:
			log.Printf("Got a synonym: %s -> %s\n", words[i], synonym.Word)
			synonyms = append(synonyms, synonym)
		case err := <-errC:
			log.Printf("Failed to get synonym for %s (%s)\n", words[i], err.Err.Error())
		}
	}
	sort.Sort(byId(synonyms))

	newWords := []string{}
	for _, s := range synonyms {
		newWords = append(newWords, s.Word)
	}

	return "So you're saying that " + strings.Join(newWords, " ")
}

func (h *SlackHandler) getSynonym(word string, i int, synC chan synonym, errC chan errorS) {
	r, err := http.Get("http://workshop.x7467.com:1080/" + word)
	defer r.Body.Close()
	if err != nil {
		errC <- errorS{i, err}
		return
	}

	if r.StatusCode == http.StatusNotFound {
		synC <- synonym{i, "[" + word + "]"}
		return
	}

	s := &SynoResponse{}
	var buf bytes.Buffer
	buf.ReadFrom(r.Body)
	if err := json.Unmarshal(buf.Bytes(), s); err != nil {
		errC <- errorS{i, err}
		return
	}

	synC <- synonym{i, s.Synonyms[rand.Intn(len(s.Synonyms))]}
}

func (s byId) Len() int {
	return len(s)
}

func (s byId) Less(i, j int) bool {
	return s[i].Id < s[j].Id
}

func (s byId) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
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
