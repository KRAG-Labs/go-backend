package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Data structure for MongoDB
type WSMessage struct {
	Content   string    `bson:"content"`
	Timestamp time.Time `bson:"timestamp"`
}

func main() {
	// 1. Connect to MongoDB (service name is 'mongo' from docker-compose)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://mongo:27017"))
	if err != nil {
		log.Fatal("Mongo Connection Error:", err)
	}

	collection := client.Database("krag_db").Collection("messages")
	fmt.Println("Connected to MongoDB successfully")

	// 2. WebSocket Handler
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("Upgrade error:", err)
			return
		}
		defer conn.Close()

		for {
			// Read message from browser/client
			_, msg, err := conn.ReadMessage()
			if err != nil {
				break
			}

			// 3. Save to MongoDB
			newEntry := WSMessage{
				Content:   string(msg),
				Timestamp: time.Now(),
			}
			_, err = collection.InsertOne(context.TODO(), newEntry)
			if err != nil {
				log.Println("Mongo Insert Error:", err)
				continue
			}

			// 4. Instant Feedback: Get the latest message count to show it worked
			count, _ := collection.CountDocuments(context.TODO(), bson.D{})
			response := fmt.Sprintf("Saved! Total messages in DB: %d", count)
			conn.WriteMessage(websocket.TextMessage, []byte(response))
		}
	})

	fmt.Println("Server starting on 0.0.0.0:8080")
	log.Fatal(http.ListenAndServe("0.0.0.0:8080", nil))
}
