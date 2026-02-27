BASE="http://localhost:8080"


echo "=== 1. GET /api/player ==="  # Get player information
curl -s -X GET "$BASE/api/player" | head -c 500
echo -e "\n"

echo "=== 2. POST /api/earn (earn coins by score) ==="
curl -s -X POST "$BASE/api/earn" \
  -H "Content-Type: application/json" \
  -d '{"score": 50}'
echo -e "\n"

echo "=== 3. POST /api/equip (equip skin) ==="
curl -s -X POST "$BASE/api/equip" \
  -H "Content-Type: application/json" \
  -d '{"skinId": "default"}'
echo -e "\n"

echo "=== 4. POST /api/user/cart/items (add skin to cart) ==="
curl -s -X POST "$BASE/api/user/cart/items" \
  -H "Content-Type: application/json" \
  -d '{"itemId": "skin_gold"}'
echo -e "\n"

echo "=== 5. POST /api/user/cart/items (add extra life) ==="
curl -s -X POST "$BASE/api/user/cart/items" \
  -H "Content-Type: application/json" \
  -d '{"itemId": "extra_life"}'
echo -e "\n"

echo "=== 6. GET /api/user/cart ==="
CART=$(curl -s -X GET "$BASE/api/user/cart")
echo "$CART"
# Extract first cart item id for PATCH/DELETE (requires jq; otherwise skip or use a fixed id)
CART_ITEM_ID=$(echo "$CART" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
echo -e "\n"

echo "=== 7. PATCH /api/user/cart/items/{id} (update quantity; id from cart) ==="
if [ -n "$CART_ITEM_ID" ]; then
  curl -s -X PATCH "$BASE/api/user/cart/items/$CART_ITEM_ID" \
    -H "Content-Type: application/json" \
    -d '{"quantity": 2}'
else
  echo "Skip: no cart item id (add item first or set CART_ITEM_ID manually)"
fi
echo -e "\n"

echo "=== 8. GET /api/user/cart (after PATCH) ==="
curl -s -X GET "$BASE/api/user/cart"
echo -e "\n"

echo "=== 9. POST /api/user/orders (checkout) ==="
curl -s -X POST "$BASE/api/user/orders" \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: test-checkout-$(date +%s)"
echo -e "\n"

echo "=== 10. POST /api/user/orders (idempotent retry — same key returns cached response) ==="
KEY="idem-key-$(date +%s)"
curl -s -X POST "$BASE/api/user/orders" -H "Idempotency-Key: $KEY" -H "Content-Type: application/json"
echo ""
curl -s -X POST "$BASE/api/user/orders" -H "Idempotency-Key: $KEY" -H "Content-Type: application/json"
echo -e "\n"

echo "=== 11. GET /api/player (after checkout) ==="
curl -s -X GET "$BASE/api/player"
echo -e "\n"

# --- Legacy endpoints ---
echo "=== 12. POST /api/cart (legacy — add to cart) ==="
curl -s -X POST "$BASE/api/cart" \
  -H "Content-Type: application/json" \
  -d '{"itemId": "skin_rainbow"}'
echo -e "\n"

echo "=== 13. GET /api/cart (legacy) ==="
curl -s -X GET "$BASE/api/cart"
echo -e "\n"

echo "=== 14. POST /api/cart/remove (legacy — remove by itemId) ==="
curl -s -X POST "$BASE/api/cart/remove" \
  -H "Content-Type: application/json" \
  -d '{"itemId": "skin_rainbow"}'
echo -e "\n"

echo "=== 15. DELETE /api/user/cart/items/{id} (remove one item; get id from GET cart) ==="
CART2=$(curl -s -X GET "$BASE/api/user/cart")
ID2=$(echo "$CART2" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
if [ -n "$ID2" ]; then
  curl -s -X DELETE "$BASE/api/user/cart/items/$ID2"
else
  echo "Skip: no cart item id"
fi
echo -e "\n"

echo "=== 16. POST /api/checkout (legacy — empty cart) ==="
curl -s -X POST "$BASE/api/checkout" -H "Content-Type: application/json"
echo -e "\n"

echo "=== 17. Validation: POST /api/earn invalid score ==="
curl -s -X POST "$BASE/api/earn" -H "Content-Type: application/json" -d '{"score": -1}'
echo -e "\n"

echo "=== 18. Validation: POST /api/user/cart/items invalid itemId ==="
curl -s -X POST "$BASE/api/user/cart/items" \
  -H "Content-Type: application/json" \
  -d '{"itemId": ""}'
echo -e "\n"

echo "=== 19. Optional: checkout with payment timeout simulation ==="
# Add item first
curl -s -X POST "$BASE/api/user/cart/items" -H "Content-Type: application/json" -d '{"itemId": "extra_life"}' > /dev/null
curl -s -X POST "$BASE/api/user/orders" \
  -H "Content-Type: application/json" \
  -H "X-Simulate-Payment-Timeout: true" \
  -H "Idempotency-Key: sim-timeout-$(date +%s)"
echo -e "\n"

echo "Done."
