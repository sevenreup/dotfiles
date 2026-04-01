import QtQuick
import QtQuick.Layouts
import Quickshell
import Quickshell.Io
import Quickshell.Services.Pipewire
import Quickshell.Widgets

Scope {
    id: root

    // ── Volume ───────────────────────────────────────────────────────────────

    PwObjectTracker {
        objects: [Pipewire.defaultAudioSink]
    }

    Connections {
        target: Pipewire.defaultAudioSink?.audio

        function onVolumeChanged() {
            root.showOsd("volume");
        }

        function onMutedChanged() {
            root.showOsd("volume");
        }
    }

    // ── Microphone ───────────────────────────────────────────────────────────

    PwObjectTracker {
        objects: [Pipewire.defaultAudioSource]
    }

    Connections {
        target: Pipewire.defaultAudioSource?.audio

        function onVolumeChanged() {
            root.showOsd("mic");
        }

        function onMutedChanged() {
            root.showOsd("mic");
        }
    }

    // ── Brightness ───────────────────────────────────────────────────────────

    property int brightness: 0
    property int maxBrightness: 1
    property real brightnessPercent: maxBrightness > 0 ? brightness / maxBrightness : 0
    property bool brightnessInitialized: false

    Process {
        id: maxBrightnessProc
        command: ["brightnessctl", "m"]
        stdout: SplitParser {
            onRead: data => {
                let val = parseInt(data.trim());
                if (!isNaN(val) && val > 0)
                    root.maxBrightness = val;
            }
        }
        Component.onCompleted: running = true
    }

    Process {
        id: brightnessProc
        command: ["brightnessctl", "g"]
        stdout: SplitParser {
            onRead: data => {
                let val = parseInt(data.trim());
                if (!isNaN(val)) {
                    if (root.brightnessInitialized && val !== root.brightness) {
                        root.showOsd("brightness");
                    }
                    root.brightness = val;
                    root.brightnessInitialized = true;
                }
            }
        }
    }

    Timer {
        interval: 250
        running: true
        repeat: true
        onTriggered: brightnessProc.running = true
    }

    // ── OSD state ────────────────────────────────────────────────────────────

    property string osdType: "volume"
    property bool shouldShowOsd: false

    function showOsd(type) {
        root.osdType = type;
        root.shouldShowOsd = true;
        hideTimer.restart();
    }

    Timer {
        id: hideTimer
        interval: 1500
        onTriggered: root.shouldShowOsd = false
    }

    // ── Helpers ───────────────────────────────────────────────────────────────

    function osdIcon() {
        switch (root.osdType) {
        case "brightness":
            return Quickshell.iconPath("display-brightness-symbolic");
        case "mic":
            {
                const muted = Pipewire.defaultAudioSource?.audio.muted ?? false;
                return Quickshell.iconPath(muted ? "microphone-sensitivity-muted-symbolic" : "microphone-sensitivity-high-symbolic");
            }
        default:
            { // volume
                const muted = Pipewire.defaultAudioSink?.audio.muted ?? false;
                const vol = Pipewire.defaultAudioSink?.audio.volume ?? 0;
                if (muted)
                    return Quickshell.iconPath("audio-volume-muted-symbolic");
                if (vol < 0.33)
                    return Quickshell.iconPath("audio-volume-low-symbolic");
                if (vol < 0.66)
                    return Quickshell.iconPath("audio-volume-medium-symbolic");
                return Quickshell.iconPath("audio-volume-high-symbolic");
            }
        }
    }

    function osdBarColor() {
        switch (root.osdType) {
        case "brightness":
            return "#f9e2af"; // yellow
        case "mic":
            return (Pipewire.defaultAudioSource?.audio.muted ?? false) ? "#f38ba8"   // red - muted
            : "#a6e3a1";  // green - active
        default:
            // volume
            return (Pipewire.defaultAudioSink?.audio.muted ?? false) ? "#f38ba8"   // red - muted
            : "#cba6f7";  // mauve
        }
    }

    function osdValue() {
        switch (root.osdType) {
        case "brightness":
            return root.brightnessPercent;
        case "mic":
            return Pipewire.defaultAudioSource?.audio.volume ?? 0;
        default:
            return Pipewire.defaultAudioSink?.audio.volume ?? 0;
        }
    }

    function osdLabel() {
        switch (root.osdType) {
        case "mic":
            return (Pipewire.defaultAudioSource?.audio.muted ?? false) ? "Muted" : "Live";
        default:
            return Math.round(osdValue() * 100) + "%";
        }
    }

    // ── Window ───────────────────────────────────────────────────────────────

    LazyLoader {
        active: root.shouldShowOsd

        PanelWindow {
            anchors {
                top: true
                right: true
                bottom: false
                left: false
            }
            margins.top: 60
            margins.right: 20
            exclusiveZone: 0

            implicitWidth: 300
            implicitHeight: 50
            color: "transparent"
            mask: Region {}

            Rectangle {
                anchors.fill: parent
                radius: height / 2
                color: "#cc1e1e2e"

                RowLayout {
                    anchors {
                        fill: parent
                        leftMargin: 12
                        rightMargin: 16
                    }
                    spacing: 10

                    IconImage {
                        implicitSize: 24
                        source: root.osdIcon()
                    }

                    Rectangle {
                        Layout.fillWidth: true
                        implicitHeight: 8
                        radius: 4
                        color: "#40ffffff"

                        Rectangle {
                            anchors {
                                left: parent.left
                                top: parent.top
                                bottom: parent.bottom
                            }
                            radius: parent.radius
                            color: root.osdBarColor()
                            implicitWidth: parent.width * Math.min(1, root.osdValue())
                        }
                    }

                    Text {
                        text: root.osdLabel()
                        color: "#cdd6f4"
                        font.pixelSize: 13
                        font.bold: true
                        Layout.minimumWidth: 38
                        horizontalAlignment: Text.AlignRight
                    }
                }
            }
        }
    }
}
