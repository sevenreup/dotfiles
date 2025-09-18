// NAME: Shuffle+
// AUTHOR: khanhas
// DESCRIPTION: True shuffle with no bias.

/// <reference path="../globals.d.ts" />

(function shufflePlus() {
    const { Player } = Spicetify;

    if (!Player) {
        setTimeout(shufflePlus, 200);
        return;
    }

    // Override shuffle function
    const originalShuffle = Player.shuffle;

    Player.shuffle = function(tracks) {
        if (!tracks || tracks.length === 0) return tracks;

        // Fisher-Yates shuffle algorithm
        const shuffled = [...tracks];
        for (let i = shuffled.length - 1; i > 0; i--) {
            const j = Math.floor(Math.random() * (i + 1));
            [shuffled[i], shuffled[j]] = [shuffled[j], shuffled[i]];
        }

        return shuffled;
    };
})();