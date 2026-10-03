# pj — Zsh integration
#
# Install: add to ~/.zshrc
#     source /path/to/pj/contrib/shell/init.zsh
#
# See docs/ARCHITECTURE.md §3.3 and docs/ACTIONS_SPEC.md §10.
#
# localoptions makes subsequent setopt calls local to the function.
# localtraps makes traps local as well — without it, `trap ... EXIT`
# inside a function is a *global* trap that only fires when the shell
# exits, causing every call to leak a tmpfile. See ACTIONS_SPEC.md §10
# for the original form; we use localtraps to make it correct.

pj() {
    setopt localoptions localtraps

    local actions rc runtime

    runtime="${XDG_RUNTIME_DIR:-${TMPDIR:-/tmp}}"
    actions=$(mktemp "$runtime/pj-actions.XXXXXX") || return 1

    case "$actions" in
        "$runtime"/*) ;;
        *)
            print -u2 "pj: invalid actions file location: $actions"
            return 1
            ;;
    esac

    # ${(q)actions} shell-quotes the path for the trap. The value is
    # expanded at trap-set time so the trap does not rely on the local
    # being alive when it fires.
    trap "rm -f ${(q)actions}" EXIT

    PJ_SHELL=zsh \
    PJ_ACTIONS_FILE="$actions" \
        command pj "$@"
    rc=$?

    if [[ -s "$actions" ]]; then
        source "$actions"
    fi

    return "$rc"
}