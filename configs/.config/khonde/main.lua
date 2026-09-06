-- A bottom taskbar styled after a retro pixel-terminal desktop: cream
-- background, square borders, burnt-orange highlight for the active
-- window, an outlined badge for the active workspace, monospace
-- everywhere. One PanelWindow per monitor.
--
-- Workspace/window rows use Row+onPress rather than Button — Button carries
-- GTK4's own border/shadow chrome that's hard to fully flatten (see
-- examples/launcher for the same fix), and a plain Row gives full control
-- over the flat, square look this theme wants.
--
-- Run with:
--   go run ./cmd/khoma local/khonde/main.lua

local Clock = component("Clock", function()
    local now, setNow = useState(os.date("%I:%M %p"))
    useInterval(function() setNow(os.date("%I:%M %p")) end, 1000)
    return Text { text = now, class = "clock" }
end)

local WorkspaceRow = component("WorkspaceRow", function()
    local hypr = require("khoma.hyprland")
    local ws = useStore(Workspaces)
    local clients = useStore(Clients)
    local active = useStore(ActiveWindow)

    -- Workspaces/ActiveWindow don't share an id, so the active workspace is
    -- found indirectly: match the focused window's class+title against the
    -- Clients list, then read that client's workspaceId.
    local activeWsId = nil
    for _, c in ipairs(clients) do
        if c.class == active.class and c.title == active.title then
            activeWsId = c.workspaceId
            break
        end
    end

    local row = {}
    for i, w in ipairs(ws) do
        row[i] = Row {
            key = "ws-" .. w.id,
            class = (w.id == activeWsId) and "ws ws-active" or "ws",
            cursor = "pointer",
            onPress = function(button) if button == 1 then hypr.dispatch("workspace " .. w.id) end end,
            Text { text = w.name },
        }
    end
    return Row { spacing = 14, children = row }
end)

local WindowList = component("WindowList", function()
    local hypr = require("khoma.hyprland")
    local clients = useStore(Clients)
    local active = useStore(ActiveWindow)

    local row = {}
    for i, c in ipairs(clients) do
        local isActive = c.class == active.class and c.title == active.title
        row[i] = Row {
            key = "win-" .. c.address,
            width = 150,
            class = isActive and "taskitem taskitem-active" or "taskitem",
            cursor = "pointer",
            onPress = function(button)
                if button == 1 then hypr.dispatch("focuswindow address:" .. c.address) end
            end,
            Text {
                text = (c.title ~= "" and c.title) or c.class,
                halign = "start",
                hexpand = true,
                ellipsize = "end",
            },
        }
    end
    return Row { spacing = 4, children = row }
end)

local VolumeIndicator = component("VolumeIndicator", function()
    local vol = useStore(Volume)
    local icon = vol.muted and "audio-volume-muted-symbolic" or "audio-volume-high-symbolic"
    return Icon { icon = icon, size = 16, class = "tray" }
end)

local BatteryIndicator = component("BatteryIndicator", function()
    local bat = useStore(Battery)
    if not bat.present then
        return nil
    end
    local icon = bat.state == "charging" and "battery-good-charging-symbolic" or "battery-good-symbolic"
    return Icon { icon = icon, size = 16, class = "tray" }
end)

local Bar = component("Bar", function(props)
    return PanelWindow {
        screen = props.screen,
        anchor = { "top", "left", "right" },
        exclusive = true,
        layer = "top",
        keyboard = "none",
        namespace = "khoma-bar",
        height = 26,
        Padding {
            left = 8,
            right = 8,
            Box {
                class = "khoma-container",
                Row {
                    class = "khoma-bar",
                    spacing = 16,
                    valign = "center",
                    WorkspaceRow {},
                    -- WindowList {},
                    Spacer {},
                    VolumeIndicator {},
                    BatteryIndicator {},
                    Clock {},
                },
            },
        },
    }
end)

App = component("App", function()
    local screens = useStore(Screens)
    local bars = {}
    for i, s in ipairs(screens) do
        bars[i] = Bar { key = "bar-" .. s.name, screen = s.name }
    end
    return bars
end)
