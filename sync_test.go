package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

var (
	authServer      *httptest.Server
	svcUpdateServer *httptest.Server
	svcNewServer    *httptest.Server
)

const successResponse = `
<?xml version='1.0' standalone='yes'?>
<response>
	<biblionumber>164442</biblionumber>
	<marcxml>
		<record>omitted</record>
	</marcxml>
	<status>ok</status>
</response>`

func init() {
	authServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := r.URL.Query().Get("userid")
		pass := r.URL.Query().Get("password")

		if user != "sync" || pass != "sync" {
			http.Error(w, "wrong userid or password", http.StatusForbidden)
			return
		}

		cookie := http.Cookie{HttpOnly: true, Path: "/", Secure: false,
			MaxAge: 0, Name: "CGISESSID", Value: "8655024ef41e104a1a2c58a6c744e69c"}
		http.SetCookie(w, &cookie)

	}))

	svcUpdateServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("CGISESSID")
		if err == nil && cookie.Value == "8655024ef41e104a1a2c58a6c744e69c" {
			w.Write([]byte(successResponse))
			return
		}
		http.Error(w, "unathorized", http.StatusForbidden)
	}))

	svcNewServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("CGISESSID")
		if err == nil && cookie.Value == "8655024ef41e104a1a2c58a6c744e69c" {
			w.Write([]byte(successResponse))
			return
		}
		http.Error(w, "unathorized", http.StatusForbidden)
	}))
}

func TestSvcAuth(t *testing.T) {
	_, err := svcAuth(authServer.URL, "sync", "wrong")
	if err == nil {
		t.Error("svcAuth: expected wrong password to result in an error")
	}

	jar, err := svcAuth(authServer.URL, "sync", "sync")
	if err != nil {
		t.Fatalf("svcAuth with correct user & pass result in error: %v", err)
	}
	u, err := url.Parse(authServer.URL)
	if err != nil {
		t.Fatal(err)
	}
	cookies := jar.Cookies(u)
	if len(cookies) != 1 {
		t.Fatalf("wanted 1 cookie, got %d", len(cookies))
	}

	if cookies[0].Name != "CGISESSID" || cookies[0].Value != "8655024ef41e104a1a2c58a6c744e69c" {
		t.Errorf("wanted session cookie from Koha, got something else: %v", cookies[0])
	}

}

func TestUpdatedManifestation(t *testing.T) {
	jar, err := svcAuth(authServer.URL, "sync", "sync")
	if err != nil {
		t.Fatal(err)
	}

	err = svcUpdate(svcUpdateServer.URL, jar, []byte("<marcxml> ... </marcxml>"), 164442)
	if err != nil {
		t.Fatal(err)
	}
}

func TestSyncNewManifesation(t *testing.T) {
	jar, err := svcAuth(authServer.URL, "sync", "sync")
	if err != nil {
		t.Fatal(err)
	}

	biblionr, err := svcNew(svcNewServer.URL, jar, []byte("<marcxml>...</marxml>"))
	if err != nil {
		t.Fatal(err)
	}

	if biblionr != 164442 {
		t.Errorf("expected biblionr 164442; got %d", biblionr)
	}
}
