package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

var (
	ctx      = context.Background()
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true }, // Allow all for testing
	}
)

func main() {
	// Connect to Redis (Notice the hostname is 'redis' matching docker-compose)
	rdb := redis.NewClient(&redis.Options{
		Addr: "redis:6379",
	})

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}
		defer conn.Close()

		for {
			// Read message
			_, msg, err := conn.ReadMessage()
			if err != nil {
				break
			}

			// 1. Save to Redis
			err = rdb.Set(ctx, "last_msg", msg, 0).Err()
			if err != nil {
				log.Println("Redis Error:", err)
			}

			// 2. Echo back
			response := fmt.Sprintf("Saved to Redis: %s", string(msg))
			conn.WriteMessage(websocket.TextMessage, []byte(response))
		}
	})

	port := ":8080"
	fmt.Println("Server starting on port", port)
	log.Fatal(http.ListenAndServe("0.0.0.0"+port, nil))
}