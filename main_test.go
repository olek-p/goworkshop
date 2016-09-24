package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestSlackHandler(t *testing.T) {
	os.Setenv("BOT_TOKEN", "123")
	r, _ := http.NewRequest("POST", "dada", strings.NewReader("token=123&user_name=Olek&trigger_word=test&text=test%20testowy"))
	t.Log(r)
	w := httptest.NewRecorder()
	h := &SlackHandler{}
	h.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatal(w.Code, w.Body.String())
	}
}
