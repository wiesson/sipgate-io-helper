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

var token string = ""
var url string = ""

type Response struct {
	Answer string `xml:"onAnswer,attr"`
	Hangup string `xml:"onHangup,attr"`
	Dial   string `xml:"Dial>Number"`
}

/*

 */

type Tunnels struct {
	Tunnel  []Tunnel `json:"tunnels"`
}

type Tunnel struct {
	PublicUrl string `json:"public_url"`
}

type SipgateIoUrls struct {
	IncomingUrl string `json:"incomingUrl"`
	OutgoingUrl string `json:"outgoingUrl"`
}

func main() {
	var user string = os.Args[1]
	var pw string = os.Args[2]

	GetUrlFromNgrok()
	APIgetToken(user, pw)
	APIsetPushUrl()
	http.HandleFunc("/", foo)
	http.ListenAndServe(":3000", nil)
}

func APIgetToken(user string, pw string) {
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(map[string]string{"username": user, "password": pw})
	res, err := http.Post("https://api.sipgate.com/v1/authorization/token", "application/json; charset=utf-8", b)
	if err != nil {
		panic(err)
	}
	var body struct {
		Token string `json:"token"`
	}

	json.NewDecoder(res.Body).Decode(&body)
	token = body.Token
}

func APIsetPushUrl() {
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(map[string]string{"incomingUrl": url, "outgoingUrl": url})

	req, _ := http.NewRequest("PUT", "https://api.sipgate.com/v1/settings/sipgateio", b)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	client := &http.Client{}
	r, err := client.Do(req)

	if err != nil {
		panic(err)
	}
	defer r.Body.Close()
}

func GetUrlFromNgrok() {
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
			url = value.PublicUrl
			break
		}
	}

	fmt.Println("Found ngrok url", url)
}

func foo(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	log.Println(r.Form)

	response := Response{Answer: url, Hangup: url, Dial: "491751109743"}

	x, err := xml.MarshalIndent(response, "", "  ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/xml")
	w.Write(x)
}
