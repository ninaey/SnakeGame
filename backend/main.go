package main

import (
	"fmt"
	"net/http"
	"os"

	"SnakeGame/handlers"
)

func main() {
	fs := http.FileServer(http.Dir("../frontend")) // Serve frontend
	http.Handle("/", fs)
	http.HandleFunc("/api/player", handlers.GetPlayerHandler)
	http.HandleFunc("/api/earn", handlers.EarnCoinsHandler)
	http.HandleFunc("/api/equip", handlers.EquipHandler)

	// Cart API
	http.HandleFunc("POST /api/user/cart/items", handlers.PostCartItemsHandler)
	http.HandleFunc("GET /api/user/cart", handlers.GetCartHandler)
	http.HandleFunc("/api/user/cart/items/{id}", handlers.CartItemsIDHandler)
	http.HandleFunc("POST /api/user/orders", handlers.CheckoutHandler)

	// Legacy routes
	http.HandleFunc("GET /api/cart", handlers.GetCartHandler)
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
