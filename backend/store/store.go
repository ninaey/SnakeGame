package store

import (
	"crypto/rand"
	"encoding/hex"
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

// ErrCartItemNotFound is returned when a cart line id is not found.
var ErrCartItemNotFound = errors.New("cart item not found")

func newCartLineID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// AddToCart adds one unit of an item to the cart. If the same item already exists as a line, quantity is incremented.
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
	for i := range cart {
		if cart[i].ItemID == itemID {
			cart[i].Quantity++
			return nil
		}
	}
	cart = append(cart, models.CartItem{
		ID: newCartLineID(), ItemID: itemID, Name: name, Price: price, Quantity: 1, Kind: kind,
	})
	return nil
}

// GetCart returns a copy of cart items and the total price (sum of price*quantity per line).
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
		total += cart[i].Price * cart[i].Quantity
	}
	return out, total
}

// UpdateCartItem sets the quantity for the cart line with the given id. If quantity < 1, the line is removed.
func UpdateCartItem(id string, quantity int) error {
	if quantity < 1 {
		mu.Lock()
		defer mu.Unlock()
		for i := range cart {
			if cart[i].ID == id {
				cart = append(cart[:i], cart[i+1:]...)
				return nil
			}
		}
		return ErrCartItemNotFound
	}
	mu.Lock()
	defer mu.Unlock()
	for i := range cart {
		if cart[i].ID == id {
			cart[i].Quantity = quantity
			return nil
		}
	}
	return ErrCartItemNotFound
}

// RemoveCartItemByID removes the cart line with the given id. Returns true if removed.
func RemoveCartItemByID(id string) bool {
	mu.Lock()
	defer mu.Unlock()
	for i := range cart {
		if cart[i].ID == id {
			cart = append(cart[:i], cart[i+1:]...)
			return true
		}
	}
	return false
}

// RemoveCartItem removes one occurrence of the given item from the cart (first match). Kept for backward compatibility.
func RemoveCartItem(itemID string) bool {
	mu.Lock()
	defer mu.Unlock()
	for i := range cart {
		if cart[i].ItemID == itemID {
			if cart[i].Quantity > 1 {
				cart[i].Quantity--
			} else {
				cart = append(cart[:i], cart[i+1:]...)
			}
			return true
		}
	}
	return false
}

// ClearCart empties the cart (e.g. after successful checkout).
func ClearCart() {
	mu.Lock()
	defer mu.Unlock()
	cart = nil
}

// CartCount returns the number of lines in the cart.
func CartCount() int {
	mu.RLock()
	defer mu.RUnlock()
	return len(cart)
}
