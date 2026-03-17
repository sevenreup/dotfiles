return {
  -- Disable snacks explorer so neo-tree is the only file explorer
  {
    "folke/snacks.nvim",
    opts = {
      explorer = { enabled = false },
    },
  },

  {
    "nvim-neo-tree/neo-tree.nvim",
    opts = {
      filesystem = {
        filtered_items = {
          visible = true,
          show_hidden_count = true,
          hide_dotfiles = false,
          hide_gitignored = false,
        },
      },
    },
  },
}
