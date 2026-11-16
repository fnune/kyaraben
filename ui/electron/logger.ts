/**
 * Structured logging for the Electron main process.
 *
 * Writes JSON lines to <kyarabenStateDir>/electron.log so that main-process
 * events (window lifecycle, IPC, daemon stdout parsing, uncaught errors)
 * are captured alongside the daemon's own kyaraben.log, making production
 * issues easier to debug.
 *
 * The state directory is resolved lazily (on first write) rather than at
 * module load time, because XDG_STATE_HOME may not be set until after the
 * Electron module graph has been evaluated.
 */
import * as fs from "node:fs";
import * as os from "node:os";
import * as path from "node:path";

export type LogLevel = "debug" | "info" | "warn" | "error";

export type LogFields = Record<string, unknown>;

let stream: fs.WriteStream | null = null;
let logPath: string | null = null;
let disabled = false;

/**
 * Resolve the Kyaraben state directory for the current instance.
 *
 * Resolved lazily (at log time, not at module load time) because
 * XDG_STATE_HOME may only be populated after the Electron module graph has
 * been evaluated. Mirrors the daemon's paths.NewPaths() resolution.
 */
function resolveStateDir(): string {
  const xdgStateHome = process.env.XDG_STATE_HOME;
  const base =
    xdgStateHome && xdgStateHome.length > 0
      ? xdgStateHome
      : path.join(os.homedir(), ".local", "state");
  return path.join(base, "kyaraben");
}

function openStream(): fs.WriteStream | null {
  if (stream) {
    return stream;
  }
  if (disabled) {
    return null;
  }
  try {
    const dir = resolveStateDir();
    fs.mkdirSync(dir, { recursive: true });
    const file = path.join(dir, "electron.log");
    const s = fs.createWriteStream(file, { flags: "a" });
    s.on("error", () => {
      stream = null;
      logPath = null;
      disabled = true;
    });
    stream = s;
    logPath = file;
    return s;
  } catch {
    disabled = true;
    return null;
  }
}

function write(level: LogLevel, message: string, fields?: LogFields): void {
  const entry: Record<string, unknown> = {
    time: new Date().toISOString(),
    level,
    pid: process.pid,
    message,
  };
  if (fields) {
    Object.assign(entry, fields);
  }
  const line = JSON.stringify(entry) + "\n";
  const s = openStream();
  if (s) {
    s.write(line);
  }
}

export function debug(message: string, fields?: LogFields): void {
  write("debug", message, fields);
}

export function info(message: string, fields?: LogFields): void {
  write("info", message, fields);
}

export function warn(message: string, fields?: LogFields): void {
  write("warn", message, fields);
}

export function error(message: string, fields?: LogFields): void {
  write("error", message, fields);
}

/** Return the absolute path of the log file, or null if it could not be opened. */
export function logFile(): string | null {
  openStream();
  return logPath;
}

/** Flush and close the log stream. Call from the app "will-quit" handler. */
export function close(): void {
  if (stream) {
    const s = stream;
    stream = null;
    logPath = null;
    s.end();
  }
}
