const tabsEl = document.getElementById('tabs');
const urlInput = document.getElementById('url');

document.getElementById('newTab').onclick = () => window.electronAPI.newTab();
document.getElementById('back').onclick = () => window.electronAPI.goBack();
document.getElementById('forward').onclick = () => window.electronAPI.goForward();
document.getElementById('reload').onclick = () => window.electronAPI.reload();
document.getElementById('clear').onclick = () => window.electronAPI.clearSession();
document.getElementById('quit').onclick = () => window.electronAPI.quitApp();

urlInput.addEventListener('keydown', (e) => {
  if (e.key === 'Enter') {
    let url = urlInput.value.trim();
    if (!/^https?:\/\//i.test(url)) url = 'https://' + url;
    window.electronAPI.loadUrl(url);
  }
});

window.electronAPI.onUrlChanged((url) => {
  urlInput.value = url || '';
});

window.electronAPI.onTabsUpdated((tabs) => {
  tabsEl.innerHTML = '';

  // ✅ activer / désactiver la barre d’URL
  urlInput.disabled = tabs.length === 0;

  if (tabs.length === 0) {
    urlInput.value = '';
  }

  tabs.forEach((t) => {
    const el = document.createElement('div');
    el.className = `tab ${t.active ? 'active' : ''}`;
    el.onclick = () => window.electronAPI.switchTab(t.index);

    if (t.favicon) {
      const img = document.createElement('img');
      img.src = t.favicon;
      img.width = 16;
      img.height = 16;
      el.appendChild(img);
    }

    const title = document.createElement('span');
    title.textContent = t.title || 'Onglet';
    el.appendChild(title);

    const close = document.createElement('button');
    close.textContent = '✕';
    close.onclick = (e) => {
      e.stopPropagation();
      window.electronAPI.closeTab(t.index);
    };

    el.appendChild(close);
    tabsEl.appendChild(el);
  });
});

