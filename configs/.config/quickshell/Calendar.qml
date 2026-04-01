import QtQuick
import QtQuick.Layouts
import Quickshell
import Quickshell.Hyprland

// Calendar widget — toggled by DootService.calendarToggled()
// Google sign-in: doot calendar setup credentials  →  then sign in via the widget
Scope {
    id: root

    // ── State ─────────────────────────────────────────────────────────────────

    property bool calendarVisible: false
    property int displayYear: new Date().getFullYear()
    property int displayMonth: new Date().getMonth() + 1
    property int selectedDay: new Date().getDate()

    // eventData: "YYYY-MM-DD" → [{title, color, allDay, accountEmail, ...}]
    property var eventData: ({})
    property var accounts: []
    property bool loading: false

    // OAuth state
    property bool authPending: false
    property string authError: ""

    // ── DootService wiring ────────────────────────────────────────────────────

    Connections {
        target: dootService

        function onCalendarToggled() {
            root.calendarVisible = !root.calendarVisible;
            if (root.calendarVisible)
                root.loadAccounts();
        }

        function onCalendarAuthComplete(email, name) {
            root.authPending = false;
            root.authError = "";
            root.loadAccounts();
        }

        function onCalendarAuthError(errorMsg) {
            root.authPending = false;
            root.authError = errorMsg;
        }
    }

    // ── Public functions ──────────────────────────────────────────────────────

    function loadAccounts() {
        dootService.request("calendar.accounts", {}, function (data, err) {
            if (!err && data) {
                root.accounts = data;
                if (data.length > 0)
                    root.loadEvents();
            }
        });
    }

    function loadEvents() {
        root.loading = true;
        dootService.request("calendar.events", {
            year: root.displayYear,
            month: root.displayMonth
        }, function (data, err) {
            root.loading = false;
            if (!err && data) {
                let map = {};
                for (let ev of data) {
                    if (ev.allDay && ev.end) {
                        let cur = new Date(ev.start + "T00:00:00");
                        let end = new Date(ev.end + "T00:00:00");
                        while (cur < end) {
                            let key = cur.getFullYear() + "-" + String(cur.getMonth() + 1).padStart(2, "0") + "-" + String(cur.getDate()).padStart(2, "0");
                            if (!map[key]) map[key] = [];
                            map[key].push(ev);
                            cur.setDate(cur.getDate() + 1);
                        }
                    } else {
                        let key = ev.start.substring(0, 10);
                        if (!map[key]) map[key] = [];
                        map[key].push(ev);
                    }
                }
                root.eventData = map;
            }
        });
    }

    function startAuth() {
        root.authPending = true;
        root.authError = "";
        dootService.request("calendar.auth.start", {}, function (data, err) {
            if (err) {
                root.authPending = false;
                root.authError = err;
            }
            // On success, the browser opens; result arrives via authComplete/authError events
        });
    }

    function removeAccount(username) {
        dootService.request("calendar.auth.remove", { username: username }, function (data, err) {
            if (!err) root.loadAccounts();
        });
    }

    function prevMonth() {
        if (root.displayMonth === 1) { root.displayMonth = 12; root.displayYear--; }
        else root.displayMonth--;
        root.loadEvents();
    }

    function nextMonth() {
        if (root.displayMonth === 12) { root.displayMonth = 1; root.displayYear++; }
        else root.displayMonth++;
        root.loadEvents();
    }

    // ── Computed: calendar grid cells (42 = 6 rows × 7 cols) ─────────────────

    readonly property var calendarCells: {
        let y = root.displayYear;
        let m = root.displayMonth;
        let firstDow = new Date(y, m - 1, 1).getDay();
        let daysInMonth = new Date(y, m, 0).getDate();
        let prevDays = new Date(y, m - 1, 0).getDate();
        let td = new Date();
        let todayY = td.getFullYear(), todayM = td.getMonth() + 1, todayD = td.getDate();

        let cells = [];
        for (let i = 0; i < 42; i++) {
            let day = i - firstDow + 1;
            let inCurrent = day >= 1 && day <= daysInMonth;
            let aY = y, aM = m, aD = day;

            if (day < 1) {
                aD = prevDays + day;
                aM = m - 1;
                if (aM < 1) { aM = 12; aY--; }
            } else if (day > daysInMonth) {
                aD = day - daysInMonth;
                aM = m + 1;
                if (aM > 12) { aM = 1; aY++; }
            }

            let key = aY + "-" + String(aM).padStart(2, "0") + "-" + String(aD).padStart(2, "0");
            cells.push({
                day: aD, month: aM, year: aY, dateKey: key,
                isCurrent: inCurrent,
                isToday: aY === todayY && aM === todayM && aD === todayD,
                isSelected: inCurrent && aD === root.selectedDay && aY === y && aM === m,
                events: root.eventData[key] || []
            });
        }
        return cells;
    }

    readonly property string monthLabel: {
        const names = ["January", "February", "March", "April", "May", "June",
                        "July", "August", "September", "October", "November", "December"];
        return (names[root.displayMonth - 1] || "") + "  " + root.displayYear;
    }

    // ── Window ────────────────────────────────────────────────────────────────

    LazyLoader {
        active: root.calendarVisible

        PanelWindow {
            id: calendarWindow
            anchors { top: true; right: true }
            margins.top: 54
            margins.right: 12
            exclusiveZone: 0
            implicitWidth: 360
            implicitHeight: root.accounts.length === 0 ? 260 : 520
            color: "transparent"

            HyprlandFocusGrab {
                active: root.calendarVisible
                windows: [calendarWindow]
                onCleared: root.calendarVisible = false
            }

            Rectangle {
                anchors.fill: parent
                radius: 14
                color: "#ee1e1e2e"
                border.color: "#40cba6f7"
                border.width: 1

                Loader {
                    anchors { fill: parent; margins: 12 }
                    sourceComponent: root.accounts.length === 0 ? emptyState : calendarView
                }
            }
        }
    }

    // ── Empty state: sign in with Google ─────────────────────────────────────

    Component {
        id: emptyState

        Column {
            spacing: 14
            anchors.centerIn: parent

            Text {
                anchors.horizontalCenter: parent.horizontalCenter
                text: "󰸗"
                font.pixelSize: 36
                color: "#45475a"
            }

            Text {
                anchors.horizontalCenter: parent.horizontalCenter
                text: "No calendars connected"
                font.pixelSize: 14
                font.weight: Font.Medium
                color: "#cdd6f4"
            }

            // Error message (auth failure or missing credentials)
            Text {
                anchors.horizontalCenter: parent.horizontalCenter
                visible: root.authError !== ""
                text: root.authError
                font.pixelSize: 11
                color: "#f38ba8"
                wrapMode: Text.WordWrap
                width: 300
                horizontalAlignment: Text.AlignHCenter
            }

            // Sign in button
            Rectangle {
                anchors.horizontalCenter: parent.horizontalCenter
                width: 200
                height: 36
                radius: 8
                color: {
                    if (root.authPending) return "#313244";
                    return signInArea.containsMouse ? "#cba6f7" : "#313244";
                }
                border.color: root.authPending ? "#45475a" : signInArea.containsMouse ? "#cba6f7" : "#45475a"
                border.width: 1

                Row {
                    anchors.centerIn: parent
                    spacing: 8

                    Text {
                        anchors.verticalCenter: parent.verticalCenter
                        text: root.authPending ? "󰑓" : "󰊫"
                        font.pixelSize: 14
                        color: root.authPending ? "#6c7086" : signInArea.containsMouse ? "#1e1e2e" : "#cdd6f4"

                        RotationAnimator on rotation {
                            running: root.authPending
                            from: 0; to: 360
                            duration: 1200
                            loops: Animation.Infinite
                        }
                    }

                    Text {
                        anchors.verticalCenter: parent.verticalCenter
                        text: root.authPending ? "Opening browser…" : "Sign in with Google"
                        font.pixelSize: 12
                        font.weight: Font.Medium
                        color: root.authPending ? "#6c7086" : signInArea.containsMouse ? "#1e1e2e" : "#cdd6f4"
                    }
                }

                MouseArea {
                    id: signInArea
                    anchors.fill: parent
                    hoverEnabled: true
                    cursorShape: Qt.PointingHandCursor
                    enabled: !root.authPending
                    onClicked: root.startAuth()
                }
            }

            Text {
                anchors.horizontalCenter: parent.horizontalCenter
                text: "Requires: doot calendar setup credentials"
                font.pixelSize: 10
                color: "#45475a"
            }
        }
    }

    // ── Calendar view ─────────────────────────────────────────────────────────

    Component {
        id: calendarView

        Column {
            spacing: 0

            // ── Month navigation header ───────────────────────────────────────

            RowLayout {
                width: parent.width
                height: 44
                spacing: 0

                Rectangle {
                    width: 32; height: 32; radius: 6
                    color: prevArea.containsMouse ? "#313244" : "transparent"
                    Text { anchors.centerIn: parent; text: "‹"; font.pixelSize: 18; color: "#a6adc8" }
                    MouseArea {
                        id: prevArea
                        anchors.fill: parent
                        hoverEnabled: true
                        cursorShape: Qt.PointingHandCursor
                        onClicked: root.prevMonth()
                    }
                }

                Text {
                    Layout.fillWidth: true
                    text: root.monthLabel
                    font.pixelSize: 14
                    font.weight: Font.Medium
                    color: "#cdd6f4"
                    horizontalAlignment: Text.AlignHCenter
                }

                Rectangle {
                    width: 32; height: 32; radius: 6
                    color: nextArea.containsMouse ? "#313244" : "transparent"
                    Text { anchors.centerIn: parent; text: "›"; font.pixelSize: 18; color: "#a6adc8" }
                    MouseArea {
                        id: nextArea
                        anchors.fill: parent
                        hoverEnabled: true
                        cursorShape: Qt.PointingHandCursor
                        onClicked: root.nextMonth()
                    }
                }
            }

            // ── Day-of-week labels ────────────────────────────────────────────

            Row {
                width: parent.width
                height: 22
                spacing: 0
                Repeater {
                    model: ["Su", "Mo", "Tu", "We", "Th", "Fr", "Sa"]
                    Text {
                        width: parent.width / 7
                        text: modelData
                        font.pixelSize: 10
                        color: "#6c7086"
                        horizontalAlignment: Text.AlignHCenter
                    }
                }
            }

            Rectangle { width: parent.width; height: 1; color: "#313244"; opacity: 0.6 }

            // ── Calendar grid ─────────────────────────────────────────────────

            Item {
                width: parent.width
                height: calGrid.height
                opacity: root.loading ? 0.4 : 1.0
                Behavior on opacity { NumberAnimation { duration: 120 } }

                GridLayout {
                    id: calGrid
                    width: parent.width
                    columns: 7
                    columnSpacing: 0
                    rowSpacing: 0

                    Repeater {
                        model: root.calendarCells

                        Rectangle {
                            id: cellRect
                            required property var modelData
                            required property int index

                            Layout.fillWidth: true
                            height: 44
                            radius: 5
                            color: modelData.isSelected && modelData.isCurrent ? "#40cba6f7" : modelData.isToday ? "#25cba6f7" : "transparent"

                            MouseArea {
                                anchors.fill: parent
                                enabled: modelData.isCurrent
                                cursorShape: Qt.PointingHandCursor
                                onClicked: root.selectedDay = cellRect.modelData.day
                            }

                            Column {
                                anchors { top: parent.top; horizontalCenter: parent.horizontalCenter; topMargin: 4 }
                                spacing: 3

                                Text {
                                    anchors.horizontalCenter: parent.horizontalCenter
                                    text: cellRect.modelData.day
                                    font.pixelSize: 12
                                    font.weight: cellRect.modelData.isToday ? Font.Bold : Font.Normal
                                    color: !cellRect.modelData.isCurrent ? "#45475a" : cellRect.modelData.isToday ? "#cba6f7" : "#cdd6f4"
                                }

                                Row {
                                    anchors.horizontalCenter: parent.horizontalCenter
                                    spacing: 2
                                    Repeater {
                                        model: Math.min(cellRect.modelData.events.length, 3)
                                        Rectangle {
                                            required property int index
                                            width: 4; height: 4; radius: 2
                                            color: cellRect.modelData.events[index]?.color ?? "#cba6f7"
                                        }
                                    }
                                }
                            }
                        }
                    }
                }
            }

            // ── Selected-day event list ───────────────────────────────────────

            property string selectedKey: root.displayYear + "-" + String(root.displayMonth).padStart(2, "0") + "-" + String(root.selectedDay).padStart(2, "0")
            property var selectedEvents: root.eventData[selectedKey] || []

            Column {
                width: parent.width
                spacing: 4
                visible: parent.selectedEvents.length > 0

                Item { width: 1; height: 8 }
                Rectangle { width: parent.width; height: 1; color: "#313244"; opacity: 0.6 }
                Item { width: 1; height: 4 }

                Text {
                    leftPadding: 4
                    text: Qt.formatDate(new Date(root.displayYear, root.displayMonth - 1, root.selectedDay), "dddd, MMMM d")
                    font.pixelSize: 11
                    color: "#6c7086"
                }

                Repeater {
                    model: parent.parent.selectedEvents

                    Rectangle {
                        required property var modelData
                        width: parent.width
                        height: evRow.implicitHeight + 8
                        radius: 6
                        color: "#20313244"

                        RowLayout {
                            id: evRow
                            anchors {
                                left: parent.left; right: parent.right; verticalCenter: parent.verticalCenter
                                leftMargin: 8; rightMargin: 8
                            }
                            spacing: 8

                            Rectangle {
                                width: 3; height: 28; radius: 1.5
                                color: modelData.color || "#cba6f7"
                                Layout.alignment: Qt.AlignVCenter
                            }

                            Column {
                                Layout.fillWidth: true
                                spacing: 1
                                Text {
                                    width: parent.width
                                    text: modelData.title || "(no title)"
                                    font.pixelSize: 12
                                    color: "#cdd6f4"
                                    elide: Text.ElideRight
                                }
                                Text {
                                    visible: !modelData.allDay
                                    text: {
                                        try {
                                            let s = new Date(modelData.start);
                                            let e = new Date(modelData.end);
                                            return Qt.formatTime(s, "h:mm AP") + " – " + Qt.formatTime(e, "h:mm AP");
                                        } catch (_) { return ""; }
                                    }
                                    font.pixelSize: 10
                                    color: "#6c7086"
                                }
                                Text {
                                    visible: modelData.allDay === true
                                    text: "All day"
                                    font.pixelSize: 10
                                    color: "#6c7086"
                                }
                            }

                            Text {
                                text: modelData.calendarName || ""
                                font.pixelSize: 10
                                color: "#45475a"
                                elide: Text.ElideRight
                                Layout.maximumWidth: 80
                            }
                        }
                    }
                }

                Item { width: 1; height: 4 }
            }

            // ── Account chips ─────────────────────────────────────────────────

            Rectangle { width: parent.width; height: 1; color: "#313244"; opacity: 0.6 }

            Flow {
                width: parent.width
                spacing: 6
                topPadding: 8
                bottomPadding: 4

                Repeater {
                    model: root.accounts

                    Rectangle {
                        required property var modelData
                        height: 24
                        width: chipLabel.implicitWidth + 28
                        radius: 12
                        color: "#313244"

                        RowLayout {
                            anchors { fill: parent; leftMargin: 8; rightMargin: 8 }
                            spacing: 4

                            Rectangle {
                                width: 14; height: 14; radius: 7; color: "#cba6f7"
                                Text {
                                    anchors.centerIn: parent
                                    text: (modelData.name || modelData.username || "?")[0].toUpperCase()
                                    font.pixelSize: 8; font.weight: Font.Bold; color: "#1e1e2e"
                                }
                            }

                            Text {
                                id: chipLabel
                                text: modelData.name || modelData.username
                                font.pixelSize: 10; color: "#a6adc8"
                                elide: Text.ElideRight
                                Layout.maximumWidth: 160
                            }

                            Text {
                                text: "×"
                                font.pixelSize: 12
                                color: rmArea.containsMouse ? "#f38ba8" : "#45475a"
                                MouseArea {
                                    id: rmArea
                                    anchors.fill: parent
                                    hoverEnabled: true
                                    cursorShape: Qt.PointingHandCursor
                                    onClicked: root.removeAccount(modelData.username)
                                }
                            }
                        }
                    }
                }

                // Add account chip — launches OAuth flow
                Rectangle {
                    height: 24
                    width: addLabel.implicitWidth + 20
                    radius: 12
                    color: addChipArea.containsMouse ? "#3d3d55" : "#252535"
                    opacity: root.authPending ? 0.5 : 1.0

                    Text {
                        id: addLabel
                        anchors.centerIn: parent
                        text: root.authPending ? "Opening browser…" : "+ Add account"
                        font.pixelSize: 10
                        color: addChipArea.containsMouse ? "#cba6f7" : "#6c7086"
                    }
                    MouseArea {
                        id: addChipArea
                        anchors.fill: parent
                        hoverEnabled: true
                        cursorShape: Qt.PointingHandCursor
                        enabled: !root.authPending
                        onClicked: root.startAuth()
                    }
                }
            }
        }
    }
}
