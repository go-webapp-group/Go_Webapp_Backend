package main

import (
	"context"
	"log"
	"os"
	"webapp/db"
	"webapp/web"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	client, err := mongo.Connect(context.TODO(), clientOptions())
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.TODO())

	//client1, err := mongo.Connect(context.TODO(), clientOptions())
	mongoDB := db.NewMongo(client)

	// CORS is enabled only in prod profile
	cors := os.Getenv("profile") == "prod"

	app1 := web.NewApp(mongoDB, cors) //////
	//appcomment := web.NewCommentApp(mongoDB, cors)

	//err = appcomment.Serve()
	err = app1.Serve()

	log.Println("Error", err)
}

func clientOptions() *options.ClientOptions {
	host := "db"
	if os.Getenv("profile") != "prod" {
		host = "0.0.0.0"
	}
	return options.Client().ApplyURI(
		"mongodb://" + host + ":27017",
	)
}
