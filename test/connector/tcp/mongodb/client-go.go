package main

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	client, err := mongo.NewClient(
		options.Client().ApplyURI(
			"mongodb://user0:pass0@localhost:27018/?authSource=admin&ssl=false",
		).SetMaxPoolSize(1),
	)
	if err != nil {
		panic(err)
		return
	}

	err = client.Connect(context.Background())
	if err != nil {
		panic(err)
		return
	}


	collections, err := client.Database("meow").ListCollectionNames(context.Background(), bson.D{{}})
	if err != nil {
		panic(err)
		return
	}

	fmt.Println(collections)
}