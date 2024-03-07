require("plugins")
require("settings")

local themeStatus, catppuccin = pcall(require, "catppuccin")

if themeStatus then
	vim.cmd("colorscheme catppuccin")
else
	return
end
