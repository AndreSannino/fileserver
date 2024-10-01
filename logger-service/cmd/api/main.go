package main

import (
	"context"
	"fmt"
	"log"
	"logger/data"
	"net"
	"net/http"
	"net/rpc"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	webPort  = "80"
	rpcPort  = "5001"
	mongoURL = "mongodb://mongo:27017"
	gRcpPort = "50001"
)

var client *mongo.Client

type Config struct {
	Models data.Models
}

func main() {
	// connect to mongo
	mongoClient, err := connectToMongo()
	if err != nil {
		log.Panic(err)
	}
	client = mongoClient

	//create a ctx to disconnect
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	//close connection
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	app := Config{
		Models: data.New(client),
	}

	// Register the RCP Server
	err = rpc.Register(new(RCPServer))
	go app.rcpListen()

	go app.gRPCListen()

	//start web server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	err = srv.ListenAndServe()
	if err != nil {
		log.Panic()
	}

	log.Println("Log Server start")
}

func (app *Config) rcpListen() error {
	log.Println("starting RCP server on port: " + rpcPort)
	listen, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", rpcPort))
	if err != nil {
		return err
	}
	defer listen.Close()

	for {
		rcpConn, err := listen.Accept()
		if err != nil {
			continue
		}
		go rpc.ServeConn(rcpConn)
	}
}

func connectToMongo() (*mongo.Client, error) {
	//connection options
	clientOpt := options.Client().ApplyURI(mongoURL)
	clientOpt.SetAuth(options.Credential{
		Username: "admin",
		Password: "password",
	})

	// connect
	c, err := mongo.Connect(context.TODO(), clientOpt)
	if err != nil {
		log.Println("Error connecting", err)
		return nil, err
	}

	err = c.Ping(context.Background(), nil)
	if err != nil {
		log.Println("Error connecting", err)
		return nil, err
	}

	log.Println("Connected to Mongo")

	return c, nil
}
