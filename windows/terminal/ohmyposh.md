# Setup Oh My Posh
Install using winget
```bash
winget install JanDeDobbeleer.OhMyPosh -s winget
```
Insall the font, I am currently using `Meslo LGM NF`

```bash
oh-my-posh font install
```

Update font in windows terminal

## Setup in terminal
Add the following to you `.profile` (or `.bash_profile`) file for bash
```bash
eval "$(oh-my-posh init bash --config PATH_TO_FOLDER/dotfiles/terminal/config.omp.json)"
```

Add the following line to your pwsh config. (You can find this path using `$PROFILE` in pwsh)
```bash
oh-my-posh init pwsh --config 'PATH_TO_FOLDER\dotfiles\terminal\config.omp.json' | Invoke-Expression
```
