// simple nostr-auth.js stub

async function checkNostrExtension() {
    return typeof window.nostr !== 'undefined';
}

async function getPublicKey() {
    if (!await checkNostrExtension()) throw new Error('Nostr extension not found');
    return await window.nostr.getPublicKey();
}

async function signEvent(event) {
    if (!await checkNostrExtension()) throw new Error('Nostr extension not found');
    return await window.nostr.signEvent(event);
}

// Expose for pages
window.nostrAuth = {
    check: checkNostrExtension,
    getPublicKey,
    signEvent
};
