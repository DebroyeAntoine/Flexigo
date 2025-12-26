const { contextBridge, ipcRenderer } = require('electron');

contextBridge.exposeInMainWorld('electronAPI', {
  newTab: () => ipcRenderer.send('new-tab'),
  switchTab: (i) => ipcRenderer.send('switch-tab', i),
  closeTab: (i) => ipcRenderer.send('close-tab', i),

  loadUrl: (url) => ipcRenderer.send('load-url', url),
  goBack: () => ipcRenderer.send('go-back'),
  goForward: () => ipcRenderer.send('go-forward'),
  reload: () => ipcRenderer.send('reload'),
  quitApp: () => ipcRenderer.send('quit'),

    clearSession: () => ipcRenderer.send('clear-session'),

  onTabsUpdated: (cb) =>
    ipcRenderer.on('tabs-updated', (_e, tabs) => cb(tabs)),
  onUrlChanged: (cb) =>
    ipcRenderer.on('url-changed', (_e, url) => cb(url))
});
