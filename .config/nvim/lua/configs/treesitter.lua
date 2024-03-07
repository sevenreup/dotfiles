local status, ts = pcall(require, "nvim-treesitter.configs")
if not status then
    return
end


ts.setup({
    ensure_installed = {
        "markdown",
        "tsx",
        "typescript",
        "javascript",
        "bash",
        "json",
        "yaml",
        "go",
        "css",
        "html",
        "lua"
    },
    uto_install = true,
    highlight = { enable = true },
    indent = { enable = true },
})