// Prevents additional console window on Windows in release
#![cfg_attr(not(debug_assertions), windows_subsystem = "windows")]

use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::path::PathBuf;
use std::process::Stdio;
use tauri::Manager;
use tokio::io::{AsyncBufReadExt, AsyncWriteExt, BufReader};
use tokio::process::{Child, Command};
use tokio::sync::Mutex;

// Global daemon process
static DAEMON: once_cell::sync::Lazy<Mutex<Option<DaemonHandle>>> =
    once_cell::sync::Lazy::new(|| Mutex::new(None));

struct DaemonHandle {
    #[allow(dead_code)]
    child: Child,
    stdin: tokio::process::ChildStdin,
    stdout: BufReader<tokio::process::ChildStdout>,
}

#[derive(Debug, Serialize, Deserialize)]
struct DaemonCommand {
    #[serde(rename = "type")]
    cmd_type: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    data: Option<serde_json::Value>,
}

#[derive(Debug, Serialize, Deserialize)]
struct DaemonEvent {
    #[serde(rename = "type")]
    event_type: String,
    #[serde(default)]
    data: serde_json::Value,
}

#[derive(Debug, Serialize, Deserialize)]
struct System {
    id: String,
    name: String,
    description: String,
    emulators: Vec<Emulator>,
}

#[derive(Debug, Serialize, Deserialize)]
struct Emulator {
    id: String,
    name: String,
}

#[derive(Debug, Serialize, Deserialize)]
struct Config {
    #[serde(rename = "userStore")]
    user_store: String,
    systems: HashMap<String, String>,
}

#[derive(Debug, Serialize, Deserialize)]
struct Status {
    #[serde(rename = "userStore")]
    user_store: String,
    #[serde(rename = "enabledSystems")]
    enabled_systems: Vec<String>,
    #[serde(rename = "installedEmulators")]
    installed_emulators: Vec<InstalledEmulator>,
    #[serde(rename = "lastApplied")]
    last_applied: String,
}

#[derive(Debug, Serialize, Deserialize)]
struct InstalledEmulator {
    id: String,
    version: String,
}

#[derive(Debug, Serialize, Deserialize)]
struct ProvisionResult {
    filename: String,
    description: String,
    required: bool,
    status: String,
    #[serde(rename = "foundPath")]
    found_path: String,
}

#[derive(Debug, Serialize, Deserialize)]
struct InstallStatus {
    installed: bool,
    #[serde(rename = "appPath")]
    app_path: Option<String>,
    #[serde(rename = "desktopPath")]
    desktop_path: Option<String>,
}

fn get_install_paths() -> (PathBuf, PathBuf) {
    let home = dirs::home_dir().unwrap_or_else(|| PathBuf::from("."));
    let bin_dir = home.join(".local/bin");
    let apps_dir = home.join(".local/share/applications");

    let app_path = bin_dir.join("kyaraben.AppImage");
    let desktop_path = apps_dir.join("kyaraben.desktop");

    (app_path, desktop_path)
}

fn find_sidecar_path(app: &tauri::AppHandle) -> Result<PathBuf, String> {
    // Get the resource directory where sidecars are placed
    let resource_dir = app
        .path()
        .resource_dir()
        .map_err(|e| format!("Failed to get resource dir: {}", e))?;

    // Construct the sidecar path with target triple
    #[cfg(target_os = "linux")]
    let sidecar_name = if cfg!(target_arch = "x86_64") {
        "kyaraben-x86_64-unknown-linux-gnu"
    } else if cfg!(target_arch = "aarch64") {
        "kyaraben-aarch64-unknown-linux-gnu"
    } else {
        "kyaraben"
    };

    #[cfg(target_os = "macos")]
    let sidecar_name = if cfg!(target_arch = "x86_64") {
        "kyaraben-x86_64-apple-darwin"
    } else if cfg!(target_arch = "aarch64") {
        "kyaraben-aarch64-apple-darwin"
    } else {
        "kyaraben"
    };

    #[cfg(target_os = "windows")]
    let sidecar_name = "kyaraben-x86_64-pc-windows-msvc.exe";

    #[cfg(not(any(target_os = "linux", target_os = "macos", target_os = "windows")))]
    let sidecar_name = "kyaraben";

    let sidecar_path = resource_dir.join("binaries").join(sidecar_name);

    if sidecar_path.exists() {
        Ok(sidecar_path)
    } else {
        // Fallback: try without target triple (dev mode)
        let fallback = resource_dir.join("binaries").join("kyaraben");
        if fallback.exists() {
            Ok(fallback)
        } else {
            // Last resort: check if kyaraben is in PATH
            Ok(PathBuf::from("kyaraben"))
        }
    }
}

async fn ensure_daemon(app: &tauri::AppHandle) -> Result<(), String> {
    let mut daemon = DAEMON.lock().await;
    if daemon.is_some() {
        return Ok(());
    }

    let binary = find_sidecar_path(app)?;

    let mut child = Command::new(&binary)
        .arg("daemon")
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::inherit())
        .spawn()
        .map_err(|e| format!("Failed to start daemon: {}", e))?;

    let stdin = child.stdin.take().ok_or("Failed to get stdin")?;
    let stdout = child.stdout.take().ok_or("Failed to get stdout")?;
    let stdout = BufReader::new(stdout);

    *daemon = Some(DaemonHandle {
        child,
        stdin,
        stdout,
    });

    // Wait for ready event
    let daemon_ref = daemon.as_mut().unwrap();
    let mut line = String::new();
    daemon_ref
        .stdout
        .read_line(&mut line)
        .await
        .map_err(|e| format!("Failed to read from daemon: {}", e))?;

    let event: DaemonEvent =
        serde_json::from_str(&line).map_err(|e| format!("Invalid daemon response: {}", e))?;

    if event.event_type != "ready" {
        return Err(format!("Expected ready event, got: {}", event.event_type));
    }

    Ok(())
}

async fn send_command(app: &tauri::AppHandle, cmd: DaemonCommand) -> Result<DaemonEvent, String> {
    ensure_daemon(app).await?;

    let mut daemon = DAEMON.lock().await;
    let daemon_ref = daemon.as_mut().ok_or("Daemon not running")?;

    let json = serde_json::to_string(&cmd).map_err(|e| format!("Failed to serialize: {}", e))?;

    daemon_ref
        .stdin
        .write_all(json.as_bytes())
        .await
        .map_err(|e| format!("Failed to write to daemon: {}", e))?;
    daemon_ref
        .stdin
        .write_all(b"\n")
        .await
        .map_err(|e| format!("Failed to write newline: {}", e))?;
    daemon_ref
        .stdin
        .flush()
        .await
        .map_err(|e| format!("Failed to flush: {}", e))?;

    let mut line = String::new();
    daemon_ref
        .stdout
        .read_line(&mut line)
        .await
        .map_err(|e| format!("Failed to read response: {}", e))?;

    let event: DaemonEvent =
        serde_json::from_str(&line).map_err(|e| format!("Invalid response: {}", e))?;

    if event.event_type == "error" {
        let error_msg = event
            .data
            .get("error")
            .and_then(|v| v.as_str())
            .unwrap_or("Unknown error");
        return Err(error_msg.to_string());
    }

    Ok(event)
}

#[tauri::command]
async fn get_systems(app: tauri::AppHandle) -> Result<Vec<System>, String> {
    let event = send_command(
        &app,
        DaemonCommand {
            cmd_type: "get_systems".to_string(),
            data: None,
        },
    )
    .await?;

    serde_json::from_value(event.data).map_err(|e| format!("Failed to parse systems: {}", e))
}

#[tauri::command]
async fn get_config(app: tauri::AppHandle) -> Result<Config, String> {
    let event = send_command(
        &app,
        DaemonCommand {
            cmd_type: "get_config".to_string(),
            data: None,
        },
    )
    .await?;

    serde_json::from_value(event.data).map_err(|e| format!("Failed to parse config: {}", e))
}

#[tauri::command]
async fn set_config(
    app: tauri::AppHandle,
    user_store: String,
    systems: HashMap<String, String>,
) -> Result<(), String> {
    let data = serde_json::json!({
        "userStore": user_store,
        "systems": systems,
    });

    send_command(
        &app,
        DaemonCommand {
            cmd_type: "set_config".to_string(),
            data: Some(data),
        },
    )
    .await?;

    Ok(())
}

#[tauri::command]
async fn status(app: tauri::AppHandle) -> Result<Status, String> {
    let event = send_command(
        &app,
        DaemonCommand {
            cmd_type: "status".to_string(),
            data: None,
        },
    )
    .await?;

    serde_json::from_value(event.data).map_err(|e| format!("Failed to parse status: {}", e))
}

#[tauri::command]
async fn doctor(app: tauri::AppHandle) -> Result<HashMap<String, Vec<ProvisionResult>>, String> {
    let event = send_command(
        &app,
        DaemonCommand {
            cmd_type: "doctor".to_string(),
            data: None,
        },
    )
    .await?;

    serde_json::from_value(event.data).map_err(|e| format!("Failed to parse doctor: {}", e))
}

#[tauri::command]
async fn apply(app: tauri::AppHandle) -> Result<Vec<String>, String> {
    let mut messages = Vec::new();

    ensure_daemon(&app).await?;

    let mut daemon = DAEMON.lock().await;
    let daemon_ref = daemon.as_mut().ok_or("Daemon not running")?;

    let cmd = DaemonCommand {
        cmd_type: "apply".to_string(),
        data: None,
    };
    let json = serde_json::to_string(&cmd).map_err(|e| format!("Failed to serialize: {}", e))?;

    daemon_ref
        .stdin
        .write_all(json.as_bytes())
        .await
        .map_err(|e| format!("Failed to write: {}", e))?;
    daemon_ref
        .stdin
        .write_all(b"\n")
        .await
        .map_err(|e| format!("Failed to write: {}", e))?;
    daemon_ref
        .stdin
        .flush()
        .await
        .map_err(|e| format!("Failed to flush: {}", e))?;

    // Read events until we get a result or error
    loop {
        let mut line = String::new();
        daemon_ref
            .stdout
            .read_line(&mut line)
            .await
            .map_err(|e| format!("Failed to read: {}", e))?;

        let event: DaemonEvent =
            serde_json::from_str(&line).map_err(|e| format!("Invalid event: {}", e))?;

        match event.event_type.as_str() {
            "progress" => {
                if let Some(msg) = event.data.get("message").and_then(|v| v.as_str()) {
                    messages.push(msg.to_string());
                }
            }
            "result" => {
                messages.push("Apply completed successfully".to_string());
                break;
            }
            "error" => {
                let error_msg = event
                    .data
                    .get("error")
                    .and_then(|v| v.as_str())
                    .unwrap_or("Unknown error");
                return Err(error_msg.to_string());
            }
            _ => {}
        }
    }

    Ok(messages)
}

#[tauri::command]
async fn get_install_status() -> Result<InstallStatus, String> {
    let (app_path, desktop_path) = get_install_paths();

    let installed = app_path.exists() && desktop_path.exists();

    Ok(InstallStatus {
        installed,
        app_path: if app_path.exists() {
            Some(app_path.to_string_lossy().to_string())
        } else {
            None
        },
        desktop_path: if desktop_path.exists() {
            Some(desktop_path.to_string_lossy().to_string())
        } else {
            None
        },
    })
}

#[tauri::command]
async fn install_app() -> Result<(), String> {
    let (app_path, desktop_path) = get_install_paths();

    // Create directories
    if let Some(parent) = app_path.parent() {
        tokio::fs::create_dir_all(parent)
            .await
            .map_err(|e| format!("Failed to create bin directory: {}", e))?;
    }
    if let Some(parent) = desktop_path.parent() {
        tokio::fs::create_dir_all(parent)
            .await
            .map_err(|e| format!("Failed to create applications directory: {}", e))?;
    }

    // Get current executable path
    let current_exe =
        std::env::current_exe().map_err(|e| format!("Failed to get current executable: {}", e))?;

    // Copy AppImage to ~/.local/bin/
    tokio::fs::copy(&current_exe, &app_path)
        .await
        .map_err(|e| format!("Failed to copy AppImage: {}", e))?;

    // Make it executable
    #[cfg(unix)]
    {
        use std::os::unix::fs::PermissionsExt;
        let perms = std::fs::Permissions::from_mode(0o755);
        tokio::fs::set_permissions(&app_path, perms)
            .await
            .map_err(|e| format!("Failed to set permissions: {}", e))?;
    }

    // Create .desktop file
    let desktop_content = format!(
        r#"[Desktop Entry]
Name=Kyaraben
Comment=Declarative emulation manager
Exec={}
Icon=applications-games
Terminal=false
Type=Application
Categories=Game;Emulator;
"#,
        app_path.to_string_lossy()
    );

    tokio::fs::write(&desktop_path, desktop_content)
        .await
        .map_err(|e| format!("Failed to create desktop file: {}", e))?;

    Ok(())
}

#[tauri::command]
async fn uninstall_app() -> Result<(), String> {
    let (app_path, desktop_path) = get_install_paths();

    // Remove files (ignore errors if they don't exist)
    let _ = tokio::fs::remove_file(&app_path).await;
    let _ = tokio::fs::remove_file(&desktop_path).await;

    Ok(())
}

fn main() {
    tauri::Builder::default()
        .invoke_handler(tauri::generate_handler![
            get_systems,
            get_config,
            set_config,
            status,
            doctor,
            apply,
            get_install_status,
            install_app,
            uninstall_app,
        ])
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
