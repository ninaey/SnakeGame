package main

import (
	"fmt"
	"net/http"
	"os"

	"SnakeGame/handlers"
)

func main() {
	fs := http.FileServer(http.Dir("../frontend"))
	http.Handle("/", fs)
	http.HandleFunc("/api/player", handlers.GetPlayerHandler)
	http.HandleFunc("/api/earn", handlers.EarnCoinsHandler)
	http.HandleFunc("/api/equip", handlers.EquipHandler)
	http.HandleFunc("/api/cart", handlers.CartHandler)
	http.HandleFunc("/api/cart/remove", handlers.RemoveCartItemHandler)
	http.HandleFunc("/api/checkout", handlers.CheckoutHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Printf("Snake server running at http://localhost:%s\n", port)
	http.ListenAndServe(":"+port, nil)
}
