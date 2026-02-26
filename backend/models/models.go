package models

import "sync"

// Skin represents a purchasable snake skin.
type Skin struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Price int    `json:"price"`
}

// LifeItem represents a consumable item (e.g. extra life).
type LifeItem struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Price int    `json:"price"`
}

// ItemKind is the type of purchasable item.
type ItemKind int

const (
	ItemKindSkin ItemKind = iota
	ItemKindLife
)

// Item is a generic purchasable (skin or life item) for catalog lookups.
type Item struct {
	ID    string   `json:"id"`
	Name  string   `json:"name"`
	Price int      `json:"price"`
	Kind  ItemKind `json:"kind"`
}

// CartItem is a single line in the cart (unique line id, item id, display name, price, quantity, kind).
type CartItem struct {
	ID       string   `json:"id"`
	ItemID   string   `json:"itemId"`
	Name     string   `json:"name"`
	Price    int      `json:"price"`
	Quantity int      `json:"quantity"`
	Kind     ItemKind `json:"kind"`
}

var (
	mu sync.RWMutex

	// Skins catalog: id -> Skin
	Skins = map[string]Skin{
		"default":      {ID: "default", Name: "Default", Price: 0},
		"skin_gold":    {ID: "skin_gold", Name: "Gold", Price: 100},
		"skin_rainbow": {ID: "skin_rainbow", Name: "Rainbow", Price: 100},
		"skin_ice":     {ID: "skin_ice", Name: "Ice", Price: 100},
		"skin_fire":    {ID: "skin_fire", Name: "Fire", Price: 100},
	}

	// LifeItems catalog: id -> LifeItem (consumables that can have quantity in cart)
	LifeItems = map[string]LifeItem{
		"extra_life":       {ID: "extra_life", Name: "Extra Life", Price: 50},
		"speed_boost":      {ID: "speed_boost", Name: "Speed Boost", Price: 30},
		"shield":           {ID: "shield", Name: "Shield", Price: 40},
		"score_multiplier": {ID: "score_multiplier", Name: "Score Multiplier", Price: 35},
	}
)

// SkinPrice returns the price for a skin by id. Second return is false if not found.
func SkinPrice(id string) (int, bool) {
	mu.RLock()
	defer mu.RUnlock()
	s, ok := Skins[id]
	return s.Price, ok
}

// LifeItemPrice returns the price for a life item by id. Second return is false if not found.
func LifeItemPrice(id string) (int, bool) {
	mu.RLock()
	defer mu.RUnlock()
	item, ok := LifeItems[id]
	return item.Price, ok
}

// ItemPrice returns the price for any item (skin or life) by id. Second return is false if not found.
func ItemPrice(id string) (int, bool) {
	if p, ok := SkinPrice(id); ok {
		return p, true
	}
	return LifeItemPrice(id)
}

// IsSkin returns true if id is a known skin.
func IsSkin(id string) bool {
	mu.RLock()
	defer mu.RUnlock()
	_, ok := Skins[id]
	return ok
}

// IsLifeItem returns true if id is a known life item.
func IsLifeItem(id string) bool {
	mu.RLock()
	defer mu.RUnlock()
	_, ok := LifeItems[id]
	return ok
}

// AllSkins returns a copy of all skins (e.g. for API listing).
func AllSkins() []Skin {
	mu.RLock()
	defer mu.RUnlock()
	out := make([]Skin, 0, len(Skins))
	for _, s := range Skins {
		out = append(out, s)
	}
	return out
}

// AllLifeItems returns a copy of all life items.
func AllLifeItems() []LifeItem {
	mu.RLock()
	defer mu.RUnlock()
	out := make([]LifeItem, 0, len(LifeItems))
	for _, item := range LifeItems {
		out = append(out, item)
	}
	return out
}

// ItemDisplay returns name, price, and kind for a catalog item by id. ok is false if unknown.
func ItemDisplay(id string) (name string, price int, kind ItemKind, ok bool) {
	mu.RLock()
	defer mu.RUnlock()
	if s, okSkin := Skins[id]; okSkin {
		return s.Name, s.Price, ItemKindSkin, true
	}
	if l, okLife := LifeItems[id]; okLife {
		return l.Name, l.Price, ItemKindLife, true
	}
	return "", 0, 0, false
}
