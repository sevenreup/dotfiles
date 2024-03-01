local status, catppuccin = pcall(require, "catppuccin")

if not status then
	return
end

catppuccin.setup({
    flavour = "mocha",
    transparent_background = false,
    integrations = {
        telescope = {
            enabled = true,
            -- style = "nvchad"
        },
        treesitter = true,
        neotree = true,
    },
})