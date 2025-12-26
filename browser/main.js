const { app, BrowserWindow, BrowserView, ipcMain } = require('electron');
const path = require('path');

let mainWindow = null;
let tabs = [];
let activeTabIndex = -1;
let currentBounds = { x: 0, y: 90, width: 1200, height: 800 };

function createWindow() {
  mainWindow = new BrowserWindow({
    width: 1200, height: 900,
    webPreferences: {
      preload: path.join(__dirname, 'preload.js'),
      contextIsolation: true
    }
  });
  mainWindow.loadFile('index.html');
  // Laisse le temps au DOM de charger avant de créer le premier onglet
  setTimeout(() => createTab('https://web.whatsapp.com'), 500);
}

function createTab(url) {
  const view = new BrowserView({
    webPreferences: { contextIsolation: true, partition: 'persist:mainSession' }
  });
  view.webContents.setUserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36");
  view.webContents.loadURL(url);

  const tab = { view, url, title: 'Chargement...' };
  tabs.push(tab);
  
  view.webContents.on('page-title-updated', (e, title) => {
    tab.title = title;
    sendTabs();
  });

  switchTab(tabs.length - 1);
}

function switchTab(index) {
  if (!tabs[index]) return;
  if (activeTabIndex !== -1) mainWindow.removeBrowserView(tabs[activeTabIndex].view);
  activeTabIndex = index;
  mainWindow.addBrowserView(tabs[activeTabIndex].view);
  tabs[activeTabIndex].view.setBounds(currentBounds);
  sendTabs();
}

// Commande de redimensionnement
ipcMain.on('set-bounds', (e, bounds) => {
  currentBounds = bounds;
  if (tabs[activeTabIndex]) {
    tabs[activeTabIndex].view.setBounds(bounds);
    // On force le scroll vers l'élément actif (input) après le redimensionnement
    forceScroll();
  }
});

function forceScroll() {
  const wc = tabs[activeTabIndex]?.view.webContents;
  if (!wc) return;
  
  // Script injecté dans la page web :
  // On cherche l'élément qui a le focus et on le centre dans la nouvelle vue
  const script = `
    if (document.activeElement) {
      document.activeElement.scrollIntoView({ block: "center", behavior: "smooth" });
    }
  `;
  // Petit délai pour laisser la page web s'adapter aux nouvelles dimensions
  setTimeout(() => {
    wc.executeJavaScript(script).catch(() => {});
  }, 200);
}

ipcMain.on('type-key', (e, key) => {
  const wc = tabs[activeTabIndex]?.view.webContents;
  if (!wc) return;
  if (key === 'Backspace') wc.sendInputEvent({ type: 'keyDown', keyCode: 'Backspace' });
  else if (key === 'Enter') wc.sendInputEvent({ type: 'keyDown', keyCode: 'Enter' });
  else wc.sendInputEvent({ type: 'char', keyCode: key });
  
  // Optionnel : scroller à chaque touche pressée pour être sûr
  forceScroll();
});

ipcMain.on('new-tab', () => createTab('https://www.google.com'));
ipcMain.on('switch-tab', (e, i) => switchTab(i));
ipcMain.on('close-tab', (e, i) => {
  if (tabs.length > 1) {
    const removed = tabs.splice(i, 1);
    removed[0].view.webContents.destroy();
    activeTabIndex = Math.min(activeTabIndex, tabs.length - 1);
    switchTab(activeTabIndex);
  }
});
ipcMain.on('load-url', (e, url) => tabs[activeTabIndex]?.view.webContents.loadURL(url));
ipcMain.on('go-back', () => tabs[activeTabIndex]?.view.webContents.goBack());
ipcMain.on('go-forward', () => tabs[activeTabIndex]?.view.webContents.goForward());
ipcMain.on('reload', () => tabs[activeTabIndex]?.view.webContents.reload());
ipcMain.on('quit', () => app.quit());

function sendTabs() {
  if (mainWindow) {
    mainWindow.webContents.send('tabs-updated', tabs.map((t, i) => ({
      index: i, active: i === activeTabIndex, title: t.title
    })));
  }
}

app.whenReady().then(createWindow);
