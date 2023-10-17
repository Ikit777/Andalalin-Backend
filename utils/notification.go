package utils

import (
	"log"

	"github.com/google/uuid"

	expo "github.com/oliveroneill/exponent-server-sdk-golang/sdk"
)

type Notification struct {
	IdUser uuid.UUID
	Title  string
	Body   string
	Token  string
}

func SendPushNotifications(data *Notification) {
	pushToken, err := expo.NewExponentPushToken(data.Token)
	if err != nil {
		panic(err)
	}

	// Create a new Expo SDK client
	client := expo.NewPushClient(nil)

	// Publish message
	response, err := client.Publish(
		&expo.PushMessage{
			To:       []expo.ExponentPushToken{pushToken},
			Body:     data.Body,
			Sound:    "default",
			Title:    data.Title,
			Priority: expo.DefaultPriority,
		},
	)

	// Check errors
	if err != nil {
		panic(err)
	}

	// Validate responses
	if response.ValidateResponse() != nil {
		log.Fatal(response.PushMessage.To, "failed")
	}
}
