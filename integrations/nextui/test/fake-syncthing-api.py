#!/usr/bin/env python3
"""Fake Syncthing API for testing the Kyaraben pak."""

import json
import sys
from http.server import HTTPServer, BaseHTTPRequestHandler

# In-memory state
state = {
    "folders": [],
    "devices": [],
    "api_key": "test-api-key-12345"
}

# Track requests for validation
requests_log = []


class FakeSyncthingHandler(BaseHTTPRequestHandler):
    def log_message(self, format, *args):
        # Log to stderr for debugging
        print(f"[API] {args[0]}", file=sys.stderr)

    def check_api_key(self):
        key = self.headers.get("X-API-Key")
        if key != state["api_key"]:
            self.send_error(401, "Invalid API key")
            return False
        return True

    def send_json(self, data, status=200):
        body = json.dumps(data).encode()
        self.send_response(status)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", len(body))
        self.end_headers()
        self.wfile.write(body)

    def read_body(self):
        length = int(self.headers.get("Content-Length", 0))
        if length:
            return json.loads(self.rfile.read(length))
        return None

    def do_GET(self):
        requests_log.append(("GET", self.path))

        if self.path == "/rest/system/status":
            self.send_json({"myID": "FAKE-DEVICE-ID"})

        elif self.path == "/rest/config/folders":
            if not self.check_api_key():
                return
            self.send_json(state["folders"])

        elif self.path.startswith("/rest/config/devices/"):
            if not self.check_api_key():
                return
            device_id = self.path.split("/")[-1]
            for dev in state["devices"]:
                if dev.get("deviceID") == device_id:
                    self.send_json(dev)
                    return
            # Return default for local device
            if device_id == "FAKE-DEVICE-ID":
                self.send_json({"deviceID": device_id, "name": ""})
            else:
                self.send_error(404)

        else:
            self.send_error(404)

    def do_PUT(self):
        requests_log.append(("PUT", self.path))

        if not self.check_api_key():
            return

        body = self.read_body()

        if self.path == "/rest/config/folders":
            # Validate folder structure
            if not isinstance(body, list):
                self.send_error(400, "Expected array of folders")
                return

            for folder in body:
                if not folder.get("id", "").startswith("kyaraben-"):
                    continue
                # Validate required fields for kyaraben folders
                required = ["id", "path", "type"]
                missing = [f for f in required if f not in folder]
                if missing:
                    self.send_error(400, f"Missing fields: {missing}")
                    return
                if folder["type"] != "sendreceive":
                    self.send_error(400, f"Expected sendreceive type")
                    return

            state["folders"] = body
            self.send_json({"status": "ok"})

        elif self.path.startswith("/rest/config/devices/"):
            device_id = self.path.split("/")[-1]
            if not body.get("deviceID"):
                self.send_error(400, "Missing deviceID")
                return
            state["devices"].append(body)
            self.send_json({"status": "ok"})

        else:
            self.send_error(404)

    def do_POST(self):
        requests_log.append(("POST", self.path))
        self.send_error(404, "Use PUT for config updates")


def run_server(port=8384):
    server = HTTPServer(("127.0.0.1", port), FakeSyncthingHandler)
    print(f"Fake Syncthing API listening on port {port}", file=sys.stderr)
    server.serve_forever()


def validate_final_state():
    """Called at end of test to validate state."""
    errors = []

    # Check folder count
    kyaraben_folders = [f for f in state["folders"] if f["id"].startswith("kyaraben-")]
    if len(kyaraben_folders) != 42:
        errors.append(f"Expected 42 kyaraben folders, got {len(kyaraben_folders)}")

    # Check folder structure
    expected_ids = []
    systems = ["gb", "gbc", "gba", "nes", "snes", "genesis", "psx",
               "mastersystem", "gamegear", "pcengine", "atari2600",
               "c64", "arcade", "ngp"]
    for sys in systems:
        for content in ["roms", "saves", "bios"]:
            expected_ids.append(f"kyaraben-{content}-{sys}")

    actual_ids = [f["id"] for f in kyaraben_folders]
    missing = set(expected_ids) - set(actual_ids)
    if missing:
        errors.append(f"Missing folder IDs: {missing}")

    # Check paths
    for folder in kyaraben_folders:
        if not folder["path"].startswith("/mnt/SDCARD/"):
            errors.append(f"Invalid path for {folder['id']}: {folder['path']}")

    return errors


if __name__ == "__main__":
    if len(sys.argv) > 1 and sys.argv[1] == "--validate":
        # Print validation results
        errors = validate_final_state()
        if errors:
            print("VALIDATION FAILED:")
            for e in errors:
                print(f"  - {e}")
            sys.exit(1)
        print("VALIDATION PASSED")
        print(f"  Folders: {len(state['folders'])}")
        print(f"  Devices: {len(state['devices'])}")
    else:
        run_server()
