package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/gorilla/mux"
)

var xmlData []byte

func main() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", Index)
	router.HandleFunc("/Reset", ResetCache)
	// log.Fatal(http.ListenAndServe("0.0.0.0:8080", router))
	log.Fatal(http.ListenAndServe("127.0.0.1:8081", router))
}

//Index is DiscoFeed with scopes
func Index(w http.ResponseWriter, r *http.Request) {

	if xmlData == nil {
		xmlData = []byte(loadXML("https://metadata.federation.renater.fr/renater/main/main-idps-renater-metadata.xml"))
	}
	data := &EntitiesDoc{}
	err := xml.Unmarshal(xmlData, data)
	if nil != err {
		fmt.Println("Error unmarshalling from XML", err)
		return
	}

	var dataFilters []EntityDescriptor
	var found bool
	var terms []string
	if r.URL.Query().Get("term") != "" {
		terms = strings.Split(r.URL.Query().Get("term"), "@")
	}
	var term string
	if len(terms) == 1 {
		term = terms[0]
	}
	if len(terms) == 2 {
		term = terms[1]
	}
	for i, entitydesc := range data.EntityDescriptors {
		data.EntityDescriptors[i].ID = entitydesc.EntityID
		dataFilter := new(EntityDescriptor)
		*dataFilter = entitydesc
		dataFilter.ID = dataFilter.EntityID
		for _, scope := range entitydesc.Scopes {
			if strings.Contains(scope.Value, term) {
				found = true
			} else {
				found = false
			}
		}
		if found {
			dataFilters = append(dataFilters, *dataFilter)
		}
	}
	if len(dataFilters) > 0 {
		sort.Sort(byDisplayName(dataFilters))
		fmt.Printf("first:%s\n", dataFilters[0].DisplayNames[0].Value)
		fmt.Printf("last:%s\n", dataFilters[len(dataFilters)-1].DisplayNames[0].Value)
	}
	sort.Sort(byDisplayName(data.EntityDescriptors))
	if r.URL.Query().Get("term") != "" {
		fmt.Printf("Search Term: %s\n", r.URL.Query().Get("term"))
		result, err := json.Marshal(dataFilters)
		if nil != err {
			fmt.Println("Error marshalling to JSON", err)
			return
		}
		fmt.Fprintf(w, "%s\n", result)
	} else {
		result, err := json.Marshal(data.EntityDescriptors)
		if nil != err {
			fmt.Println("Error marshalling to JSON", err)
			return
		}
		fmt.Fprintf(w, "%s\n", result)
	}

}

//ResetCache reload XML
func ResetCache(w http.ResponseWriter, r *http.Request) {
	xmlData = nil
	fmt.Fprintf(w, "%s\n", "{\"ok\"}")
}

func loadXML(url string) string {

	response, err := http.Get(url)
	if err != nil {
		fmt.Printf("%s", err)
		os.Exit(1)
	} else {
		defer response.Body.Close()
		contents, err := ioutil.ReadAll(response.Body)
		if err != nil {
			fmt.Printf("%s", err)
			os.Exit(1)
		}
		return string(contents)
	}
	return ""
}
