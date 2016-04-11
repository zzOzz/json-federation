package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
)

var xmlData []byte

// Result sample
// [
// {
//  "entityID": "https://federation.unimes.fr/idp/shibboleth",
//  "DisplayNames": [
//   {
//   "value": "Université de Nimes",
//   "lang": "en"
//   },
//   {
//   "value": "Université de Nimes",
//   "lang": "fr"
//   }
//  ],
//  "Descriptions": [
//   {
//   "value": "All members of the UNIMES community: staff, students, library readers, alumni, staff from other institutions working locally, guests, etc.",
//   "lang": "en"
//   },
//   {
//   "value": "Tous les membres de la communauté UNIMES : personnels, étudiants, lecteurs des bibliothèques, anciens étudiants, personnels d'autres établissement saillant dans l'université, invités, prestataires, anciens personnels gardant une activité.",
//   "lang": "fr"
//   }
//  ],
//  "InformationURLs": [
//   {
//   "value": "http://www.unimes.fr",
//   "lang": "fr"
//   }
//  ],
//  "Logos": [
//   {
//   "value": "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAABAAAAAQCAIAAACQkWg2AAAACXBIWXMAAABIAAAASABGyWs+AAAACXZwQWcAAAAQAAAAEABcxq3DAAABLUlEQVQoz5WSv0rDUBjFT0socUgIXku9dhKHqMTNJR06iIMP0EFw8wlacCmuLoJvUMjS2b2DS2ddJIWQoZOYprQXQoKYW8U43BhaS9Cc6Qzf7/tf+kw+UERlAFav32l3Adze3Fm9fhDGnXbXcce5wC9pqixMEMYAyrOnXKBaq2VxGflVPV5mVoDZdLpeQTBZjJS5ZHieRKqMIbcugau3QWtDGfGldO+tSFPltML2FvljO5q+0pI/ZwXWWvgO/1LgFgR+ZpAAmM2G2WzAvz7deSQVBuBidyDMuiQAB/oegMTHkTICYEeGMLkzZG/zMD/xOCUV5nEKTbcjw44Mj9N77wzAZPKaVqC07rjjfeBQcQCwBQHAXtKUbEFM8gyolNYBlMR7B2G8yXPbEBIP8g2W5Ws8XJqA0gAAACV0RVh0ZGF0ZTpjcmVhdGUAMjAxNS0wNy0yMFQxMDo1MDowNCswMjowMB3sEi8AAAAldEVYdGRhdGU6bW9kaWZ5ADIwMTUtMDctMjBUMTA6NTA6MDQrMDI6MDBssaqTAAAAAElFTkSuQmCC",
//   "height": "16",
//   "width": "16"
//   }
//  ]
// }

// EntitiesDoc is Root element of idps array
type EntitiesDoc struct {
	// XMLName          xml.Name           `xml:"EntitiesDescriptor"`
	EntityDescriptors []EntityDescriptor `xml:"EntityDescriptor"`
}

// EntityDescriptor is idp descriptor
type EntityDescriptor struct {
	// XMLName xml.Name `xml:"EntityDescriptor"`
	EntityID        string           `xml:"entityID,attr" json:"entityID"`
	DisplayNames    []DisplayName    `xml:"IDPSSODescriptor>Extensions>UIInfo>DisplayName"`
	Descriptions    []Description    `xml:"IDPSSODescriptor>Extensions>UIInfo>Description" json:"Description,omitempty"`
	InformationURLs []InformationURL `xml:"IDPSSODescriptor>Extensions>UIInfo>InformationURL"`
	Logos           []Logo           `xml:"IDPSSODescriptor>Extensions>UIInfo>Logo" json:"Logos,omitempty"`
	Scopes          []Scope          `xml:"IDPSSODescriptor>Extensions>Scope" json:"Scopes,omitempty"`
}

// DisplayName is idp info (per lang)
type DisplayName struct {
	// Scope string   `xml:"Scope"`
	Value string `xml:",chardata" json:"value"`
	Lang  string `xml:"lang,attr" json:"lang"`
}

// Description is idp info (fr,en)
type Description struct {
	// Scope string   `xml:"Scope"`
	Value string `xml:",chardata" json:"value"`
	Lang  string `xml:"lang,attr" json:"lang"`
}

// InformationURL is idp url info (fr,en)
type InformationURL struct {
	// Scope string   `xml:"Scope"`
	Value string `xml:",chardata" json:"value"`
	Lang  string `xml:"lang,attr" json:"lang"`
}

// Logo is base64 logo or logo url whith size info (height,width)
type Logo struct {
	// Scope string   `xml:"Scope"`
	Value  string `xml:",chardata" json:"value"`
	Height string `xml:"height,attr" json:"height"`
	Width  string `xml:"width,attr" json:"width"`
}

// Scope idp scope (email domain(s))
type Scope struct {
	Value string `xml:",chardata" json:"value"`
}

func main() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", Index)
	router.HandleFunc("/Reset", ResetCache)
	log.Fatal(http.ListenAndServe(":8080", router))
}

//Index is DiscoFeed with scopes
func Index(w http.ResponseWriter, r *http.Request) {

	if xmlData == nil {
		xmlData = []byte(loadXML("https://federation.renater.fr/renater/idps-renater-metadata.xml"))
	}
	data := &EntitiesDoc{}
	err := xml.Unmarshal(xmlData, data)
	if nil != err {
		fmt.Println("Error unmarshalling from XML", err)
		return
	}

	var etablissements []tEtablissement
	var found bool
	for _, entitydesc := range data.EntityDescriptors {
		etablissement := new(tEtablissement)
		for _, displayName := range entitydesc.DisplayNames {
			etablissement.ID = entitydesc.EntityID
			if displayName.Lang == "fr" {
				etablissement.Text = displayName.Value
			}
		}
		// fmt.Println("index:", entitydesc)
		if len(entitydesc.Logos) != 0 {
			etablissement.Logo = entitydesc.Logos[0].Value
		}
		for _, scope := range entitydesc.Scopes {
			etablissement.Scope = append(etablissement.Scope, scope.Value)
			if strings.Contains(scope.Value, r.URL.Query().Get("term")) {
				found = true
			} else {
				found = false
			}
		}
		if found {
			etablissements = append(etablissements, *etablissement)
		}
	}

	if r.URL.Query().Get("term") != "" {
		fmt.Printf("Term: %s\n", r.URL.Query().Get("term"))
		result, err := json.Marshal(etablissements)
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
		//fmt.Printf("%s\n", string(contents))
		return string(contents)
	}
	return ""
}

// type tEtablissements struct {
// 	etablissement []tEtablissement
// }

type tEtablissement struct {
	ID    string   `json:"id"`
	Logo  string   `json:"logo"`
	Text  string   `json:"text"`
	Scope []string `json:"scope"`
}

// [
// {"id":"urn:mace:cru.fr:federation:univ-lyon1.fr",
// "logo":"data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAABAAAAAQCAIAAACQkWg2AAAACXBIWXMAAABIAAAASABGyWs+AAAACXZwQWcAAAAQAAAAEABcxq3DAAAA7ElEQVQoz41SsW3DMBA8GS5+CBrgGgbi5jWFCw9CJKUFDxIgnoLfmEDWECDvkO+YghRtUSzEisDx7u/+2P39MA6MjWeSPQ5MH58b32vAro08RcOAp6yRfUPm3mMUAPoAuVihu7U2RoFlchGn61quJuj4C4DOHkDKpmHQMNSW9NbRxcPwa47h4g2AzvxMIBffI+p3X+7kot772pLeuoWz0zWrJGiUsvpV6ElgFz2Si7BcYsxrtVxs0MWn6Ivh82+YM5x90iB7hGEyrI+vRM48sySUJWbVMBTtqrtG03mO9U2oTXgV0iBMogFbzyT/VNdfIJ4V1XcAAAAldEVYdGRhdGU6Y3JlYXRlADIwMTUtMDctMjBUMTA6NTA6MDMrMDI6MDDYSyyhAAAAJXRFWHRkYXRlOm1vZGlmeQAyMDE1LTA3LTIwVDEwOjUwOjAzKzAyOjAwqRaUHQAAAABJRU5ErkJggg==",
// "text":"Université de Lyon 1 - Claude Bernard",
// "scope":["univ-lyon1.fr"]}
// ]
