package main

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"strconv"
)

var svcNotAuthorized = errors.New("unauthorized")
var ErrNotManifestation = errors.New("only manifestations are synced to Koha")

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

// svcAuth attemps to authenticate against Koha's /svc endpoint.
// If sucessfull, it returns a cookiejar with the authenticated cookie.
func svcAuth(kohaPath, user, pass string) (http.CookieJar, error) {
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

// svcNew takes the given marcxml record and send it with a  POST request
// to /svc/new_bib. It returns the biblionumber it got assigned fro Koha.
func svcNew(kohaPath string, jar http.CookieJar, marc []byte) (int, error) {
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

// svcUpdate takes the given marcxml record and send it with a POST request
// to the /svc/{biblio} endpoint.
func svcUpdate(kohaPath string, jar http.CookieJar, marc []byte, biblio int) error {
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

// svcDelete sends a DELETE request to /svc/biblio/{biblionr}
// TODO this depends on patch #12590
// http://bugs.koha-community.org/bugzilla3/show_bug.cgi?id=12590
func svcDelete(kohaPath string, jar http.CookieJar, biblio int) error {
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
	var profile string

	q, err := qBank.Prepare("resource",
		struct{ Graph, Res string }{cfg.RDFStore.DefaultGraph, uri})
	if err != nil {
		return 0, true, fmt.Errorf("error preparing query: %v", err.Error())
	}

	res, err := db.Query(q)
	if err != nil {
		return 0, true, fmt.Errorf("db.Query: %v", err.Error())
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
		return 0, false, ErrNotManifestation
	}

	// Make sure we are authenticated to Koha
	if kohaCookies == nil {
		kohaCookies, err = svcAuth(cfg.KohaPath, cfg.KohaSyncUser, cfg.KohaSyncPass)
		if err != nil {
			return 0, true, fmt.Errorf("authenticate to Koha: %v", err.Error())
		}
	}

	// Generate MARCXML record of RDF resource
	q, err = qBank.Prepare("rdf2marc",
		struct{ Graph, Res string }{cfg.RDFStore.DefaultGraph, uri})
	if err != nil {
		return 0, true, fmt.Errorf("error preparing query: %v", err.Error())
	}

	res, err = db.Query(q)
	if err != nil {
		return 0, true, fmt.Errorf("db.Query: %v", err.Error())
	}

	rec, err := convertRDF2MARC(res)
	if err != nil {
		return 0, true, fmt.Errorf("convertRDF2MARC: %v", err.Error())
	}

	marc, err := xml.Marshal(rec)
	if err != nil {
		return 0, true, fmt.Errorf("MARC xml marhsal: %v", err.Error())
	}

	bibnr, err := svcNew(cfg.KohaPath, kohaCookies, marc)
	if err != nil {
		return 0, true, fmt.Errorf("svcNew: %v", err.Error())
	}

	// store the koha id as property on the RDF resource
	q, err = qBank.Prepare("insertKohaID",
		struct {
			Graph  string
			Res    string
			KohaID int
		}{
			cfg.RDFStore.DefaultGraph,
			uri,
			bibnr,
		})
	body, err := db.Proxy(q)
	if err != nil {
		return 0, true, fmt.Errorf("error preparing query: %v", err.Error())
	}
	defer body.Close()
	r, err := ioutil.ReadAll(body)
	if err != nil {
		return 0, true, fmt.Errorf("db.Query: %v", err.Error())
	}

	// TODO check if we can use db.Query, is the message bound to some variable?
	if bytes.Index(r, []byte("1 (or less) triples")) == -1 {
		// TODO now what? retry will create another duplicate biblio record in Koha..
		if err != nil {
			return 0, false, fmt.Errorf("insert koha ID on resource: %v", err.Error())
		}
	}

	return bibnr, false, nil
}

func syncUpdateResource(uri string, biblionr int) (bool, error) {
	var profile string
	q, err := qBank.Prepare("resource",
		struct{ Graph, Res string }{cfg.RDFStore.DefaultGraph, uri})
	if err != nil {
		return true, fmt.Errorf("error preparing query: %v", err.Error())
	}
	res, err := db.Query(q)
	if err != nil {
		return true, fmt.Errorf("db.Query: %v", err.Error())
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

	if profile != "manifestation" {
		return false, ErrNotManifestation
	}

	if bibnrStr != "" {
		bibnr, err = strconv.Atoi(bibnrStr)
		if err != nil {
			return false, errors.New("kohaID on resource is not an integer")
		}
	}

	if bibnr == 0 {
		return false, errors.New("cannot sync to Koha; missing kohaID on resource")
		// TODO set task to "create", and retry?
	}

	// Make sure we are authenticated to Koha
	if kohaCookies == nil {
		kohaCookies, err = svcAuth(cfg.KohaPath, cfg.KohaSyncUser, cfg.KohaSyncPass)
		if err != nil {
			return true, fmt.Errorf("authenticate to Koha: %v", err.Error())
		}
	}

	// Generate MARCXML record of RDF resource
	q, err = qBank.Prepare("rdf2marc",
		struct{ Graph, Res string }{cfg.RDFStore.DefaultGraph, uri})
	if err != nil {
		return true, fmt.Errorf("error preparing query: %v", err.Error())
	}

	res, err = db.Query(q)
	if err != nil {
		return true, fmt.Errorf("db.Query: %v", err.Error())
	}

	rec, err := convertRDF2MARC(res)
	if err != nil {
		return true, fmt.Errorf("convertRDF2MARC: %v", err.Error())
	}

	marc, err := xml.Marshal(rec)
	if err != nil {
		return true, fmt.Errorf("MARC xml marhsal: %v", err.Error())
	}

	// we're updating
	err = svcUpdate(cfg.KohaPath, kohaCookies, marc, bibnr)
	if err != nil {
		return true, fmt.Errorf("svcUpdate: %v", err.Error())
	}

	return false, nil
}

func syncDeleteResource(biblionr int) (bool, error) {
	err := svcDelete(cfg.KohaPath, kohaCookies, biblionr)
	if err != nil {
		return true, err
	}
	return false, nil
}
