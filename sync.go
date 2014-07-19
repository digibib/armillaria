package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"strconv"

	"github.com/digibib/armillaria/sparql"
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

func syncCreateResource(uri string) (int, bool, error) {
	r, err := db.Query(fmt.Sprintf(resourceQuery, cfg.RDFStore.DefaultGraph, uri, uri, uri))
	if err != nil {
		return 0, true, fmt.Errorf("db.Query: %v", err.Error())
	}

	var res *sparql.Results
	var profile string
	err = json.Unmarshal(r, &res)
	if err != nil {
		return 0, true, fmt.Errorf("parse SPARQL response: %v", err.Error())
	}

	if len(res.Results.Bindings) == 0 {
		return 0, false, errors.New("cannot sync non-existing resource to Koha")
	}

	for _, b := range res.Results.Bindings {
		if b["p"].Value == "armillaria://internal/profile" {
			profile = b["o"].Value
			break
		}
	}

	if profile != "manifestation" {
		return 0, false, errors.New("only manifestations are synced to Koha")
	}

	// Make sure we are authenticated to Koha
	if kohaCookies == nil {
		kohaCookies, err = syncKohaAuth(cfg.KohaPath, cfg.KohaSyncUser, cfg.KohaSyncPass)
		if err != nil {
			return 0, true, fmt.Errorf("authenticate to Koha: %v", err.Error())
		}
	}

	// Generate MARCXML record of RDF resource
	r, err = db.Query(fmt.Sprintf(queryRDF2MARC, cfg.RDFStore.DefaultGraph, uri, uri))
	if err != nil {
		return 0, true, fmt.Errorf("db.Query: %v", err.Error())
	}

	err = json.Unmarshal(r, &res)
	if err != nil {
		return 0, true, fmt.Errorf("parse SPARQL response: %v", err.Error())
	}

	rec, err := convertRDF2MARC(*res)
	if err != nil {
		return 0, true, fmt.Errorf("convertRDF2MARC: %v", err.Error())
	}

	marc, err := xml.Marshal(rec)
	if err != nil {
		return 0, true, fmt.Errorf("MARC xml marhsal: %v", err.Error())
	}

	bibnr, err := syncNewManifestation(cfg.KohaPath, kohaCookies, marc)
	if err != nil {
		return 0, true, fmt.Errorf("syncNewManifestation: %v", err.Error())
	}

	// store the koha id as property on the RDF resource
	r, err = db.Query(fmt.Sprintf(insertKohaIDQuery, cfg.RDFStore.DefaultGraph, uri, uri, bibnr, uri))
	if err != nil {
		return 0, true, fmt.Errorf("db.Query: %v", err.Error())
	}

	if bytes.Index(r, []byte("1 (or less) triples")) == -1 {
		// TODO now what? retry will create another duplicate biblio record in Koha..
		if err != nil {
			return 0, false, fmt.Errorf("insert koha ID on resource: %v", err.Error())
		}
	}

	return bibnr, false, nil
}

func syncUpdateResource(uri string, biblionr int) (bool, error) {
	r, err := db.Query(fmt.Sprintf(resourceQuery, cfg.RDFStore.DefaultGraph, uri, uri, uri))
	if err != nil {
		return true, fmt.Errorf("db.Query: %v", err.Error())
	}

	var res *sparql.Results
	var profile string
	err = json.Unmarshal(r, &res)
	if err != nil {
		return true, fmt.Errorf("parse SPARQL response: %v", err.Error())
	}

	if len(res.Results.Bindings) == 0 {
		return false, errors.New("cannot sync non-existing resource to Koha")
	}

	var bibnrStr string
	var bibnr int
	for _, b := range res.Results.Bindings {
		if b["p"].Value == "armillaria://internal/profile" {
			profile = b["o"].Value
		}
		if b["p"].Value == "armillaria://internal/kohaID" {
			bibnrStr = b["o"].Value
		}
	}

	if bibnrStr != "" {
		bibnr, err = strconv.Atoi(bibnrStr)
		if err != nil {
			return false, errors.New("kohaID on resource is not an integer")
		}
	}

	if bibnr == 0 {
		return false, errors.New("cannot sync to Koha; missing kohaID on resource")
	}

	if profile != "manifestation" {
		return false, errors.New("only manifestations are synced to Koha")
	}

	// Make sure we are authenticated to Koha
	if kohaCookies == nil {
		kohaCookies, err = syncKohaAuth(cfg.KohaPath, cfg.KohaSyncUser, cfg.KohaSyncPass)
		if err != nil {
			return true, fmt.Errorf("authenticate to Koha: %v", err.Error())
		}
	}

	// Generate MARCXML record of RDF resource
	r, err = db.Query(fmt.Sprintf(queryRDF2MARC, cfg.RDFStore.DefaultGraph, uri, uri))
	if err != nil {
		return true, fmt.Errorf("db.Query: %v", err.Error())
	}

	err = json.Unmarshal(r, &res)
	if err != nil {
		return true, fmt.Errorf("parse SPARQL response: %v", err.Error())
	}

	rec, err := convertRDF2MARC(*res)
	if err != nil {
		return true, fmt.Errorf("convertRDF2MARC: %v", err.Error())
	}

	marc, err := xml.Marshal(rec)
	if err != nil {
		return true, fmt.Errorf("MARC xml marhsal: %v", err.Error())
	}

	// we're updating
	err = syncUpdatedManifestation(cfg.KohaPath, kohaCookies, marc, bibnr)
	if err != nil {
		return true, fmt.Errorf("syncUpdatedManifestation: %v", err.Error())
	}

	return false, nil
}

func syncDeleteResource(biblionr int) (bool, error) {
	err := syncDeletedManifestation(cfg.KohaPath, kohaCookies, biblionr)
	if err != nil {
		return true, err
	}
	return false, nil
}
