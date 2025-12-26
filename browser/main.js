const {
  app,
  BrowserWindow,
  BrowserView,
  ipcMain,
  session
} = require('electron');
const path = require('path');

/* ------------------ Utils ------------------ */

function extractUrlFromArgv(argv) {
  for (const arg of argv.slice(1)) {
    if (/^(https?:\/\/|www\.)/i.test(arg)) {
      return arg.startsWith('http') ? arg : `https://${arg}`;
    }
  }
  return null;
}

/* ------------------ State ------------------ */

let mainWindow = null;
let tabs = [];
let activeTabIndex = -1;

const HEADER_HEIGHT = 80;

const DEFAULT_URL = 'https://www.google.com';
let initialUrl = extractUrlFromArgv(process.argv) || DEFAULT_URL;

const CUSTOM_UA =
  "Mozilla/5.0 (Macintosh; Intel Mac OS X 14_0) " +
  "AppleWebKit/537.36 (KHTML, like Gecko) " +
  "Chrome/120.0.0.0 Safari/537.36";

/* ------------------ Window ------------------ */

function createWindow() {
  mainWindow = new BrowserWindow({
    width: 1200,
    height: 800,
    webPreferences: {
      preload: path.join(__dirname, 'preload.js'),
      contextIsolation: true,
      nodeIntegration: false
    }
  });

  mainWindow.loadFile('index.html');

  mainWindow.on('resize', resizeActiveView);

  // Onglet initial
  createTab(initialUrl);
}

/* ------------------ Tabs ------------------ */

function createTab(url) {
  const view = new BrowserView({
    webPreferences: {
      contextIsolation: true,
      partition: 'persist:mainSession'
    }
  });

  view.webContents.setUserAgent(CUSTOM_UA);
  view.webContents.loadURL(url);

  const tab = {
    view,
    url,
    title: 'Nouvel onglet',
    favicon: null
  };

  tabs.push(tab);
  const index = tabs.length - 1;

  // URL
  view.webContents.on('did-navigate', (_e, newUrl) => {
    tab.url = newUrl;
    if (index === activeTabIndex) {
      mainWindow.webContents.send('url-changed', newUrl);
    }
  });

  // Titre
  view.webContents.on('page-title-updated', (_e, title) => {
    tab.title = title || tab.url;
    sendTabs();
  });

  // Favicon
  view.webContents.on('page-favicon-updated', (_e, favicons) => {
    if (favicons && favicons.length > 0) {
      tab.favicon = favicons[0];
      sendTabs();
    }
  });

  switchTab(index);
  sendTabs();
}

function switchTab(index) {
  if (!tabs[index]) return;

  if (activeTabIndex !== -1) {
    mainWindow.removeBrowserView(tabs[activeTabIndex].view);
  }

  activeTabIndex = index;
  mainWindow.setBrowserView(tabs[activeTabIndex].view);
  resizeActiveView();

  mainWindow.webContents.send('url-changed', tabs[activeTabIndex].url);
  sendTabs();
}

function closeTab(index) {
  if (!tabs[index]) return;

  const tab = tabs[index];
  mainWindow.removeBrowserView(tab.view);
  tab.view.webContents.destroy();

  tabs.splice(index, 1);

  if (activeTabIndex >= tabs.length) {
    activeTabIndex = tabs.length - 1;
  }

  if (tabs.length > 0) {
    switchTab(activeTabIndex);
  }

  sendTabs();
}

function resizeActiveView() {
  if (!mainWindow || activeTabIndex === -1) return;

  const bounds = mainWindow.getContentBounds();

  tabs[activeTabIndex].view.setBounds({
    x: 0,
    y: HEADER_HEIGHT,
    width: bounds.width,
    height: bounds.height - HEADER_HEIGHT
  });
}

function sendTabs() {
  if (!mainWindow) return;

  mainWindow.webContents.send(
    'tabs-updated',
    tabs.map((t, i) => ({
      index: i,
      active: i === activeTabIndex,
      title: t.title,
      favicon: t.favicon
    }))
  );
}

/* ------------------ IPC ------------------ */

ipcMain.on('new-tab', () => createTab(DEFAULT_URL));
ipcMain.on('switch-tab', (_e, i) => switchTab(i));
ipcMain.on('close-tab', (_e, i) => closeTab(i));

ipcMain.on('load-url', (_e, url) => {
  tabs[activeTabIndex]?.view.webContents.loadURL(url);
});

ipcMain.on('go-back', () => {
  const wc = tabs[activeTabIndex]?.view.webContents;
  if (wc?.canGoBack()) wc.goBack();
});

ipcMain.on('go-forward', () => {
  const wc = tabs[activeTabIndex]?.view.webContents;
  if (wc?.canGoForward()) wc.goForward();
});

ipcMain.on('reload', () => {
  tabs[activeTabIndex]?.view.webContents.reload();
});

ipcMain.on('clear-session', async () => {
  const ses = session.fromPartition('persist:mainSession');
  await ses.clearStorageData();
  await ses.clearCache();
});

ipcMain.on('quit', () => {
  app.quit(); // Simple et propre
});


/* ------------------ Single Instance ------------------ */

const gotLock = app.requestSingleInstanceLock();

if (!gotLock) {
  app.quit();
} else {
  app.on('second-instance', (_event, argv) => {
    const incomingUrl = extractUrlFromArgv(argv);
    if (incomingUrl) {
      createTab(incomingUrl);
    }

    if (mainWindow) {
      if (mainWindow.isMinimized()) mainWindow.restore();
      mainWindow.focus();
    }
  });
}

/* ------------------ macOS open-url ------------------ */

app.on('open-url', (event, url) => {
  event.preventDefault();

  if (!url.startsWith('http')) return;

  if (mainWindow) {
    createTab(url);
    mainWindow.focus();
  } else {
    initialUrl = url;
  }
});

/* ------------------ App lifecycle ------------------ */

app.whenReady().then(createWindow);

app.on('window-all-closed', () => {
  app.quit();
});
