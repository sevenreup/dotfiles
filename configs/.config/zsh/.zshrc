# Shared, cross-platform zsh config — windows & macOS.
# Keep machine-specific env/paths (SDKs, package managers, etc.) out of here;
# they belong in ~/.zshrc.local, which is not tracked by this repo.

export EDITOR=code

# Enable Powerlevel10k instant prompt. Must stay close to the top: nothing
# above this point should print to the console.
if [[ -r "${XDG_CACHE_HOME:-$HOME/.cache}/p10k-instant-prompt-${(%):-%n}.zsh" ]]; then
  source "${XDG_CACHE_HOME:-$HOME/.cache}/p10k-instant-prompt-${(%):-%n}.zsh"
fi

# Deno completions, if installed
[[ -d "$HOME/.zsh/completions" ]] && FPATH="$HOME/.zsh/completions:$FPATH"

# Poor-man's plugin manager: git-clone plugins on first run so windows (Git
# Bash) and macOS get the same look/feel without needing a package manager.
ZSH_PLUGIN_DIR="$HOME/.config/zsh/plugins"

_zsh_plugin_load() {
  local repo="$2" entry="$3" dir="$ZSH_PLUGIN_DIR/$1"
  if [[ ! -d "$dir" ]]; then
    mkdir -p "$ZSH_PLUGIN_DIR"
    git clone --quiet --depth=1 "$repo" "$dir" &>/dev/null
  fi
  [[ -f "$dir/$entry" ]] && source "$dir/$entry"
}

_zsh_plugin_load powerlevel10k https://github.com/romkatv/powerlevel10k.git powerlevel10k.zsh-theme
_zsh_plugin_load zsh-autosuggestions https://github.com/zsh-users/zsh-autosuggestions.git zsh-autosuggestions.zsh
_zsh_plugin_load zsh-history-substring-search https://github.com/zsh-users/zsh-history-substring-search.git zsh-history-substring-search.zsh

if (( $+widgets[history-substring-search-up] )); then
  bindkey '^[[A' history-substring-search-up
  bindkey '^[[B' history-substring-search-down
fi

# Prompt theme — same look on every machine. Run `p10k configure` to change it,
# then copy ~/.config/zsh/p10k.zsh back into this repo.
[[ -f "$HOME/.config/zsh/p10k.zsh" ]] && source "$HOME/.config/zsh/p10k.zsh"

# History
export HISTCONTROL=ignoreboth
export HISTORY_IGNORE="(&|[bf]g|c|clear|history|exit|q|pwd|* --help)"
export HISTSIZE=10000
export SAVEHIST=10000
setopt SHARE_HISTORY

# Completion
autoload -Uz compinit
compinit

# zsh-syntax-highlighting must be sourced last, after compinit and every
# other plugin that wraps zle widgets.
_zsh_plugin_load zsh-syntax-highlighting https://github.com/zsh-users/zsh-syntax-highlighting.git zsh-syntax-highlighting.zsh

unset -f _zsh_plugin_load

# Aliases
alias c="clear"
alias claudeWork="CLAUDE_CONFIG_DIR=~/.claude-work claude"
alias n="ninja"
if command -v nproc &>/dev/null; then
  alias make="make -j$(nproc)"
  alias ninja="ninja -j$(nproc)"
elif command -v sysctl &>/dev/null; then
  alias make="make -j$(sysctl -n hw.ncpu)"
  alias ninja="ninja -j$(sysctl -n hw.ncpu)"
fi
