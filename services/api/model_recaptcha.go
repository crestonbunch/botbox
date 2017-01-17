package api

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

const RecaptchaUrl = "https://www.google.com/recaptcha/api/siteverify"

type RecaptchaModel interface {
	// Verify that the token sent from the user is valid with the app secret.
	Verify(token string) (bool, error)
}

type Recaptcha struct {
	RecaptchaUrl    string
	RecaptchaSecret string
}

func (m *Recaptcha) Verify(token string) (bool, error) {
	vals := url.Values{}
	vals.Set("secret", m.RecaptchaSecret)
	vals.Set("response", token)

	log.Println(string(token))
	response, err := http.PostForm(m.RecaptchaUrl, vals)

	if err != nil {
		log.Println(err)
		return false, err
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Println(err)
		return false, err
	}
	log.Println(string(body))

	msg := struct {
		Success   bool   `json:'success'`
		Timestamp string `json:'challenge_ts'`
		Hostname  string `json:'hostname'`
	}{}

	err = json.Unmarshal(body, &msg)
	if err != nil {
		log.Println(err)
		return false, err
	}

	return msg.Success, nil
}
