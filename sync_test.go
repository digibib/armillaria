package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

var authServer *httptest.Server

func init() {
	authServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := r.URL.Query().Get("userid")
		pass := r.URL.Query().Get("password")

		if user != "sync" || pass != "sync" {
			http.Error(w, "missing required parameter: query", http.StatusForbidden)
			return
		}

		cookie := http.Cookie{HttpOnly: true, Path: "/", Secure: false,
			MaxAge: 0, Name: "CGISESSID", Value: "8655024ef41e104a1a2c58a6c744e69c"}
		http.SetCookie(w, &cookie)

	}))
}

func TestSyncKohaAuth(t *testing.T) {
	_, err := syncKohaAuth(authServer.URL, "sync", "wrong")
	if err == nil {
		t.Error("syncKohaAuth: expected wrong password to result in an error")
	}

	jar, err := syncKohaAuth(authServer.URL, "sync", "sync")
	if err != nil {
		t.Error("syncKohaAuth with correct user & pass result in error: %v", err)
	}
	u, err := url.Parse(authServer.URL)
	if err != nil {
		t.Fatal(err)
	}
	cookies := jar.Cookies(u)
	if len(cookies) != 1 {
		t.Fatal("wanted 1 cookie, got %d", len(cookies))
	}

	if cookies[0].Name != "CGISESSID" || cookies[0].Value != "8655024ef41e104a1a2c58a6c744e69c" {
		t.Errorf("wanted session cookie from Koha, got something else: %v", cookies[0])
	}

}

func TestUpdatedManifestation(t *testing.T) {
	cfg = &config{
		KohaPath:     "localhost:1234",
		KohaSyncUser: "sync",
		KohaSyncPass: "sync",
	}
}
