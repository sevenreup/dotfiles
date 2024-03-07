local fn = vim.fn

-- Automatically install packer
local install_path = fn.stdpath("data") .. "/site/pack/packer/start/packer.nvim"
if fn.empty(fn.glob(install_path)) > 0 then
	PACKER_BOOTSTRAP = fn.system({
		"git",
		"clone",
		"--depth",
		"1",
		"https://github.com/wbthomason/packer.nvim",
		install_path,
	})
	print("Installing packer close and reopen Neovim...")
	vim.cmd([[packadd packer.nvim]])
end

-- Autocommand that reloads neovim whenever you save the plugins.lua file
vim.cmd([[
  augroup packer_user_config
    autocmd!
    autocmd BufWritePost plugins.lua source <afile> | PackerSync
  augroup end
]])

-- Use a protected call so we don't error out on first use
local status_ok, packer = pcall(require, "packer")
if not status_ok then
	return
end

-- Have packer use a popup window
packer.init({
	display = {
		open_fn = function()
			return require("packer.util").float({ border = "rounded" })
		end,
	},
})

-- Install your plugins here
return packer.startup(function(use)
	 -- Packer
	 use("wbthomason/packer.nvim")

	 -- Common utilities
	 use("nvim-lua/plenary.nvim")
 
	 -- Icons
	 use("nvim-tree/nvim-web-devicons")

	 -- Treesitter
	 use({
        "nvim-treesitter/nvim-treesitter",
        run = function()
            require("nvim-treesitter.install").update({ with_sync = true })
        end,
        config = function()
            require("configs.treesitter")
        end,
    })

	  -- Telescope
	  use({
        "nvim-telescope/telescope.nvim",
        tag = "0.1.5",
        requires = { { "nvim-lua/plenary.nvim" } },
    })

	-- File manager
    use({
        "nvim-neo-tree/neo-tree.nvim",
        branch = "v2.x",
        requires = {
            "nvim-lua/plenary.nvim",
            "nvim-tree/nvim-web-devicons",
            "MunifTanjim/nui.nvim",
        },
    })

	-- Themes
	use({ 
		"catppuccin/nvim", 
		as = "catppuccin",
		config = function()
			require("configs.catppuccin")
		end,
	 })

	 -- Background Transparent
	use({
		"xiyaowong/nvim-transparent",
		config = function()
			require("configs.transparent")
		end, 
	})

	use({
        "neovim/nvim-lspconfig",
		requires = {
			'williamboman/mason.nvim',
			'williamboman/mason-lspconfig.nvim',
			'WhoIsSethDaniel/mason-tool-installer.nvim',
        },
        config = function()
            require("configs.lsp")
        end,
    })

	use("folke/todo-comments.nvim")

	use({
		"mfussenegger/nvim-dap",
		requires = {
			'rcarriga/nvim-dap-ui',

			-- Installs the debug adapters for you
			'williamboman/mason.nvim',
			'jay-babu/mason-nvim-dap.nvim',

			-- Add your own debuggers here
			'leoluz/nvim-dap-go',
		},
		config = function()
			require("configs.debug")
		end,
	})

	if PACKER_BOOTSTRAP then
		require("packer").sync()
	end
end)
