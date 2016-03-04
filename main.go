package main

import (
	"encoding/json"
	"fmt"
	"os"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"
)

type Token struct {
	Value string `json:"token"`
	Webhook string `json:"webhook"`
	Channel string `json:"channel"`
}

type Config struct {
	Listen string `json:"listen"`
	Username string `json:"username"`
	IconUrl string `json:"icon_url"`
	Tokens []Token `json:"tokens"`
}

type Alert struct {
	CreatedAt        time.Time `json:"created_at"`
	ApplicationName  string    `json:"application_name"`
	AccountName      string    `json:"account_name"`
	Severity         string    `json:"severity"`
	Message          string    `json:"message"`
	ShortDescription string    `json:"short_description"`
	LongDescription  string    `json:"long_description"`
	Url              string    `json:"alert_url"`
}

type Deployment struct {
	CreatedAt       time.Time `json:"created_at"`
	ApplicationName string    `json:"application_name"`
	AccountName     string    `json:"account_name"`
	Changelog       string    `json:"changelog"`
	Revision        string    `json:"revision"`
	Url             string    `json:"deployment_url"`
	DeployedBy      string    `json:"deployed_by"`
}

type Post struct {
	Channel  string `json:"channel"`
	Text     string `json:"text"`
	Username string `json:"username"`
	IconUrl  string `json:"icon_url"`
}

func main() {
	config := Config{}
	var configFile string

	if len(os.Args) > 1 {
		configFile = os.Args[1]
	} else {
		configFile = "config.json"
	}

	configContent, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(configContent, &config)
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	for _, token := range(config.Tokens) {
		path := fmt.Sprintf("/webhook/%s", token.Value)
		mux.HandleFunc(path, func(c Config, t Token) (func (http.ResponseWriter, *http.Request)) {
			return func(w http.ResponseWriter, r *http.Request) {
				webhookHandler(w, r, c, t)
			}
		}(config, token))
	}

	log.Fatal(http.ListenAndServe(config.Listen, mux))
}

func webhookHandler(w http.ResponseWriter, r *http.Request, config Config, token Token) {
	if r.Method != "POST" {
		log.Printf("[%s] bad method: %s", r.RemoteAddr, r.Method)
		http.NotFound(w, r)
		return
	}

	if r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
		log.Printf("bad content type: %s", r.Header.Get("Content-Type"))
		http.Error(w, "406 Not Acceptable", http.StatusNotAcceptable)
		return
	}

	err := r.ParseForm()
	if err != nil {
		log.Printf("[%s] error parsing request body: %s", r.RemoteAddr, err)
		http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
		return
	}

	if alertPayload, ok := r.PostForm["alert"]; ok {
		alert := Alert{}
		err = json.Unmarshal([]byte(alertPayload[0]), &alert)
		if err != nil {
			log.Printf("[%s] error parsing JSON: %s", r.RemoteAddr, err)
			http.Error(w, "400 Bad Request", http.StatusBadRequest)
			return
		}
		post := Post{Username: config.Username, Channel: token.Channel, Text: alertText(alert), IconUrl: config.IconUrl}
		if ok = webhookSender(post, token.Webhook); !ok {
			log.Printf("[%s] error sending post: %+v", r.RemoteAddr, post)
			http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
			return
		}
		log.Printf("[%s] post sent: %+v", r.RemoteAddr, post)
	} else if deploymentPayload, ok := r.PostForm["deployment"]; ok {
		deployment := Deployment{}
		err := json.Unmarshal([]byte(deploymentPayload[0]), &deployment)
		if err != nil {
			log.Printf("[%s] error parsing JSON: %s", r.RemoteAddr, err)
			http.Error(w, "400 Bad Request", http.StatusBadRequest)
			return
		}
		post := Post{Username: config.Username, Channel: token.Channel, Text: deploymentText(deployment), IconUrl: config.IconUrl}
		if ok = webhookSender(post, token.Webhook); !ok {
			log.Printf("[%s] error sending post: %+v", r.RemoteAddr, post)
			http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
			return
		}
		log.Printf("[%s] post sent: %+v", r.RemoteAddr, post)
	} else {
		log.Printf("[%s] invalid message: %+v", r.RemoteAddr, r.PostForm)
		http.Error(w, "400 Bad Request", http.StatusBadRequest)
	}
}

func webhookSender(post Post, webhook string) (ok bool) {
	bytes, err := json.Marshal(post)
	if err != nil {
		log.Printf("error marshalling JSON: %s", err)
		return
	}
	payload := []string{string(bytes)}
	res, err := http.PostForm(webhook, url.Values{"payload": payload})
	if err != nil {
		log.Printf("error sending webhook: %s", err)
		return
	}
	ioutil.ReadAll(res.Body)

	return true
}

func alertText(alert Alert) string {
	return fmt.Sprintf("**[%s](%s): %s**\n*%s*", alert.ShortDescription, alert.Url, alert.Severity, alert.Message)
}

func deploymentText(deployment Deployment) string {
	return fmt.Sprintf("**[%s deployed](%s) revision %s**\n```\n%s\n```", deployment.ApplicationName, deployment.Url, deployment.Revision, deployment.Changelog)
}
