package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type Payload struct {
	Key               string `json:"key"`
	Txn               string `json:"txn"`
	Language          string `json:"language"`
	Country           string `json:"country"`
	Limit             int    `json:"limit"`
	Address           string `json:"address"`
	GeographicAddress bool   `json:"geographicAddress"`
	AddressID         string `json:"addressId"`
}

type FetchResult struct {
	Result struct {
		Code int    `json:"code"`
		Text string `json:"text"`
	} `json:"result"`
	EcadID       string `json:"ecadId"`
	EcadIDStatus struct {
		Code int    `json:"code"`
		Text string `json:"text"`
	} `json:"ecadIdStatus"`
	AddressTypeID int `json:"addressTypeId"`
	EircodeInfo   struct {
		EcadID  string `json:"ecadId"`
		Eircode string `json:"eircode"`
	} `json:"eircodeInfo"`
	PostalAddress struct {
		English []string `json:"english"`
		Irish   []string `json:"irish"`
	} `json:"postalAddress"`
	GeographicAddress struct {
		English []string `json:"english"`
		Irish   []string `json:"irish"`
	} `json:"geographicAddress"`
	AdministrativeInfo struct {
		EcadID      string `json:"ecadId"`
		Release     string `json:"release"`
		LaID        string `json:"laId"`
		DedID       string `json:"dedId"`
		SmallAreaID string `json:"smallAreaId"`
		Gaeltacht   bool   `json:"gaeltacht"`
	} `json:"administrativeInfo"`
	BuildingInfo struct {
		EcadID            string `json:"ecadId"`
		BuildingTypeID    int    `json:"buildingTypeId"`
		HolidayHome       bool   `json:"holidayHome"`
		UnderConstruction bool   `json:"underConstruction"`
		BuildingUse       string `json:"buildingUse"`
		Vacant            bool   `json:"vacant"`
	} `json:"buildingInfo"`
	SpatialInfo struct {
		EcadID string `json:"ecadId"`
		Ing    struct {
			Location struct {
				Easting  float64 `json:"easting"`
				Northing float64 `json:"northing"`
			} `json:"location"`
		} `json:"ing"`
		Itm struct {
			Location struct {
				Easting  float64 `json:"easting"`
				Northing float64 `json:"northing"`
			} `json:"location"`
		} `json:"itm"`
		Etrs89 struct {
			Location struct {
				Longitude float64 `json:"longitude"`
				Latitude  float64 `json:"latitude"`
			} `json:"location"`
		} `json:"etrs89"`
		SpatialAccuracy string `json:"spatialAccuracy"`
	} `json:"spatialInfo"`
	RelatedEcadIds struct {
		BuildingEcadID          string   `json:"buildingEcadId"`
		ThoroughfareEcadIds     []string `json:"thoroughfareEcadIds"`
		LocalityEcadIds         []string `json:"localityEcadIds"`
		PostTownEcadIds         []string `json:"postTownEcadIds"`
		PostCountyEcadIds       []string `json:"postCountyEcadIds"`
		GeographicCountyEcadIds []string `json:"geographicCountyEcadIds"`
	} `json:"relatedEcadIds"`
	DateInfo struct {
		Created  string `json:"created"`
		Modified string `json:"modified"`
	} `json:"dateInfo"`
	Input struct {
		Key     string `json:"key"`
		Txn     string `json:"txn"`
		EcadID  string `json:"ecadId"`
		History bool   `json:"history"`
	} `json:"input"`
	Links []struct {
		Rel  string `json:"rel"`
		Href string `json:"href"`
	} `json:"links"`
}

type SearchResult struct {
	TotalOptions int `json:"totalOptions"`
	Options      []struct {
		DisplayName string `json:"displayName"`
		AddressID   string `json:"addressId"`
		AddressType struct {
			Code int    `json:"code"`
			Text string `json:"text"`
		} `json:"addressType"`
		Links []struct {
			Rel  string `json:"rel"`
			Href string `json:"href"`
		} `json:"links"`
	} `json:"options"`
	Input struct {
		Key               string `json:"key"`
		Txn               string `json:"txn"`
		Language          string `json:"language"`
		Country           string `json:"country"`
		Limit             int    `json:"limit"`
		Address           string `json:"address"`
		GeographicAddress bool   `json:"geographicAddress"`
	} `json:"input"`
	Links []struct {
		Rel  string `json:"rel"`
		Href string `json:"href"`
	} `json:"links"`
}

func search_address(add string, res chan FetchResult, err chan string, k string) {
	link := "https://api-finder.eircode.ie/Latest/finderautocomplete?key=" + k + "&address=" + add + "&language=en&geographicAddress=true"
	r, e := http.Get(link)
	if e != nil {
		err <- e.Error()
		return
	}

	defer r.Body.Close()
	body, _ := io.ReadAll(r.Body)

	json_res := SearchResult{}
	json.Unmarshal(body, &json_res)
	if json_res.TotalOptions == 0 {
		err <- "No results found"
		return
	}

	p := Payload{
		Key:               json_res.Input.Key,
		Txn:               json_res.Input.Txn,
		Language:          json_res.Input.Language,
		Country:           json_res.Input.Country,
		Limit:             json_res.Input.Limit,
		GeographicAddress: json_res.Input.GeographicAddress,
		AddressID:         json_res.Options[0].AddressID,
		Address:           add,
	}
	fetch_postal_code(p, res, err)

}

func fetch_postal_code(payload Payload, res chan FetchResult, err chan string) {

	link := "https://api-finder.eircode.ie/Latest/findergetecaddata?key=" + payload.Key + "&addressId=" + payload.AddressID + "=&txn=" + payload.Txn + "&history=false&geographicAddress=true&clientVersion=e98fe302"
	r, e := http.Get(link)
	if e != nil {
		err <- e.Error()
		return
	}

	defer r.Body.Close()
	body, e := io.ReadAll(r.Body)

	if e != nil {
		err <- e.Error()
		return
	}

	json_res := FetchResult{}
	json.Unmarshal(body, &json_res)

	res <- json_res
}

const (
	API_KEY = "_32e333cd-9e4f-46e6-93cf-a78872b69138"
)

func main() {

	// get args
	args := os.Args

	// check args

	if len(args) < 2 {
		fmt.Println("Usage: ./main <address>")
		return
	}

	var key string = API_KEY

	if os.Getenv("API_KEY") != "" {
		key = os.Getenv("API_KEY")
	}

	address := strings.Join(args[1:], "%20")

	r := make(chan FetchResult, 1)
	e := make(chan string, 1)
	go search_address(address, r, e, key)

	fmt.Println("Looking for: ", strings.Join(args[1:], " "))

	select {
	case res := <-r:
		fmt.Println("EIRCODE: ", res.EircodeInfo.Eircode)

	case err := <-e:
		fmt.Println("Error: ", err)
	}

}
