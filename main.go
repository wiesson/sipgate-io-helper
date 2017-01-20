package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"strings"
	"flag"
	"net/url"
	"strconv"
	"os"
)

const openIdUrl string = "login/sipgate-apps/protocol/openid-connect/token"
const settingsPushApiUrl string = "v1/settings/sipgateio"

type Response struct {
	Answer string `xml:"onAnswer,attr"`
	Hangup string `xml:"onHangup,attr"`
	// Dial   string `xml:"Dial>Number"`
}

type Tunnels struct {
	Tunnel []Tunnel `json:"tunnels"`
}

type Tunnel struct {
	PublicUrl string `json:"public_url"`
}

type PushApiUrls struct {
	IncomingUrl string `json:"incomingUrl"`
	OutgoingUrl string `json:"outgoingUrl"`
}

type API struct {
	user        string
	password    string
	AccessToken string
	PushApiUrl  string
	Env         string
	ApiBaseUrl  string
}

func (a *API) setEnv() {
	if a.Env == "dev" {
		a.ApiBaseUrl = "https://api.dev.sipgate.com/"
		return
	}

	a.ApiBaseUrl = "https://api.sipgate.com/"
}

func (a *API) GetSipgateApiToken() string {
	payload := url.Values{}
	payload.Set("client_id", "sipgate-app-web")
	payload.Add("username", a.user)
	payload.Add("password", a.password)
	payload.Add("grant_type", "password")

	client := &http.Client{}
	req, _ := http.NewRequest("POST", a.ApiBaseUrl + openIdUrl , bytes.NewBufferString(payload.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(payload.Encode())))

	res, err := client.Do(req)

	if err != nil {
		panic(err)
	}

	if res.StatusCode != 200 {
		panic("Got wrong status code during token request")
	}

	var responseBody struct {
		AccessToken   string `json:"access_token"`
		refreshToken  string `json:"refresh_token"`
		identityToken string `json:"identity_token"`
	}

	json.NewDecoder(res.Body).Decode(&responseBody)
	a.AccessToken = responseBody.AccessToken

	fmt.Print("Got AccessToken from sipgate api: \n\n", a.AccessToken, "\n\n")

	return a.AccessToken
}

func (a *API) GetNgrokUrl() {
	tunnels := &Tunnels{}
	response, err := http.Get("http://127.0.0.1:4040/api/tunnels")

	if err != nil {
		fmt.Println("Can't find local ngrok")
		os.Exit(0)
	}

	defer response.Body.Close()

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
	json.NewEncoder(b).Encode(map[string]string{"incomingUrl": a.PushApiUrl, "outgoingUrl": a.PushApiUrl})

	req, _ := http.NewRequest("PUT", a.ApiBaseUrl + settingsPushApiUrl, b)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", a.AccessToken))
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

	pushApiResponse := &Response{Answer: a.PushApiUrl, Hangup: a.PushApiUrl}

	x, err := xml.MarshalIndent(pushApiResponse, "", "  ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/xml")
	w.Write(x)
}

func main() {
	email := flag.String("email", "username", "Your sipgate email address")
	password := flag.String("password", "password", "The password for your sipgate account")
	env := flag.String("env", "dev", "dev, live")
	token := flag.Bool("token", false, "Request an access token only")

	flag.Parse()

	if *email == "username" {
		fmt.Println("Please enter your username")
		os.Exit(0)
	}

	if *password == "password" {
		fmt.Println("Please enter your password")
		os.Exit(0)
	}

	api := &API{user:*email, password:*password, Env:*env}
	api.setEnv()
	api.GetSipgateApiToken()

	if *token {
		return
	}

	api.GetNgrokUrl()
	api.SetPushApiUrl()

	fmt.Print("\n\n")

	http.HandleFunc("/", api.pushApiResponseHandler)
	http.ListenAndServe(":3000", nil)
}
