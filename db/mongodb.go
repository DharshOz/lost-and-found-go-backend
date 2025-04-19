package db

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	Client *mongo.Client
	dbName = "test" // Your database name
)

func ConnectMongoDB(uri string) {
	serverAPIOptions := options.ServerAPI(options.ServerAPIVersion1)
	clientOptions := options.Client().
		ApplyURI(uri).
		SetServerAPIOptions(serverAPIOptions).
		SetConnectTimeout(10 * time.Second).
		SetServerSelectionTimeout(30 * time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var err error
	Client, err = mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal("❌ MongoDB Connection Error:", err)
	}

	// Send a ping to confirm a successful connection
	err = Client.Ping(ctx, nil)
	if err != nil {
		log.Fatal("❌ MongoDB Ping Failed:", err)
	}

	fmt.Println("✅ Successfully connected to MongoDB!")
}

// DisconnectMongoDB gracefully disconnects from MongoDB
func DisconnectMongoDB() {
	if Client == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := Client.Disconnect(ctx)
	if err != nil {
		log.Printf("⚠️ MongoDB Disconnection Error: %v", err)
		return
	}

	fmt.Println("✅ Successfully disconnected from MongoDB")
}

// GetCollection returns a MongoDB collection instance
func GetCollection(name string) *mongo.Collection {
	if Client == nil {
		log.Fatal("❌ MongoDB client not initialized")
	}
	return Client.Database(dbName).Collection(name)
}
