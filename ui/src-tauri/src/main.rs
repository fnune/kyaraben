// Prevents additional console window on Windows in release
#![cfg_attr(not(debug_assertions), windows_subsystem = "windows")]

use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::process::Stdio;
use tokio::io::{AsyncBufReadExt, AsyncWriteExt, BufReader};
use tokio::process::{Child, Command};
use tokio::sync::Mutex;

// Global daemon process
static DAEMON: once_cell::sync::Lazy<Mutex<Option<DaemonHandle>>> =
    once_cell::sync::Lazy::new(|| Mutex::new(None));

struct DaemonHandle {
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

async fn ensure_daemon() -> Result<(), String> {
    let mut daemon = DAEMON.lock().await;
    if daemon.is_some() {
        return Ok(());
    }

    // Find kyaraben binary - look in common locations
    let binary = find_kyaraben_binary()?;

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

fn find_kyaraben_binary() -> Result<String, String> {
    // Check common locations
    let locations = vec![
        // Same directory as the UI app
        std::env::current_exe()
            .ok()
            .and_then(|p| p.parent().map(|p| p.join("kyaraben")))
            .map(|p| p.to_string_lossy().to_string()),
        // In PATH
        Some("kyaraben".to_string()),
        // Development location
        Some("../kyaraben".to_string()),
    ];

    for loc in locations.into_iter().flatten() {
        if std::path::Path::new(&loc).exists() || which::which(&loc).is_ok() {
            return Ok(loc);
        }
    }

    Err("Could not find kyaraben binary".to_string())
}

async fn send_command(cmd: DaemonCommand) -> Result<DaemonEvent, String> {
    ensure_daemon().await?;

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
        let error_msg = event.data.get("error").and_then(|v| v.as_str()).unwrap_or("Unknown error");
        return Err(error_msg.to_string());
    }

    Ok(event)
}

#[tauri::command]
async fn get_systems() -> Result<Vec<System>, String> {
    let event = send_command(DaemonCommand {
        cmd_type: "get_systems".to_string(),
        data: None,
    })
    .await?;

    serde_json::from_value(event.data).map_err(|e| format!("Failed to parse systems: {}", e))
}

#[tauri::command]
async fn get_config() -> Result<Config, String> {
    let event = send_command(DaemonCommand {
        cmd_type: "get_config".to_string(),
        data: None,
    })
    .await?;

    serde_json::from_value(event.data).map_err(|e| format!("Failed to parse config: {}", e))
}

#[tauri::command]
async fn set_config(user_store: String, systems: HashMap<String, String>) -> Result<(), String> {
    let data = serde_json::json!({
        "userStore": user_store,
        "systems": systems,
    });

    send_command(DaemonCommand {
        cmd_type: "set_config".to_string(),
        data: Some(data),
    })
    .await?;

    Ok(())
}

#[tauri::command]
async fn status() -> Result<Status, String> {
    let event = send_command(DaemonCommand {
        cmd_type: "status".to_string(),
        data: None,
    })
    .await?;

    serde_json::from_value(event.data).map_err(|e| format!("Failed to parse status: {}", e))
}

#[tauri::command]
async fn doctor() -> Result<HashMap<String, Vec<ProvisionResult>>, String> {
    let event = send_command(DaemonCommand {
        cmd_type: "doctor".to_string(),
        data: None,
    })
    .await?;

    serde_json::from_value(event.data).map_err(|e| format!("Failed to parse doctor: {}", e))
}

#[tauri::command]
async fn apply() -> Result<Vec<String>, String> {
    let mut messages = Vec::new();

    // Send apply command and collect progress events
    ensure_daemon().await?;

    let mut daemon = DAEMON.lock().await;
    let daemon_ref = daemon.as_mut().ok_or("Daemon not running")?;

    let cmd = DaemonCommand {
        cmd_type: "apply".to_string(),
        data: None,
    };
    let json = serde_json::to_string(&cmd).map_err(|e| format!("Failed to serialize: {}", e))?;

    daemon_ref.stdin.write_all(json.as_bytes()).await.map_err(|e| format!("Failed to write: {}", e))?;
    daemon_ref.stdin.write_all(b"\n").await.map_err(|e| format!("Failed to write: {}", e))?;
    daemon_ref.stdin.flush().await.map_err(|e| format!("Failed to flush: {}", e))?;

    // Read events until we get a result or error
    loop {
        let mut line = String::new();
        daemon_ref.stdout.read_line(&mut line).await.map_err(|e| format!("Failed to read: {}", e))?;

        let event: DaemonEvent = serde_json::from_str(&line).map_err(|e| format!("Invalid event: {}", e))?;

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
                let error_msg = event.data.get("error").and_then(|v| v.as_str()).unwrap_or("Unknown error");
                return Err(error_msg.to_string());
            }
            _ => {}
        }
    }

    Ok(messages)
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
        ])
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
