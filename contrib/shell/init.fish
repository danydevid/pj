# pj — Fish integration
#
# Install: add to ~/.config/fish/config.fish
#     source /path/to/pj/contrib/shell/init.fish
#
# See docs/ARCHITECTURE.md §3.3 and docs/ACTIONS_SPEC.md §11.

function pj
    set -lx PJ_SHELL fish
    command pj $argv
end