"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const electron_1 = require("electron");
// Expose IPC methods to the renderer process
electron_1.contextBridge.exposeInMainWorld('electron', {
    invoke: (channel, ...args) => {
        // Whitelist of allowed channels
        const validChannels = [
            'get_systems',
            'get_config',
            'set_config',
            'status',
            'doctor',
            'apply',
            'get_install_status',
            'install_app',
            'uninstall_app',
        ];
        if (validChannels.includes(channel)) {
            return electron_1.ipcRenderer.invoke(channel, ...args);
        }
        throw new Error(`Invalid IPC channel: ${channel}`);
    },
});
