package main

import (
	"golang.org/x/text/collate"
	"golang.org/x/text/language"
)

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
	ID              string           `json:"id"`
	EntityID        string           `xml:"entityID,attr" json:"entityID"`
	DisplayNames    []DisplayName    `xml:"IDPSSODescriptor>Extensions>UIInfo>DisplayName"`
	Descriptions    []Description    `xml:"IDPSSODescriptor>Extensions>UIInfo>Description" json:"Descriptions,omitempty"`
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

//Sort entries
type byDisplayName []EntityDescriptor

func (a byDisplayName) Len() int      { return len(a) }
func (a byDisplayName) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a byDisplayName) Less(i, j int) bool {
	//fmt.Printf("%s\n", a[j].DisplayNames[0].Value)
	cl := collate.New(language.French)
	return cl.CompareString(a[i].DisplayNames[0].Value, a[j].DisplayNames[0].Value) < 0
}
