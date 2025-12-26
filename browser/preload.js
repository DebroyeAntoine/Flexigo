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
  toggleKeyboard: () => ipcRenderer.send('toggle-keyboard'),
  typeKey: (key) => ipcRenderer.send('type-key', key),
  
  // NOUVEAU : On définit les dimensions
  setBounds: (bounds) => ipcRenderer.send('set-bounds', bounds),

  onTabsUpdated: (cb) => ipcRenderer.on('tabs-updated', (e, tabs) => cb(tabs)),
  onUrlChanged: (cb) => ipcRenderer.on('url-changed', (e, url) => cb(url)),
  onKeyboardToggleRequest: (cb) => ipcRenderer.on('keyboard-toggle-request', (e) => cb())
});
