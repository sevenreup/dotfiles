# Entry point for windows & macOS. Symlinked to ~/.zshrc by mac/setup.sh and
# windows/setup.sh. Machine-specific env/paths go in ~/.zshrc.local instead of
# here, so they never end up in this repo.

source "$HOME/.config/zsh/.zshrc"

[[ -f "$HOME/.zshrc.local" ]] && source "$HOME/.zshrc.local"
