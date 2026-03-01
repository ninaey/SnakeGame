/**
 * store.js — Cart and store UI.
 * Depends on globals defined in index.html: state, skins, toast, apiGet, apiPost, apiPatch, apiDelete, loadPlayer
 */

const cartData = { items: [], total: 0 };

async function loadCart() {
    try {
        const c = await apiGet('/api/user/cart');
        cartData.items = c.items || [];
        cartData.total = c.total || 0;
    } catch (e) {
        cartData.items = [];
        cartData.total = 0;
    }
}

function renderCart() {
    // UPDATED: Sync the badge on the Store screen
    const itemCount = cartData.items.reduce((n, it) => n + (it.quantity || 1), 0);
    const badge = document.getElementById('cartCountBadge');
    if (badge) badge.textContent = '(' + itemCount + ')';

    // UPDATED: Fill the dedicated Cart screen elements
    const cartCoins = document.getElementById('cartCoinsDisplay');
    if (cartCoins) cartCoins.textContent = state.balance;

    const totalEl = document.getElementById('cartTotal');
    if (totalEl) totalEl.textContent = cartData.total;

    const listEl = document.getElementById('cartList');
    if (!listEl) return;

    if (cartData.items.length === 0) {
        listEl.innerHTML = '<p class="cart-empty" style="color:var(--textDim);font-size:0.9rem;">Cart is empty</p>';
        document.getElementById('btnCheckout').disabled = true;
        return;
    }

    document.getElementById('btnCheckout').disabled = state.balance < cartData.total;
    
    listEl.innerHTML = cartData.items.map(it => {
        const qty = it.quantity ?? 1;
        return `
            <div class="cart-item">
                <div class="cart-item-info">
                    <div style="font-weight:600;">${it.name || it.itemId}</div>
                    <div class="cart-item-price">${it.price} 🪙</div>
                </div>
                <div class="cart-item-qty">
                    <button class="qty-btn" onclick="updateQty('${it.itemId}', ${qty - 1})">−</button>
                    <span class="qty-num">${qty}</span>
                    <button class="qty-btn" onclick="updateQty('${it.itemId}', ${qty + 1})">+</button>
                </div>
                <button class="cart-item-remove" onclick="updateQty('${it.itemId}', 0)">Remove</button>
            </div>
        `;
    }).join('');
}

async function renderStore() {
    document.getElementById('storeCoins').textContent = state.balance;
    const livesEl = document.getElementById('storeLives');
    const skinsEl = document.getElementById('storeSkins');

    livesEl.innerHTML = `
        <div class="store-card">
            <div class="preview" style="display:flex;align-items:center;justify-content:center;font-size:2rem;">❤️</div>
            <div class="name">Extra Life</div>
            <div class="price">50 🪙</div>
            <button onclick="addToCart('extra_life', 50)">Add to Cart</button>
        </div>
    `;

    skinsEl.innerHTML = Object.entries(skins).map(([id, s]) => {
        const owned = state.ownedSkins.includes(id);
        const equipped = state.equippedSkin === id;
        const priceText = s.price === 0 ? 'FREE' : `${s.price} 🪙`;

        return `
            <div class="store-card ${owned ? 'owned' : ''} ${equipped ? 'equipped' : ''}">
                <div class="preview" style="background:${s.head}; border: 2px solid ${s.body}"></div>
                <div class="name">${s.name}</div>
                <div class="price ${s.price === 0 ? 'free' : ''}">${priceText}</div>
                ${!owned ? 
                    `<button onclick="addToCart('${id}', ${s.price})">Add to Cart</button>` : 
                    `<button class="secondary" onclick="equipSkin('${id}')" ${equipped ? 'disabled' : ''}>
                        ${equipped ? 'Equipped' : 'Equip'}
                    </button>`
                }
            </div>
        `;
    }).join('');

    await loadCart();
    renderCart();
}

async function updateQty(itemId, newQty) {
    try {
        if (newQty <= 0) {
            await apiDelete(`/api/user/cart/items/${itemId}`);
        } else {
            await apiPatch(`/api/user/cart/items/${itemId}`, { quantity: newQty });
        }
        await loadCart();
        renderCart();
    } catch (e) {
        toast('Update failed', 'error');
    }
}

async function equipSkin(skinId) {
    try {
        const res = await apiPost('/api/equip', { skinId });
        state.equippedSkin = res.EquippedSkin || skinId;
        renderStore();
        toast('Skin equipped!');
    } catch (e) {
        toast('Could not equip skin', 'error');
    }
}

async function addToCart(itemId, price) {
    // We removed the immediate balance check here so user can add to cart and see it, 
    // but the Checkout button will still be disabled if they can't afford it.
    try {
        const res = await apiPost('/api/user/cart/items', { itemId });
        cartData.items = res.items || [];
        cartData.total = res.total || 0;
        renderCart();
        toast('Added to cart', 'success');
    } catch (e) {
        toast('Could not add to cart', 'error');
    }
}

async function checkoutCart() {
    if (cartData.items.length === 0) { toast('Cart is empty', 'error'); return; }
    if (state.balance < cartData.total) { toast('Not enough coins', 'error'); return; }
    try {
        const res = await apiPost('/api/user/orders', {});
        if (res.Status === 'Success') {
            state.balance      = res.Balance      ?? state.balance;
            state.ownedSkins   = res.OwnedSkins   ?? state.ownedSkins;
            state.equippedSkin = res.EquippedSkin ?? state.equippedSkin;
            if ('ExtraLives' in res) state.extraLives = Math.max(0, Number(res.ExtraLives) || 0);
            
            await loadPlayer();
            await loadCart();
            renderCart();
            
            document.getElementById('storeCoins').textContent = state.balance;
            toast('Purchase successful!', 'success');
            showScreen('menu'); // Go back to menu after successful purchase
        }
    } catch (e) {
        toast('Checkout failed', 'error');
    }
}

// Attach the checkout function to the button in index.html
document.getElementById('btnCheckout').onclick = checkoutCart;