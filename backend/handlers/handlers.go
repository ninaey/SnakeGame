package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"SnakeGame/models"
	"SnakeGame/payment"
	"SnakeGame/retry"
	"SnakeGame/store"
)

const coinsPerScore = 2 // coins per 10 points
const idempotencyTTL = 24 * time.Hour

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
	idempotencyMu    sync.RWMutex
	idempotencyCache = make(map[string]*idempotencyEntry)
)

type idempotencyEntry struct {
	StatusCode int
	Body       []byte
	CreatedAt  time.Time
}

func getIdempotency(key string) (statusCode int, body []byte, ok bool) {
	if key == "" {
		return 0, nil, false
	}
	idempotencyMu.Lock()
	defer idempotencyMu.Unlock()
	ent, exists := idempotencyCache[key]
	if !exists || ent == nil {
		return 0, nil, false
	}
	if time.Since(ent.CreatedAt) > idempotencyTTL {
		delete(idempotencyCache, key)
		return 0, nil, false
	}
	bodyCopy := make([]byte, len(ent.Body))
	copy(bodyCopy, ent.Body)
	return ent.StatusCode, bodyCopy, true
}

func setIdempotency(key string, statusCode int, body []byte) {
	if key == "" {
		return
	}
	bodyCopy := make([]byte, len(body))
	copy(bodyCopy, body)
	idempotencyMu.Lock()
	defer idempotencyMu.Unlock()
	idempotencyCache[key] = &idempotencyEntry{
		StatusCode: statusCode,
		Body:       bodyCopy,
		CreatedAt:  time.Now(),
	}
}

func allowCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Idempotency-Key, X-Simulate-Payment-Timeout")
}

// writeValidationError sends a 400 Bad Request with a consistent JSON error body.
func writeValidationError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
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
		writeValidationError(w, "invalid score")
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
		writeValidationError(w, "invalid skinId")
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
	writeValidationError(w, "skin not owned")
}

// cartResponse builds the common cart JSON (items + total).
func cartResponse(items []models.CartItem, total int) map[string]interface{} {
	itemsResp := make([]map[string]interface{}, len(items))
	for i, it := range items {
		itemsResp[i] = map[string]interface{}{
			"id":       it.ID,
			"itemId":   it.ItemID,
			"name":     it.Name,
			"price":    it.Price,
			"quantity": it.Quantity,
		}
	}
	return map[string]interface{}{"items": itemsResp, "total": total}
}

// POST /api/user/cart/items — add an item to the cart
func PostCartItemsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		return
	}
	var req struct {
		ItemID string `json:"itemId"`
	}
	if json.NewDecoder(r.Body).Decode(&req) != nil || req.ItemID == "" {
		writeValidationError(w, "invalid itemId")
		return
	}
	allowCORS(w)
	w.Header().Set("Content-Type", "application/json")
	if err := store.AddToCart(req.ItemID); err != nil {
		writeValidationError(w, err.Error())
		return
	}
	items, total := store.GetCart()
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(cartResponse(items, total))
}

// GET /api/user/cart — view the contents of the cart
func GetCartHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		return
	}
	allowCORS(w)
	w.Header().Set("Content-Type", "application/json")
	items, total := store.GetCart()
	json.NewEncoder(w).Encode(cartResponse(items, total))
}

// PATCH /api/user/cart/items/:id — update an item (e.g. change quantity)
func PatchCartItemHandler(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodPatch {
		return
	}
	var req struct {
		Quantity *int `json:"quantity"`
	}
	if json.NewDecoder(r.Body).Decode(&req) != nil || req.Quantity == nil {
		writeValidationError(w, "quantity required")
		return
	}
	allowCORS(w)
	w.Header().Set("Content-Type", "application/json")
	if err := store.UpdateCartItem(id, *req.Quantity); err != nil {
		if err == store.ErrCartItemNotFound {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "cart item not found"})
			return
		}
		writeValidationError(w, err.Error())
		return
	}
	items, total := store.GetCart()
	json.NewEncoder(w).Encode(cartResponse(items, total))
}

// DELETE /api/user/cart/items/:id — remove an item from the cart
func DeleteCartItemHandler(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodDelete {
		return
	}
	allowCORS(w)
	w.Header().Set("Content-Type", "application/json")
	if !store.RemoveCartItemByID(id) {
		http.Error(w, `{"error":"cart item not found"}`, http.StatusNotFound)
		return
	}
	items, total := store.GetCart()
	json.NewEncoder(w).Encode(cartResponse(items, total))
}

// CartItemsIDHandler routes PATCH and DELETE by path /api/user/cart/items/<id>
func CartItemsIDHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		allowCORS(w)
		w.WriteHeader(http.StatusNoContent)
		return
	}
	id := r.PathValue("id")
	if id == "" {
		writeValidationError(w, "id required")
		return
	}
	switch r.Method {
	case http.MethodPatch:
		PatchCartItemHandler(w, r, id)
	case http.MethodDelete:
		DeleteCartItemHandler(w, r, id)
	}
}

// Legacy handlers for backward compatibility (old frontend paths)
func CartHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		return
	}
	var req struct {
		ItemID string `json:"itemId"`
	}
	if json.NewDecoder(r.Body).Decode(&req) != nil || req.ItemID == "" {
		writeValidationError(w, "invalid itemId")
		return
	}
	allowCORS(w)
	w.Header().Set("Content-Type", "application/json")
	if err := store.AddToCart(req.ItemID); err != nil {
		writeValidationError(w, err.Error())
		return
	}
	items, total := store.GetCart()
	json.NewEncoder(w).Encode(cartResponse(items, total))
}

func RemoveCartItemHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		return
	}
	var req struct {
		ItemID string `json:"itemId"`
	}
	if json.NewDecoder(r.Body).Decode(&req) != nil || req.ItemID == "" {
		writeValidationError(w, "invalid itemId")
		return
	}
	allowCORS(w)
	w.Header().Set("Content-Type", "application/json")
	store.RemoveCartItem(req.ItemID)
	items, total := store.GetCart()
	json.NewEncoder(w).Encode(cartResponse(items, total))
}

// doCheckout runs the checkout logic and returns the HTTP status code and response body.
// It validates first (400 on empty cart or insufficient balance), then calls the payment
// gateway with retry (exponential backoff). Stop conditions: success, non-retryable error,
// max attempts, or context cancelled. Used so the response can be cached for idempotency.
func doCheckout(ctx context.Context, gw payment.Gateway, idempotencyKey string) (statusCode int, body []byte) {
	statusCode = http.StatusOK
	items, _ := store.GetCart()
	if len(items) == 0 {
		out := map[string]interface{}{
			"Status":  "Fail",
			"Message": "Cart is empty",
		}
		body, _ = json.Marshal(out)
		return statusCode, body
	}

	playerMu.Lock()
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
				continue
			}
		}
		chargeTotal += it.Price * it.Quantity
	}
	if player.Balance < chargeTotal {
		playerMu.Unlock()
		out := map[string]interface{}{
			"Status":  "Fail",
			"Message": "Not enough coins",
			"Balance": player.Balance,
		}
		body, _ = json.Marshal(out)
		return statusCode, body
	}
	playerMu.Unlock()

	// Call payment gateway with retry (exponential backoff). Stop conditions: success,
	// non-retryable error, max attempts, or context cancelled.
	cfg := retry.DefaultConfig()
	cfg.MaxAttempts = 5
	cfg.InitialDelay = 100 * time.Millisecond
	cfg.MaxDelay = 5 * time.Second
	err := retry.Do(ctx, cfg, func() error {
		return gw.Charge(ctx, chargeTotal, idempotencyKey)
	})
	if err != nil {
		// Retries exhausted or non-retryable
		statusCode = http.StatusServiceUnavailable
		out := map[string]interface{}{
			"Status":  "Fail",
			"Message": "Payment temporarily unavailable. Please try again.",
		}
		body, _ = json.Marshal(out)
		return statusCode, body
	}

	// Gateway succeeded: apply balance deduction and purchases
	playerMu.Lock()
	defer playerMu.Unlock()

	player.Balance -= chargeTotal

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
		if it.ItemID == "extra_life" {
			player.ExtraLives += it.Quantity
		}
	}
	if lastNewSkin != "" {
		player.EquippedSkin = lastNewSkin
	}

	store.ClearCart()

	out := map[string]interface{}{
		"Status":       "Success",
		"Message":      "Purchase complete!",
		"Balance":      player.Balance,
		"OwnedSkins":   player.OwnedSkins,
		"EquippedSkin": player.EquippedSkin,
		"ExtraLives":   player.ExtraLives,
	}
	body, _ = json.Marshal(out)
	return statusCode, body
}

// CheckoutHandler processes the cart: only charges for items the player does not
// already own (skins already in OwnedSkins are skipped). Prevents deducting coins
// for duplicate skins. Uses Idempotency-Key header: repeated requests with the
// same key within 24 hours receive the cached response without re-processing.
// Set header X-Simulate-Payment-Timeout: true to simulate gateway timeout (for testing retry).
func CheckoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		return
	}
	allowCORS(w)
	w.Header().Set("Content-Type", "application/json")

	key := r.Header.Get("Idempotency-Key")
	if key != "" {
		if status, cached, ok := getIdempotency(key); ok {
			w.WriteHeader(status)
			w.Write(cached)
			return
		}
	}

	gw := &payment.StubGateway{
		SimulateTimeout: r.Header.Get("X-Simulate-Payment-Timeout") == "true",
	}
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	status, body := doCheckout(ctx, gw, key)
	if key != "" {
		setIdempotency(key, status, body)
	}
	w.WriteHeader(status)
	w.Write(body)
}
