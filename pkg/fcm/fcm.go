package fcm

import (
	"context"
	"log"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

var client *messaging.Client

func Init(credentialsJSON string) error {
	opt := option.WithCredentialsJSON([]byte(credentialsJSON))
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		return err
	}
	client, err = app.Messaging(context.Background())
	if err != nil {
		return err
	}
	log.Println("FCM initialized")
	return nil
}

func Send(ctx context.Context, token *string, title, body string, data map[string]string) {
	if client == nil || *token == "" {
		log.Println("[FCM] Skipping send: client nil or token missing")
		return
	}
	_, err := client.Send(ctx, &messaging.Message{
		Token: *token,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data: data,
		Android: &messaging.AndroidConfig{
			Priority: "high",
		},
		APNS: &messaging.APNSConfig{
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{Sound: "default"},
			},
		},
	})
	if err != nil {
		log.Printf("FCM send error: %v", err)
	}
}
