const { webFrame, ipcRenderer } = require('electron');

ipcRenderer.invoke('get-hook-code', window.location.href).then(function(code) {
  webFrame.executeJavaScript(code);
});
