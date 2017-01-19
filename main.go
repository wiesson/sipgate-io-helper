package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

type Response struct {
	Answer string `xml:"onAnswer,attr"`
	Hangup string `xml:"onHangup,attr"`
	Dial   string `xml:"Dial>Number"`
}

type Tunnels struct {
	Tunnel []Tunnel `json:"tunnels"`
}

type Tunnel struct {
	PublicUrl string `json:"public_url"`
}

type SipgateIoUrls struct {
	IncomingUrl string `json:"incomingUrl"`
	OutgoingUrl string `json:"outgoingUrl"`
}

type API struct {
	user       string
	password   string
	Token      string
	PushApiUrl string
}

func (a *API) GetAccessToken() string {
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(map[string]string{"username": a.user, "password": a.password})
	res, err := http.Post("https://api.sipgate.com/v1/authorization/Token", "application/json; charset=utf-8", b)

	if err != nil {
		panic(err)
	}

	var body struct {
		Token string `json:"Token"`
	}

	json.NewDecoder(res.Body).Decode(&body)
	a.Token = body.Token

	fmt.Println("Got Token from sipgate api: ", a.Token)

	return a.Token
}

func (a *API) GetNgrokUrl() {
	tunnels := &Tunnels{}
	response, err := http.Get("http://127.0.0.1:4040/api/tunnels")
	defer response.Body.Close()

	if err != nil {
		panic(err)
	}

	err = json.NewDecoder(response.Body).Decode(tunnels)

	if err != nil {
		panic(err)
	}

	for _, value := range tunnels.Tunnel {
		if strings.HasPrefix(value.PublicUrl, "https") {
			a.PushApiUrl = value.PublicUrl
			break
		}
	}

	fmt.Println("Found ngrok url: ", a.PushApiUrl)
}

func (a *API) SetPushApiUrl() {
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(map[string]string{"incomingUrl": url, "outgoingUrl": url})

	req, _ := http.NewRequest("PUT", "https://api.sipgate.com/v1/settings/sipgateio", b)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", a.Token))
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	client := &http.Client{}
	r, err := client.Do(req)

	if err != nil {
		panic(err)
	}
	defer r.Body.Close()
}

func (a *API) pushApiResponseHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	log.Println(r.Form)

	response := Response{Answer: a.PushApiUrl, Hangup: a.PushApiUrl}

	x, err := xml.MarshalIndent(response, "", "  ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/xml")
	w.Write(x)
}

func main() {
	api := &API{user:os.Args[1], password:os.Args[2]}
	api.GetAccessToken()
	api.SetPushApiUrl()
	api.GetNgrokUrl()

	http.HandleFunc("/", api.pushApiResponseHandler)
	http.ListenAndServe(":3000", nil)
}
