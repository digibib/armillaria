package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
)

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

// 3 kinds of sync: create,update,delete

// 1. Manifestation created
// -> POST /svc/new_bib
//    hold on to biblionumber and store in RDF:
//    -> insert { <manifestation> armillaria:kohaId biblionumber }
func syncUpdatedManifestation(rec *marcRecord, uri string) error {
	return nil
}

// 2. Manifestation updated
// -> POST /svc/{biblionumber}
func syncNewManifestation(rec *marcRecord, uri string) (string, error) {
	return "biblionumber", nil
}

// 3. Manifestation deleted
// -> DELETE /svc/{biblionumber}
func syncDeletedManifestation(biblio string) error {
	return errors.New("svc API does not support deletion yet")
}
