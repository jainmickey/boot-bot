package slacknotifier

import (
	"fmt"
	"net/http"
	"strings"
)

type Slack struct {
	webHook string
	name    string
	face    string
}

func New(webHook string) Slack {
	return Slack{
		webHook: webHook,
		name:    "Notifier",
		face:    "robot_face"}
}

func (slack *Slack) Name(name string) {
	slack.name = name
}

func (slack *Slack) Face(face string) {
	slack.face = face
}

func (slack Slack) Notify(text string) error {
	json := fmt.Sprintf(`{"text":"%s"}`, text)

	req, err := http.NewRequest("POST", slack.webHook, strings.NewReader(json))
	if err != nil {
		return fmt.Errorf("Can't connect to host %s: %s", slack.webHook, err.Error())
	}

	req.Header.Set("Content-Type", "application/json")

	client := http.Client{}
	_, err = client.Do(req)

	return err
}
