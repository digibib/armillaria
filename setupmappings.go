// +build ignore

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"strings"
)

func urlify(s string) string { return fmt.Sprintf("<%s>", s) }

func main() {
	cb, err := ioutil.ReadFile("data/mappings/_common")
	if err != nil {
		log.Fatal(err)
	}
	var common preMappings
	err = json.Unmarshal(cb, &common)
	if err != nil {
		log.Fatal(err)
	}

	files, err := filepath.Glob("data/mappings/*.json")
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {
		properties := make(map[string]interface{})
		var m preMappings
		for _, v := range common {
			for ki, vi := range v {
				properties[ki] = vi
			}
		}

		b, err := ioutil.ReadFile(f)
		if err != nil {
			log.Fatal(err)
		}
		err = json.Unmarshal(b, &m)
		if err != nil {
			log.Fatal(err)
		}
		for _, v := range m {
			for ki, vi := range v {
				properties[ki] = vi
			}
		}
		var out []byte

		profile := strings.TrimSuffix(strings.TrimPrefix(f, "data/mappings/"), ".json")
		out, err = json.Marshal(properties)
		var export bytes.Buffer
		export.WriteString("{\"")
		export.WriteString(profile)
		export.WriteString("\":{\"_id\": {\"path\": \"uri\"},\"properties\":")
		export.Write(out)
		export.WriteString("}}")
		//fmt.Println(string(export.Bytes()))

		req, err := http.NewRequest(
			"PUT",
			"http://localhost:9200/public/"+profile+"/_mapping",
			bytes.NewReader(export.Bytes()))
		if err != nil {
			log.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()
		b, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		if resp.StatusCode > 300 || resp.StatusCode < 200 {
			fmt.Printf("ERROR setting mappings for /public/%s\n", profile)
			fmt.Println(string(b))
		} else {
			fmt.Printf("OK set mappings for /public/%s\n", profile)
		}
	}

}
