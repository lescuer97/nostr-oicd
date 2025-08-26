// NIP-07 helper utilities used by alternative flows or debugging.

// Note: The current login page and fragment do not rely on these helpers directly.

async function checkNostrExtension() {
    return typeof window.nostr !== 'undefined' && typeof window.nostr.signEvent === 'function';
}

async function getPublicKey() {
    if (!await checkNostrExtension()) throw new Error('Nostr extension not found');
    return await window.nostr.getPublicKey();
}

async function signEvent(event) {
    if (!await checkNostrExtension()) throw new Error('Nostr extension not found');
    return await window.nostr.signEvent(event);
}

// Optional single-click login utility (not used by current HTMX flow).
// Left for convenience and potential future use in a non-HTMX flow.
async function singleClickLogin() {
    try {
        if (!await checkNostrExtension()) {
            alert('Nostr NIP-07 extension not found');
            return;
        }
        const res = await fetch('/api/auth/challenge?json=1', { credentials: 'same-origin' });
        if (!res.ok) throw new Error('failed to fetch challenge');
        const data = await res.json();
        const challenge = data.challenge;
        const ev = {
            kind: 2222,
            content: challenge,
            created_at: Math.floor(Date.now()/1000)
        };
        const signed = await signEvent(ev);
        const form = new FormData();
        form.append('signed_event', JSON.stringify(signed));
        const loginRes = await fetch('/api/auth/login', { method: 'POST', body: form, credentials: 'same-origin' });
        const text = await loginRes.text();
        const card = document.querySelector('#login-card');
        if (card) card.outerHTML = text;
    } catch (e) {
        alert('login failed: ' + e);
    }
}

// Expose for potential manual testing
window.nostrAuth = {
    check: checkNostrExtension,
    getPublicKey,
    signEvent,
    singleClickLogin
};
