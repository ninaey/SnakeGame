/**
 * Store module: cart and store UI. Uses globals: state, API, toast, apiGet, apiPost, apiPatch, apiDelete, skins, loadPlayer
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
    const itemCount = cartData.items.reduce((n, it) => n + (it.quantity || 1), 0);
    document.getElementById('cartCount').textContent = '(' + itemCount + ')';
    document.getElementById('cartTotal').textContent = cartData.total;
    const listEl = document.getElementById('cartList');
    if (cartData.items.length === 0) {
        listEl.innerHTML = '<p class="cart-empty" style="color:var(--textDim);font-size:0.9rem;">Cart is empty</p>';
        document.getElementById('btnCheckout').disabled = true;
        return;
    }
    document.getElementById('btnCheckout').disabled = state.balance < cartData.total;
    listEl.innerHTML = cartData.items.map((it) => {
        const qty = it.quantity ?? 1;
        const lineTotal = (it.price || 0) * qty;
        return `<div class="cart-item" data-id="${it.id}">
            <span class="cart-item-info">${it.name}${qty > 1 ? ' ×' + qty : ''}</span>
            <span class="cart-item-price">${lineTotal} 🪙</span>
            <div class="cart-item-qty">
                <button type="button" class="qty-btn" data-id="${it.id}" data-delta="-1">−</button>
                <span class="qty-num">${qty}</span>
                <button type="button" class="qty-btn" data-id="${it.id}" data-delta="1">+</button>
            </div>
            <button type="button" class="cart-item-remove" data-id="${it.id}">Remove</button>
        </div>`;
    }).join('');
    listEl.querySelectorAll('.cart-item-remove').forEach(btn => {
        btn.onclick = async () => {
            try {
                await apiDelete('/api/user/cart/items/' + btn.dataset.id);
                await loadCart();
                renderCart();
                toast('Removed from cart', 'success');
            } catch (e) { toast('Could not remove', 'error'); }
        };
    });
    listEl.querySelectorAll('.qty-btn').forEach(btn => {
        btn.onclick = async () => {
            const id = btn.dataset.id;
            const delta = parseInt(btn.dataset.delta, 10);
            const it = cartData.items.find(i => i.id === id);
            if (!it) return;
            const newQty = Math.max(1, (it.quantity || 1) + delta);
            try {
                await apiPatch('/api/user/cart/items/' + id, { quantity: newQty });
                await loadCart();
                renderCart();
                toast('Cart updated', 'success');
            } catch (e) { toast('Could not update quantity', 'error'); }
        };
    });
}

async function renderStore() {
    document.getElementById('storeCoins').textContent = state.balance;
    // LIVES section: only Extra Life (Speed Boost, Shield, Score Multiplier removed)
    const consumables = [
        { id: 'extra_life', name: 'Extra Life', price: 50, color: '135deg, #ff4757, #ff6b81' }
    ];
    const livesEl = document.getElementById('storeLives');
    livesEl.innerHTML = consumables.map(c =>
        `<div class="store-card">
            <div class="preview" style="background: linear-gradient(${c.color});"></div>
            <div class="name">${c.name}</div>
            <div class="price">${c.price} 🪙</div>
            <button id="buy-${c.id}">Add to cart</button>
        </div>`
    ).join('');
    consumables.forEach(c => {
        const btn = document.getElementById('buy-' + c.id);
        if (btn) btn.onclick = () => addToCart(c.id, c.price);
    });

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
    const btnCheckout = document.getElementById('btnCheckout');
    if (btnCheckout) btnCheckout.onclick = checkoutCart;
}

async function addToCart(itemId, price) {
    if (state.balance < price) { toast('Not enough coins', 'error'); return; }
    try {
        const res = await apiPost('/api/user/cart/items', { itemId });
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
        const res = await apiPost('/api/user/orders', {});
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
