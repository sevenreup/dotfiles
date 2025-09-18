// NAME: Keyboard Shortcut
// AUTHOR: khanhas
// DESCRIPTION: Register a few more keybinds to support keyboard navigation.

/// <reference path="../globals.d.ts" />

(function keyboardShortcut() {
    const { Keyboard } = Spicetify;

    if (!Keyboard) {
        setTimeout(keyboardShortcut, 200);
        return;
    }

    // Register keyboard shortcuts
    Keyboard.registerShortcut({
        key: Keyboard.KEYS["ARROW_UP"],
        ctrl: true,
        callback: () => {
            Spicetify.Player.setVolume(Math.min(1, Spicetify.Player.getVolume() + 0.1));
        }
    });

    Keyboard.registerShortcut({
        key: Keyboard.KEYS["ARROW_DOWN"],
        ctrl: true,
        callback: () => {
            Spicetify.Player.setVolume(Math.max(0, Spicetify.Player.getVolume() - 0.1));
        }
    });

    Keyboard.registerShortcut({
        key: Keyboard.KEYS.L,
        ctrl: true,
        shift: true,
        callback: () => {
            Spicetify.Player.toggleRepeat();
        }
    });

    Keyboard.registerShortcut({
        key: Keyboard.KEYS.S,
        ctrl: true,
        shift: true,
        callback: () => {
            Spicetify.Player.toggleShuffle();
        }
    });
})();