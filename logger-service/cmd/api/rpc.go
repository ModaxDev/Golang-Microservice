package main

import (
	"context"
	"log-service/data"
	"time"
)

// RPCServer is a type that defines the methods we can call via RPC
type RPCServer struct {
}

// RPCPayload is a type that defines the payload we can send via RPC
type RPCPayload struct {
	Name string
	Data string
}

// LogInfo is a method that can be called via RPC
func (r *RPCServer) LogInfo(payload RPCPayload, resp *string) error {
	collection := client.Database("logs").Collection("logs")
	_, err := collection.InsertOne(context.TODO(), data.LogEntry{
		Name:      payload.Name,
		Data:      payload.Data,
		CreatedAt: time.Now(),
	})
	if err != nil {
		return err
	}

	*resp = "Processed payload via RPC" + payload.Name
	return nil
}
