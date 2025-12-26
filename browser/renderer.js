const urlInput = document.getElementById('url');
const kbdContainer = document.getElementById('keyboard-container');
const viewPlaceholder = document.getElementById('view-placeholder');
const tabsEl = document.getElementById('tabs');

// Fonction CRUCIALE : Calcule la zone d'affichage du site
function updateViewBounds() {
  const rect = viewPlaceholder.getBoundingClientRect();
  window.electronAPI.setBounds({
    x: Math.floor(rect.left),
    y: Math.floor(rect.top),
    width: Math.floor(rect.width),
    height: Math.floor(rect.height)
  });
}

// Surveiller le redimensionnement de la fenêtre
window.addEventListener('resize', updateViewBounds);

// Toggle Clavier
document.getElementById('toggleKbd').onclick = () => {
  kbdContainer.classList.toggle('hidden');
  // On attend un peu que le CSS s'applique avant de recalculer
  setTimeout(updateViewBounds, 50);
};

window.electronAPI.onKeyboardToggleRequest(() => {
    kbdContainer.classList.toggle('hidden');
    setTimeout(updateViewBounds, 50);
});

// Envoi des touches
document.querySelectorAll('.key').forEach(btn => {
  btn.addEventListener('mousedown', (e) => {
    e.preventDefault();
    const key = btn.getAttribute('data-key') || btn.innerText.toLowerCase();

    if (document.activeElement === urlInput) {
      if (key === 'backspace') urlInput.value = urlInput.value.slice(0, -1);
      else if (key === 'enter') window.electronAPI.loadUrl(urlInput.value);
      else urlInput.value += (key === ' ' ? ' ' : key);
    } else {
      let sendKey = key;
      if (key === 'effacer') sendKey = 'Backspace';
      if (key === 'entrer') sendKey = 'Enter';
      window.electronAPI.typeKey(sendKey);
    }
  });
});

/* --- Autres contrôles --- */
document.getElementById('back').onclick = () => window.electronAPI.goBack();
document.getElementById('forward').onclick = () => window.electronAPI.goForward();
document.getElementById('reload').onclick = () => window.electronAPI.reload();
document.getElementById('newTab').onclick = () => window.electronAPI.newTab();
document.getElementById('quit').onclick = () => window.electronAPI.quitApp();

urlInput.addEventListener('keydown', (e) => {
  if (e.key === 'Enter') {
    let url = urlInput.value.trim();
    if (!/^https?:\/\//i.test(url)) url = 'https://' + url;
    window.electronAPI.loadUrl(url);
  }
});

window.electronAPI.onUrlChanged((url) => { urlInput.value = url || ''; });

window.electronAPI.onTabsUpdated((tabs) => {
  tabsEl.innerHTML = '';
  tabs.forEach((t) => {
    const el = document.createElement('div');
    el.className = `tab ${t.active ? 'active' : ''}`;
    el.innerHTML = `<span>${t.title || 'Onglet'}</span>`;
    el.onclick = () => window.electronAPI.switchTab(t.index);
    const close = document.createElement('button');
    close.textContent = '✕';
    close.onclick = (e) => { e.stopPropagation(); window.electronAPI.closeTab(t.index); };
    el.appendChild(close);
    tabsEl.appendChild(el);
  });
});

// Initialisation au chargement
setTimeout(updateViewBounds, 100);
