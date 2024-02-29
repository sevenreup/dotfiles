# My dotfiles

This will be the home for all my configurations
## Windows Terminal
For this I am using [Oh My Posh](https://ohmyposh.dev/)
More info about config is [here](/others/terminal/ohmyposh.md)
## Neovim setup
### Features
- [wbthomason/packer](https://github.com/wbthomason/packer.nvim): A use-package inspired plugin manager for Neovim.
- [nvim-neo-tree/neo-tree](https://github.com/nvim-neo-tree/neo-tree.nvim): Neovim plugin to manage the file system and other tree like structures.
- [nvim-tree/nvim-web-devicons](https://github.com/nvim-tree/nvim-web-devicons): lua fork of vim-web-devicons for neovim.
- [nvim-telescope/telescope.nvim](https://github.com/nvim-telescope/telescope.nvim): Highly extendable fuzzy finder over lists.
### Setup
Run the `Neovim symlink` command in the symlink.bat file
```bash
mklink /J %LocalAppData%\nvim %cd%\.config\nvim
```
This will create a symlink between the nvim config folder to the `~AppData/Local/nvim` folder that nvm uses on windows.

and run the command:
```bash
nvim +PackerSync
```

or in Neovim Editor:
```bash
:PackerSync
```
