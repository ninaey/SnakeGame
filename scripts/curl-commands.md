# Curl commands to test Snake Game API

Base URL: `http://localhost:8080` (or set `$BASE`).

---

## Player

**Get player**
```bash
curl -s -X GET http://localhost:8080/api/player
```

**Earn coins (by score; 2 coins per 10 points)**
```bash
curl -s -X POST http://localhost:8080/api/earn \
  -H "Content-Type: application/json" \
  -d "{\"score\": 50}"
```

**Equip skin**
```bash
curl -s -X POST http://localhost:8080/api/equip \
  -H "Content-Type: application/json" \
  -d "{\"skinId\": \"default\"}"
```

---

## Cart (REST)

**Add item to cart**
```bash
curl -s -X POST http://localhost:8080/api/user/cart/items \
  -H "Content-Type: application/json" \
  -d "{\"itemId\": \"skin_gold\"}"
```
Item IDs: `skin_gold`, `skin_rainbow`, `skin_ice`, `skin_fire`, `extra_life`

**Get cart**
```bash
curl -s -X GET http://localhost:8080/api/user/cart
```

**Update cart item quantity** (replace `{CART_ITEM_ID}` with `id` from GET cart)
```bash
curl -s -X PATCH http://localhost:8080/api/user/cart/items/{CART_ITEM_ID} \
  -H "Content-Type: application/json" \
  -d "{\"quantity\": 2}"
```

**Remove cart item by id**
```bash
curl -s -X DELETE http://localhost:8080/api/user/cart/items/{CART_ITEM_ID}
```

---

## Checkout

**Checkout (create order)**
```bash
curl -s -X POST http://localhost:8080/api/user/orders \
  -H "Content-Type: application/json"
```

**Checkout with idempotency key** (repeat same key → same cached response)
```bash
curl -s -X POST http://localhost:8080/api/user/orders \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: my-unique-key-123"
```

**Simulate payment timeout** (for testing retries)
```bash
curl -s -X POST http://localhost:8080/api/user/orders \
  -H "Content-Type: application/json" \
  -H "X-Simulate-Payment-Timeout: true" \
  -H "Idempotency-Key: test-timeout-1"
```

---

## Legacy endpoints

**Add to cart (legacy)**
```bash
curl -s -X POST http://localhost:8080/api/cart \
  -H "Content-Type: application/json" \
  -d "{\"itemId\": \"skin_rainbow\"}"
```

**Get cart (legacy)**
```bash
curl -s -X GET http://localhost:8080/api/cart
```

**Remove from cart by itemId (legacy)**
```bash
curl -s -X POST http://localhost:8080/api/cart/remove \
  -H "Content-Type: application/json" \
  -d "{\"itemId\": \"skin_rainbow\"}"
```

**Checkout (legacy)**
```bash
curl -s -X POST http://localhost:8080/api/checkout \
  -H "Content-Type: application/json"
```

---

## Validation / error cases

**Invalid score**
```bash
curl -s -X POST http://localhost:8080/api/earn \
  -H "Content-Type: application/json" \
  -d "{\"score\": -1}"
```

**Invalid / empty itemId**
```bash
curl -s -X POST http://localhost:8080/api/user/cart/items \
  -H "Content-Type: application/json" \
  -d "{\"itemId\": \"\"}"
```

**Equip skin not owned**
```bash
curl -s -X POST http://localhost:8080/api/equip \
  -H "Content-Type: application/json" \
  -d "{\"skinId\": \"skin_fire\"}"
```
(Expect 400 if you don’t own `skin_fire`.)
