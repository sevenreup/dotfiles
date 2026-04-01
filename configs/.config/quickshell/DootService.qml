import QtQuick
import Quickshell
import Quickshell.Io

// DootService — connects to the doot Unix socket daemon.
//
// Usage in any QML file:
//   DootService.onBrightnessChanged: percent => { ... }
//   DootService.request("brightness.set", { value: 80 })
//
Singleton {
    id: root

    // ── State ─────────────────────────────────────────────────────────────────

    property bool connected: false

    readonly property string socketPath: {
        const dir = Quickshell.env("XDG_RUNTIME_DIR")
            || ("/run/user/" + Quickshell.env("UID"));
        return dir + "/doot.sock";
    }

    // ── Events (server → client) ──────────────────────────────────────────────

    signal brightnessChanged(real percent)
    signal volumeChanged(real volume, bool muted)
    signal micChanged(real volume, bool muted)

    signal calendarToggled()
    signal calendarAuthComplete(string email, string name)
    signal calendarAuthError(string errorMsg)

    // ── Socket ────────────────────────────────────────────────────────────────

    Socket {
        id: socket
        path: root.socketPath
        connected: true

        onConnectedChanged: {
            root.connected = socket.connected;
            if (socket.connected) {
                console.debug("[DootService] connected to", root.socketPath);
            } else {
                console.debug("[DootService] disconnected — retrying in 2s");
                reconnectTimer.start();
            }
        }

        onError: err => {
            console.warn("[DootService] socket error:", err, "— retrying in 2s");
            root.connected = false;
            reconnectTimer.start();
        }

        parser: SplitParser {
            splitMarker: "\n"
            onRead: data => {
                const line = data.trim();
                if (!line) return;
                try {
                    const msg = JSON.parse(line);
                    if (msg.type === "event") {
                        console.debug("[DootService] ← event ", msg.event, JSON.stringify(msg.data));
                        root._handleEvent(msg);
                    } else if (msg.type === "response") {
                        console.debug("[DootService] ← response id=" + msg.id,
                            msg.error ? "error=" + msg.error : "data=" + JSON.stringify(msg.data));
                        root._handleResponse(msg);
                    }
                } catch(e) {
                    console.warn("[DootService] bad JSON:", line);
                }
            }
        }
    }

    // Reconnect after 2s if disconnected
    Timer {
        id: reconnectTimer
        interval: 2000
        onTriggered: socket.connected = true
    }

    // ── Pending requests ──────────────────────────────────────────────────────

    property var _pending: ({})

    // ── Public API ────────────────────────────────────────────────────────────

    // Send a request to doot. Optional callback(data, error) is called on response.
    function request(method, params, callback) {
        const id = Date.now().toString(36) + Math.random().toString(36).slice(2);
        if (callback) root._pending[id] = callback;
        const msg = JSON.stringify({ type: "request", id, method, data: params || {} });
        socket.write(msg + "\n");
    }

    // ── Internal ──────────────────────────────────────────────────────────────

    function _handleEvent(msg) {
        const d = msg.data || {};
        switch (msg.event) {
        case "brightness":         root.brightnessChanged(d.percent ?? 0);                           break;
        case "volume":             root.volumeChanged(d.volume ?? 0, d.muted ?? false);              break;
        case "mic":                root.micChanged(d.volume ?? 0, d.muted ?? false);                 break;
        case "calendar.toggle":    root.calendarToggled();                                           break;
        case "calendar.authComplete": root.calendarAuthComplete(d.email ?? "", d.name ?? "");       break;
        case "calendar.authError":    root.calendarAuthError(d.error ?? "unknown error");            break;
        }
    }

    function _handleResponse(msg) {
        const cb = root._pending[msg.id];
        if (!cb) return;
        delete root._pending[msg.id];
        cb(msg.data || null, msg.error || null);
    }
}
