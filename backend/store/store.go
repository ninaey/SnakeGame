package store

import (
	"errors"
	"sync"

	"SnakeGame/models"
)

var (
	mu   sync.RWMutex
	cart []models.CartItem
)

// ErrUnknownItem is returned when adding an item not in the catalog.
var ErrUnknownItem = errors.New("unknown item")

// ErrDefaultSkin is returned when trying to add the free default skin to cart.
var ErrDefaultSkin = errors.New("default skin cannot be purchased")

// AddToCart adds one unit of an item to the cart. Skins and life items are valid.
func AddToCart(itemID string) error {
	name, price, kind, ok := models.ItemDisplay(itemID)
	if !ok {
		return ErrUnknownItem
	}
	if itemID == "default" && price == 0 {
		return ErrDefaultSkin
	}
	mu.Lock()
	defer mu.Unlock()
	cart = append(cart, models.CartItem{ItemID: itemID, Name: name, Price: price, Kind: kind})
	return nil
}

// GetCart returns a copy of cart items and the total price.
func GetCart() ([]models.CartItem, int) {
	mu.RLock()
	defer mu.RUnlock()
	if len(cart) == 0 {
		return nil, 0
	}
	out := make([]models.CartItem, len(cart))
	var total int
	for i := range cart {
		out[i] = cart[i]
		total += cart[i].Price
	}
	return out, total
}

// RemoveCartItem removes one occurrence of the given item from the cart (first match).
// Returns true if an item was removed.
func RemoveCartItem(itemID string) bool {
	mu.Lock()
	defer mu.Unlock()
	for i := range cart {
		if cart[i].ItemID == itemID {
			cart = append(cart[:i], cart[i+1:]...)
			return true
		}
	}
	return false
}

// RemoveCartItemAt removes the cart line at index i. Returns false if index out of range.
func RemoveCartItemAt(i int) bool {
	mu.Lock()
	defer mu.Unlock()
	if i < 0 || i >= len(cart) {
		return false
	}
	cart = append(cart[:i], cart[i+1:]...)
	return true
}

// ClearCart empties the cart (e.g. after successful checkout).
func ClearCart() {
	mu.Lock()
	defer mu.Unlock()
	cart = nil
}

// CartCount returns the number of items in the cart.
func CartCount() int {
	mu.RLock()
	defer mu.RUnlock()
	return len(cart)
}
