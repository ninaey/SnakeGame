package handlers

import (
	"encoding/json"
	"net/http"
	"sync"

	"SnakeGame/models"
	"SnakeGame/store"
)

const coinsPerScore = 2 // coins per 10 points

var (
	playerMu sync.RWMutex
	player   = struct {
		Balance      int      `json:"Balance"`
		OwnedSkins   []string `json:"OwnedSkins"`
		EquippedSkin string   `json:"EquippedSkin"`
		ExtraLives   int      `json:"ExtraLives"`
	}{
		Balance:      200,
		OwnedSkins:   []string{"default"},
		EquippedSkin: "default",
	}
)

func allowCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func GetPlayerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		return
	}
	allowCORS(w)
	w.Header().Set("Content-Type", "application/json")
	playerMu.RLock()
	defer playerMu.RUnlock()
	json.NewEncoder(w).Encode(player)
}

func EarnCoinsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		return
	}
	var req struct {
		Score int `json:"score"`
	}
	if json.NewDecoder(r.Body).Decode(&req) != nil || req.Score < 0 {
		http.Error(w, `{"error":"invalid score"}`, http.StatusBadRequest)
		return
	}
	allowCORS(w)
	w.Header().Set("Content-Type", "application/json")
	playerMu.Lock()
	earned := (req.Score / 10) * coinsPerScore
	player.Balance += earned
	playerMu.Unlock()
	json.NewEncoder(w).Encode(map[string]interface{}{
		"earned":  earned,
		"balance": player.Balance,
	})
}

func EquipHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		return
	}
	var req struct {
		SkinID string `json:"skinId"`
	}
	if json.NewDecoder(r.Body).Decode(&req) != nil || req.SkinID == "" {
		http.Error(w, `{"error":"invalid skinId"}`, http.StatusBadRequest)
		return
	}
	allowCORS(w)
	w.Header().Set("Content-Type", "application/json")
	playerMu.Lock()
	defer playerMu.Unlock()
	for _, s := range player.OwnedSkins {
		if s == req.SkinID {
			player.EquippedSkin = req.SkinID
			json.NewEncoder(w).Encode(map[string]string{"equipped": req.SkinID})
			return
		}
	}
	http.Error(w, `{"error":"skin not owned"}`, http.StatusBadRequest)
}

func CartHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		return
	}
	var req struct {
		ItemID string `json:"itemId"`
	}
	if json.NewDecoder(r.Body).Decode(&req) != nil || req.ItemID == "" {
		http.Error(w, `{"error":"invalid itemId"}`, http.StatusBadRequest)
		return
	}
	allowCORS(w)
	w.Header().Set("Content-Type", "application/json")
	if err := store.AddToCart(req.ItemID); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	items, total := store.GetCart()
	itemsResp := make([]map[string]interface{}, len(items))
	for i, it := range items {
		itemsResp[i] = map[string]interface{}{
			"itemId": it.ItemID,
			"name":   it.Name,
			"price":  it.Price,
		}
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"items": itemsResp, "total": total})
}

func RemoveCartItemHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		return
	}
	var req struct {
		ItemID string `json:"itemId"`
	}
	if json.NewDecoder(r.Body).Decode(&req) != nil || req.ItemID == "" {
		http.Error(w, `{"error":"invalid itemId"}`, http.StatusBadRequest)
		return
	}
	allowCORS(w)
	w.Header().Set("Content-Type", "application/json")
	store.RemoveCartItem(req.ItemID)
	items, total := store.GetCart()
	itemsResp := make([]map[string]interface{}, len(items))
	for i, it := range items {
		itemsResp[i] = map[string]interface{}{
			"itemId": it.ItemID,
			"name":   it.Name,
			"price":  it.Price,
		}
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"items": itemsResp, "total": total})
}

// CheckoutHandler processes the cart: only charges for items the player does not
// already own (skins already in OwnedSkins are skipped). Prevents deducting coins
// for duplicate skins.
func CheckoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		return
	}
	allowCORS(w)
	w.Header().Set("Content-Type", "application/json")

	items, _ := store.GetCart()
	if len(items) == 0 {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"Status":  "Fail",
			"Message": "Cart is empty",
		})
		return
	}

	playerMu.Lock()
	defer playerMu.Unlock()

	// Compute the amount we actually charge: skip price for skins already owned
	var chargeTotal int
	for _, it := range items {
		if models.IsSkin(it.ItemID) {
			alreadyOwned := false
			for _, s := range player.OwnedSkins {
				if s == it.ItemID {
					alreadyOwned = true
					break
				}
			}
			if alreadyOwned {
				continue // do not charge again for skin already owned
			}
		}
		chargeTotal += it.Price
	}

	if player.Balance < chargeTotal {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"Status":  "Fail",
			"Message": "Not enough coins",
			"Balance": player.Balance,
		})
		return
	}

	// Deduct only the chargeable total (not cart total, which may include already-owned skins)
	player.Balance -= chargeTotal

	// Apply purchases: add new skins to OwnedSkins, count extra lives
	var lastNewSkin string
	for _, it := range items {
		if models.IsSkin(it.ItemID) {
			alreadyOwned := false
			for _, s := range player.OwnedSkins {
				if s == it.ItemID {
					alreadyOwned = true
					break
				}
			}
			if !alreadyOwned {
				player.OwnedSkins = append(player.OwnedSkins, it.ItemID)
				lastNewSkin = it.ItemID
			}
		}
		if models.IsLifeItem(it.ItemID) {
			player.ExtraLives++
		}
	}

	// Equip last newly purchased skin if any
	if lastNewSkin != "" {
		player.EquippedSkin = lastNewSkin
	}

	store.ClearCart()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"Status":       "Success",
		"Message":      "Purchase complete!",
		"Balance":      player.Balance,
		"OwnedSkins":   player.OwnedSkins,
		"EquippedSkin": player.EquippedSkin,
		"ExtraLives":   player.ExtraLives,
	})
}
