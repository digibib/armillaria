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

// syncNewManifestation takes the given marcxml record and send it with a
//  POST request to /svc/new_bib. It returns the biblionumber it got assigned fro Koha.
func syncNewManifestation(kohaPath string, jar http.CookieJar, marc []byte) (int, error) {
	path := fmt.Sprintf("%s/cgi-bin/koha/svc/new_bib", kohaPath)
	client := &http.Client{Jar: jar}
	req, err := http.NewRequest("POST", path, bytes.NewReader(marc))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "text/xml")
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		switch resp.StatusCode {
		case 403:
			return 0, svcNotAuthorized
		case 404:
			return 0, errors.New("not found")
		case 400:
			return 0, errors.New("bad request")
		default:
			// Should never be reached.
			// There are no other errors, according to the svc source code.
			return 0, errors.New("unknown error")
		}
	}

	var svcRes svcResponse
	d := xml.NewDecoder(resp.Body)

	if err := d.Decode(&svcRes); err != nil {
		return 0, err
	}
	if svcRes.Status != "ok" {
		return 0, errors.New(svcRes.Error)
	}
	return svcRes.Biblio, nil
}

// syncUpdatedManifestation takes the given marcxml record and send it with a
// POST request to the /svc/{biblio} endpoint.
func syncUpdatedManifestation(kohaPath string, jar http.CookieJar, marc []byte, biblio int) error {
	path := fmt.Sprintf("%s/cgi-bin/koha/svc/bib/%d", kohaPath, biblio)
	client := &http.Client{Jar: jar}
	req, err := http.NewRequest("POST", path, bytes.NewReader(marc))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "text/xml; charset=utf-8")
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

// syncDeletedManifestation sends a DELETE request to /svc/biblio/{biblionr}
func syncDeletedManifestation(kohaPath string, jar http.CookieJar, biblio int) error {
	path := fmt.Sprintf("%s/cgi-bin/koha/svc/bib/%d", kohaPath, biblio)
	client := &http.Client{Jar: jar}
	req, err := http.NewRequest("DELETE", path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "text/xml; charset=utf-8")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		var svcRes svcResponse
		d := xml.NewDecoder(resp.Body)

		if err := d.Decode(&svcRes); err != nil {
			return err
		}
		return errors.New(svcRes.Error)
	}

	return nil
}
