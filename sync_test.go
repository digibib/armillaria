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
)

const updateResponse = `<?xml version='1.0' standalone='yes'?>
<response>
  <biblionumber>164442</biblionumber>
  <marcxml>
<record xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"  xsi:schemaLocation="http://www.loc.gov/MARC21/slim http://www.loc.gov/standards/marcxml/schema/MARC21slim.xsd" xmlns="http://www.loc.gov/MARC21/slim">
  <leader>00679cam a22002291  4500</leader>
  <controlfield tag="000">c</controlfield>
  <controlfield tag="008">                      a          00nob  </controlfield>
  <datafield tag="015" ind1=" " ind2=" ">
    <subfield code="a">241</subfield>
    <subfield code="b">BibliofilID</subfield>
  </datafield>
  <datafield tag="019" ind1=" " ind2=" ">
    <subfield code="b">l</subfield>
  </datafield>
  <datafield tag="020" ind1=" " ind2=" ">
    <subfield code="a">8205183546</subfield>
    <subfield code="c">Nkr199,99.</subfield>
  </datafield>
  <datafield tag="100" ind1=" " ind2="0">
    <subfield code="a">Brox, Ottar</subfield>
    <subfield code="d">1932-</subfield>
    <subfield code="j">n.</subfield>
    <subfield code="3">10033300</subfield>
  </datafield>
  <datafield tag="245" ind1="1" ind2="0">
    <subfield code="a">Bygdenæringene kan saktens bli lønnsomme</subfield>
    <subfield code="c">Ottar Brox</subfield>
  </datafield>
  <datafield tag="260" ind1=" " ind2=" ">
    <subfield code="a">Oslo</subfield>
    <subfield code="b">Gyldendal</subfield>
    <subfield code="c">1989</subfield>
  </datafield>
  <datafield tag="300" ind1=" " ind2=" ">
    <subfield code="a">187 s.</subfield>
    <subfield code="b">fig.</subfield>
  </datafield>
  <datafield tag="999" ind1=" " ind2=" ">
    <subfield code="c">164442</subfield>
    <subfield code="d">164442</subfield>
  </datafield>
</record>
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
			w.Write([]byte(updateResponse))
			return
		}
		http.Error(w, "unathorized", http.StatusForbidden)
	}))
}

func TestSyncKohaAuth(t *testing.T) {
	_, err := syncKohaAuth(authServer.URL, "sync", "wrong")
	if err == nil {
		t.Error("syncKohaAuth: expected wrong password to result in an error")
	}

	jar, err := syncKohaAuth(authServer.URL, "sync", "sync")
	if err != nil {
		t.Fatal("syncKohaAuth with correct user & pass result in error: %v", err)
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
	jar, err := syncKohaAuth(authServer.URL, "sync", "sync")
	if err != nil {
		t.Fatal(err)
	}

	err = syncUpdatedManifestation(svcUpdateServer.URL, jar, []byte("<marcxml> ... </marcxml>"), 164442)
	if err != nil {
		t.Fatalf("%v", err)
	}

}
