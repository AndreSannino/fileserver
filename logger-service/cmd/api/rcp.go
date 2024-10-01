package main

import (
	"context"
	"log"
	"logger/data"
	"time"
)

type RCPServer struct{}

type RCPPayload struct {
	Name string
	Data string
}

func (r *RCPServer) LogInfo(payload RCPPayload, resp *string) error {
	collection := client.Database("logs").Collection("logs")
	_, err := collection.InsertOne(context.TODO(), data.LogEntry{
		Name:      payload.Name,
		Data:      payload.Data,
		CreatedAt: time.Now(),
	})

	if err != nil {
		log.Println(err)
		return err
	}

	*resp = "Process payload via RCP: " + payload.Name
	return nil
}
