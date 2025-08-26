function showToast(message, severity) {
    const container = document.getElementById('toast-container');
    if (!container) { alert(message); return; }
    const id = 'toast-' + Date.now();
    const bg = severity === 'error' ? 'bg-red-100' : 'bg-green-100';
    const text = severity === 'error' ? 'text-red-800' : 'text-green-800';
    const el = document.createElement('div');
    el.id = id;
    el.className = bg + ' p-3 rounded shadow text-sm ' + text;
    el.innerText = message;
    // support htmx remove-me extension semantics by adding data-remove-after
    el.setAttribute('data-remove-after', '5000');
    container.appendChild(el);
}

// Expose to window so inline script can call it
window.showToast = showToast;

// Listen for HTMX server-triggered events named 'notify'.
// When server sets HX-Trigger: notify:{"message":"..."}, htmx will dispatch an event named 'notify'.
document.body.addEventListener('notify', function(e){
    var detail = e.detail || {};
    var msg = detail.message || (typeof detail === 'string' ? detail : JSON.stringify(detail));
    var severity = detail.severity || 'info';
    var remove = detail.remove || '5s';

    // Find the stable snackbar container
    var container = document.getElementById('htmx-snackbar');
    if (!container) {
        // fallback: append to toast-container
        showToast(msg, severity);
        return;
    }

    // Build inner HTML
    var inner = '';
    if (detail.html) {
        inner = detail.html;
    } else {
        var cls = (severity === 'error') ? 'bg-red-100 text-red-800' : 'bg-green-100 text-green-800';
        inner = '<div class="p-3 rounded shadow text-sm ' + cls + '">' + msg + '</div>';
    }

    container.innerHTML = inner;
    // set remove-me attribute for HTMX remove-me extension (e.g., "1s")
    container.setAttribute('remove-me', remove);
    // ensure hx-ext contains remove-me so extension processes it
    var hxext = container.getAttribute('hx-ext') || '';
    if (!hxext.includes('remove-me')) {
        container.setAttribute('hx-ext', (hxext + ' remove-me').trim());
    }
    // make visible
    container.classList.remove('hidden');
});
