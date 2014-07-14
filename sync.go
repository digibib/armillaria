package main

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
)

var svcNotAuthorized = errors.New("unauthorized")

// svcResponse represents the XML response we get from the svc API when updating
// or creating a biblio.
type svcResponse struct {
	XMLName xml.Name `xml:"response"`
	Biblio  int      `xml:"biblionumber"`
	Status  string   `xml:"status"`
	Error   string   `xml:"error"`
	// We are not interested in the marcxml record,
	// so we ignore the rest of the response.
}

// syncKohaAuth attemps to authenticate against Koha's /svc endpoint.
// If sucessfull, it returns a cookiejar with the authenticated cookie.
func syncKohaAuth(kohaPath, user, pass string) (http.CookieJar, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("%s/cgi-bin/koha/svc/authentication?userid=%s&password=%s", kohaPath, user, pass)
	client := &http.Client{Jar: jar}
	resp, err := client.Get(path)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("unauthorized")
	}
	return client.Jar, nil
}

// 1. Manifestation created
// -> POST /svc/new_bib
//    hold on to biblionumber and store in RDF:
//    -> insert { <manifestation> armillaria:kohaId biblionumber }
func syncNewManifestation(kohaPath string, jar http.CookieJar, rec marcRecord) (int, error) {
	return 122345, nil
}

// 2. Manifestation updated
// -> POST /svc/{biblionumber}
func syncUpdatedManifestation(kohaPath string, jar http.CookieJar, marc []byte, biblio int) error {
	path := fmt.Sprintf("%s/cgi-bin/koha/svc/bib/%d", kohaPath, biblio)
	client := &http.Client{Jar: jar}
	req, err := http.NewRequest("GET", path, bytes.NewReader(marc))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "text/xml")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		switch resp.StatusCode {
		case 403:
			return svcNotAuthorized
		case 404:
			return errors.New("not found")
		case 400:
			return errors.New("bad request")
		default:
			// Should never be reached.
			// There are no other errors, according to the svc source code.
			return errors.New("unknown error")
		}
	}

	var svcRes svcResponse
	d := xml.NewDecoder(resp.Body)

	if err := d.Decode(&svcRes); err != nil {
		return err
	}
	if svcRes.Status != "ok" {
		return errors.New(svcRes.Error)
	}

	return nil
}

// 3. Manifestation deleted
// -> DELETE /svc/{biblionumber}
func syncDeletedManifestation(kohaPath string, jar http.CookieJar, biblio string) error {
	return errors.New("svc API does not support deletion yet")
}
