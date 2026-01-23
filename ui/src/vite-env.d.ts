/// <reference types="vite/client" />

declare global {
  type CommandType =
    | "status"
    | "doctor"
    | "apply"
    | "get_systems"
    | "get_config"
    | "set_config"
    | "get_install_status"
    | "install_app"
    | "uninstall_app";

  interface ElectronAPI {
    invoke<T>(command: CommandType, data?: unknown): Promise<T>;
  }

  interface Window {
    electron: ElectronAPI;
  }
}

declare module "*.module.css" {
  const classes: { [key: string]: string };
  export default classes;
}

export {};
