/**
 * Store module: player balance, skins catalog, cart, and store UI.
 * Depends on globals: state, API, toast
 */

const skins = {
    default: { name: 'Default', price: 0, head: '#2ed573', body: '#27ae60', eye: '#1a1a20' },
    skin_gold: { name: 'Gold', price: 100, head: '#ffd700', body: '#daa520', eye: '#1a1a20' },
    skin_rainbow: { name: 'Rainbow', price: 100, head: '#ff6b9d', body: '#c44dff', eye: '#1a1a20', gradient: true },
    skin_ice: { name: 'Ice', price: 100, head: '#87ceeb', body: '#b0e0e6', eye: '#1a1a20' },
    skin_fire: { name: 'Fire', price: 100, head: '#ff6b35', body: '#f7931e', eye: '#1a1a20' }
};

function getSkin(id) {
    return skins[id] || skins.default;
}

async function loadPlayer() {
    try {
        const p = await apiGet('/api/player');
        state.balance = p.Balance ?? 0;
        state.ownedSkins = p.OwnedSkins || ['default'];
        state.equippedSkin = p.EquippedSkin || 'default';
    } catch (e) {
        state.balance = 0;
        state.ownedSkins = ['default'];
        state.equippedSkin = 'default';
    }
}

const cartData = { items: [], total: 0 };

async function loadCart() {
    try {
        const c = await apiGet('/api/cart');
        cartData.items = c.items || [];
        cartData.total = c.total || 0;
    } catch (e) {
        cartData.items = [];
        cartData.total = 0;
    }
}

function renderCart() {
    document.getElementById('cartCount').textContent = '(' + cartData.items.length + ')';
    document.getElementById('cartTotal').textContent = cartData.total;
    const listEl = document.getElementById('cartList');
    if (cartData.items.length === 0) {
        listEl.innerHTML = '<p class="cart-empty" style="color:var(--textDim);font-size:0.9rem;">Cart is empty</p>';
        document.getElementById('btnCheckout').disabled = true;
        return;
    }
    document.getElementById('btnCheckout').disabled = state.balance < cartData.total;
    listEl.innerHTML = cartData.items.map((it, i) =>
        `<div class="cart-item" data-item-id="${it.itemId}" data-index="${i}">
            <span class="cart-item-info">${it.name}</span>
            <span class="cart-item-price">${it.price} 🪙</span>
            <button type="button" class="cart-item-remove" data-item-id="${it.itemId}">Remove</button>
        </div>`
    ).join('');
    listEl.querySelectorAll('.cart-item-remove').forEach(btn => {
        btn.onclick = async () => {
            try {
                await apiPost('/api/cart/remove', { itemId: btn.dataset.itemId });
                await loadCart();
                renderCart();
                toast('Removed from cart', 'success');
            } catch (e) { toast('Could not remove', 'error'); }
        };
    });
}

async function renderStore() {
    document.getElementById('storeCoins').textContent = state.balance;
    const livesEl = document.getElementById('storeLives');
    livesEl.innerHTML = `
        <div class="store-card">
            <div class="preview" style="background: linear-gradient(135deg, #ff4757, #ff6b81);"></div>
            <div class="name">Extra Life</div>
            <div class="price">50 🪙</div>
            <button id="buyLife">Add to cart</button>
        </div>`;
    document.getElementById('buyLife').onclick = () => addToCart('extra_life', 50);

    const skinsEl = document.getElementById('storeSkins');
    skinsEl.innerHTML = Object.entries(skins).map(([id, s]) => {
        const owned = state.ownedSkins.includes(id);
        const equipped = state.equippedSkin === id;
        const priceText = s.price === 0 ? 'Free' : s.price + ' 🪙';
        return `
            <div class="store-card ${owned ? 'owned' : ''} ${equipped ? 'equipped' : ''}">
                <div class="preview" style="background: linear-gradient(135deg, ${s.head}, ${s.body});"></div>
                <div class="name">${s.name}</div>
                <div class="price ${s.price === 0 ? 'free' : ''}">${priceText}</div>
                ${!owned ? `<button id="buy-${id}" ${state.balance < s.price ? 'disabled' : ''}>Add to cart</button>` : `<button class="secondary" id="equip-${id}" ${equipped ? 'disabled' : ''}>${equipped ? 'Equipped' : 'Equip'}</button>`}
            </div>`;
    }).join('');

    Object.keys(skins).forEach(id => {
        const s = skins[id];
        const owned = state.ownedSkins.includes(id);
        const equipped = state.equippedSkin === id;
        const buyBtn = document.getElementById('buy-' + id);
        const equipBtn = document.getElementById('equip-' + id);
        if (buyBtn) buyBtn.onclick = () => addToCart(id, s.price);
        if (equipBtn && !equipped) equipBtn.onclick = async () => {
            try {
                await apiPost('/api/equip', { skinId: id });
                state.equippedSkin = id;
                renderStore();
                toast('Skin equipped!', 'success');
            } catch (e) { toast('Could not equip', 'error'); }
        };
    });
    await loadCart();
    renderCart();
}

async function addToCart(itemId, price) {
    if (state.balance < price) { toast('Not enough coins', 'error'); return; }
    try {
        const res = await apiPost('/api/cart', { itemId });
        cartData.items = res.items || [];
        cartData.total = res.total || 0;
        renderCart();
        toast('Added to cart', 'success');
    } catch (e) {
        toast((e && e.json && e.json.error) || (e && e.message) || 'Could not add to cart', 'error');
    }
}

async function checkoutCart() {
    if (cartData.items.length === 0) { toast('Cart is empty', 'error'); return; }
    if (state.balance < cartData.total) { toast('Not enough coins', 'error'); return; }
    try {
        const res = await apiPost('/api/checkout', {});
        if (res.Status === 'Success') {
            state.balance = res.Balance ?? state.balance;
            state.ownedSkins = res.OwnedSkins ?? state.ownedSkins;
            state.equippedSkin = res.EquippedSkin ?? state.equippedSkin;
            // Use server ExtraLives as source of truth (never default to 1; 0 or missing => 0)
            if ('ExtraLives' in res) state.extraLives = Math.max(0, Number(res.ExtraLives) || 0);
            await loadPlayer();
            await loadCart();
            renderCart();
            document.getElementById('storeCoins').textContent = state.balance;
            toast(res.Message || 'Purchase complete!', 'success');
        } else {
            toast(res.Message || 'Checkout failed', 'error');
        }
    } catch (e) { toast('Checkout failed', 'error'); }
}
