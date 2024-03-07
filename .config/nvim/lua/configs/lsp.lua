local status, nvim_lsp = pcall(require, "lspconfig")
if not status then
    return
end

require('mason').setup()

local protocol = require("vim.lsp.protocol")

require('mason-tool-installer').setup { ensure_installed = ensure_installed }

-- You can add other tools here that you want Mason to install
-- for you, so that they are available from within Neovim.
local ensure_installed = vim.tbl_keys(servers or {})
vim.list_extend(ensure_installed, {
'stylua', -- Used to format lua code
})

require('mason-lspconfig').setup {
    handlers = {
        function(server_name)
        local server = servers[server_name] or {}
        -- This handles overriding only values explicitly passed
        -- by the server configuration above. Useful when disabling
        -- certain features of an LSP (for example, turning off formatting for tsserver)
        server.capabilities = vim.tbl_deep_extend('force', {}, capabilities, server.capabilities or {})
        require('lspconfig')[server_name].setup(server)
        end,
    },
}

local on_attach = function(client, bufnr)
    -- format on save
    if client.server_capabilities.documentFormattingProvider then
        vim.api.nvim_create_autocmd("BufWritePre", {
            group = vim.api.nvim_create_augroup("Format", { clear = true }),
            buffer = bufnr,
            callback = function()
                vim.lsp.buf.formatting_seq_sync()
            end,
        })
    end
end