package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
)

var (
	mu sync.RWMutex
	// In-memory player state (reset on server restart; can be replaced with DB)
	player = struct {
		Balance      int      `json:"balance"`
		OwnedSkins   []string `json:"ownedSkins"`
		EquippedSkin string   `json:"equippedSkin"`
	}{
		Balance:      200,
		OwnedSkins:   []string{"default"},
		EquippedSkin: "default",
	}
)

const (
	priceExtraLife = 50
	priceSkin      = 100
	coinsPerScore  = 2 // coins per 10 points
)

type PurchaseRequest struct {
	ItemID string `json:"itemId"`
}

type EquipRequest struct {
	SkinID string `json:"skinId"`
}

func allowCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func getPlayerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		return
	}
	allowCORS(w)
	w.Header().Set("Content-Type", "application/json")
	mu.RLock()
	defer mu.RUnlock()
	json.NewEncoder(w).Encode(player)
}

func earnCoinsHandler(w http.ResponseWriter, r *http.Request) {
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
	mu.Lock()
	earned := (req.Score / 10) * coinsPerScore
	player.Balance += earned
	mu.Unlock()
	json.NewEncoder(w).Encode(map[string]interface{}{
		"earned":   earned,
		"balance":  player.Balance,
	})
}

func equipHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		return
	}
	var req EquipRequest
	if json.NewDecoder(r.Body).Decode(&req) != nil || req.SkinID == "" {
		http.Error(w, `{"error":"invalid skinId"}`, http.StatusBadRequest)
		return
	}
	allowCORS(w)
	w.Header().Set("Content-Type", "application/json")
	mu.Lock()
	defer mu.Unlock()
	for _, s := range player.OwnedSkins {
		if s == req.SkinID {
			player.EquippedSkin = req.SkinID
			json.NewEncoder(w).Encode(map[string]string{"equipped": req.SkinID})
			return
		}
	}
	http.Error(w, `{"error":"skin not owned"}`, http.StatusBadRequest)
}

func checkoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		return
	}
	var req PurchaseRequest
	if json.NewDecoder(r.Body).Decode(&req) != nil || req.ItemID == "" {
		http.Error(w, `{"error":"invalid itemId"}`, http.StatusBadRequest)
		return
	}
	allowCORS(w)
	w.Header().Set("Content-Type", "application/json")
	mu.Lock()
	defer mu.Unlock()

	var cost int
	switch req.ItemID {
	case "extra_life":
		cost = priceExtraLife
		if player.Balance < cost {
			mu.Unlock()
			json.NewEncoder(w).Encode(map[string]interface{}{
				"Status": "Fail",
				"Message": "Not enough coins. Need 50.",
				"Balance": player.Balance,
			})
			return
		}
		player.Balance -= cost
		json.NewEncoder(w).Encode(map[string]interface{}{
			"Status":  "Success",
			"Message": "Extra life purchased! Use it when you lose a life.",
			"Balance": player.Balance,
			"Item":    "extra_life",
		})
		return
	case "skin_gold", "skin_rainbow", "skin_ice", "skin_fire":
		cost = priceSkin
		for _, s := range player.OwnedSkins {
			if s == req.ItemID {
				json.NewEncoder(w).Encode(map[string]interface{}{
					"Status":  "Success",
					"Message": "You already own this skin.",
					"Balance": player.Balance,
					"Item":    req.ItemID,
				})
				return
			}
		}
		if player.Balance < cost {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"Status":  "Fail",
				"Message": "Not enough coins. Skins cost 100.",
				"Balance": player.Balance,
			})
			return
		}
		player.Balance -= cost
		player.OwnedSkins = append(player.OwnedSkins, req.ItemID)
		player.EquippedSkin = req.ItemID
		json.NewEncoder(w).Encode(map[string]interface{}{
			"Status":  "Success",
			"Message": "Skin purchased and equipped!",
			"Balance": player.Balance,
			"Item":    req.ItemID,
			"OwnedSkins": player.OwnedSkins,
		})
		return
	default:
		json.NewEncoder(w).Encode(map[string]interface{}{
			"Status": "Fail",
			"Message": "Unknown item: " + req.ItemID,
		})
	}
}

func main() {
	fs := http.FileServer(http.Dir("../frontend"))
	http.Handle("/", fs)
	http.HandleFunc("/api/player", getPlayerHandler)
	http.HandleFunc("/api/earn", earnCoinsHandler)
	http.HandleFunc("/api/equip", equipHandler)
	http.HandleFunc("/api/checkout", checkoutHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Printf("Snake server running at http://localhost:%s\n", port)
	http.ListenAndServe(":"+port, nil)
}
