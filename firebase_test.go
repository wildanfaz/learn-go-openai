package main

import (
	"context"
	"testing"

	firebase "firebase.google.com/go"

	"google.golang.org/api/option"
)

func TestFirebase(t *testing.T) {
	opt := option.WithCredentialsFile("warkop-2a742-firebase-adminsdk-cyn4d-1f6b0fdbd5.json")
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
	  t.Error(err)
	}

	client, err := app.Firestore(context.Background())
	if err != nil {
	  t.Error(err)
	}

	ref := client.Collection("users").Doc("muhamadwildanfaz@gmail.com")

	_, err = ref.Set(context.Background(), map[string]interface{}{"displayName": "Muhamad Wildan Faz", "age": 20})
	if err != nil {
	  t.Error(err)
	}

	t.Log("Done")
}