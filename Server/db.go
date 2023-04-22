package main

import (
    "context"
    "fmt"
    "log"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

type ImageData struct {
    Name string
    Type string
    Data []byte
}

func main() {
    // Set client options
    clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")

    // Connect to MongoDB
    client, err := mongo.Connect(context.Background(), clientOptions)
    if err != nil {
        log.Fatal(err)
    }

    // Check the connection
    err = client.Ping(context.Background(), nil)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("Connected to MongoDB!")

    // Get handle for database
    database := client.Database("myDatabase")

    // Get handle for collection
    collection := database.Collection("myCollection")

    // Insert data
    imageData := ImageData{
        Name: "myImage",
        Type: "png",
        Data: []byte{},
    }

    insertResult, err := collection.InsertOne(context.Background(), imageData)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("Inserted data with ID:", insertResult.InsertedID)
}
