package main

import (
	"fmt"
	"net/http"
	"os"

	"SnakeGame/handlers"
)

func main() {
	fs := http.FileServer(http.Dir("../frontend"))            // Serve frontend: when running from backend/, frontend is at ../frontend
	http.Handle("/", fs)                                      // serve the frontend
	http.HandleFunc("/api/player", handlers.GetPlayerHandler) // get player information
	http.HandleFunc("/api/earn", handlers.EarnCoinsHandler)   // earn coins
	http.HandleFunc("/api/equip", handlers.EquipHandler)      // equip a skin
	// Requirement: cart API
	http.HandleFunc("POST /api/user/cart/items", handlers.PostCartItemsHandler) // add an item to the cart
	http.HandleFunc("GET /api/user/cart", handlers.GetCartHandler)              // get the cart
	http.HandleFunc("/api/user/cart/items/{id}", handlers.CartItemsIDHandler)   // update an item (e.g. change quantity)
	http.HandleFunc("POST /api/user/orders", handlers.CheckoutHandler)          // process the cart: only charges for items the player does not already own (skins already in OwnedSkins are skipped). Prevents deducting coins for duplicate skins. Uses Idempotency-Key header: repeated requests with the same key within 24 hours receive the cached response without re-processing. Set header X-Simulate-Payment-Timeout: true to simulate gateway timeout (for testing retry).
	// Legacy routes (backward compatible)
	http.HandleFunc("GET /api/cart", handlers.GetCartHandler)           // get the cart
	http.HandleFunc("/api/cart", handlers.CartHandler)                  // add an item to the cart
	http.HandleFunc("/api/cart/remove", handlers.RemoveCartItemHandler) // remove an item from the cart
	http.HandleFunc("/api/checkout", handlers.CheckoutHandler)          // process the cart: only charges for items the player does not already own (skins already in OwnedSkins are skipped). Prevents deducting coins for duplicate skins. Uses Idempotency-Key header: repeated requests with the same key within 24 hours receive the cached response without re-processing. Set header X-Simulate-Payment-Timeout: true to simulate gateway timeout (for testing retry).

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Printf("Snake server running at http://localhost:%s\n", port)
	http.ListenAndServe(":"+port, nil)
}
