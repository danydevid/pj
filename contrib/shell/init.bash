# pj - Bash integration
#
# Install: add to ~/.bashrc
#     source /path/to/pj/contrib/shell/init.bash
#
# See docs/ARCHITECTURE.md section 3.3, docs/ACTIONS_SPEC.md section 9, 12.

pj() {
    local actions rc runtime

    # Prefer $XDG_RUNTIME_DIR, then $TMPDIR, then /tmp.
    if [ -n "${XDG_RUNTIME_DIR:-}" ]; then
        runtime="$XDG_RUNTIME_DIR"
    elif [ -n "${TMPDIR:-}" ]; then
        runtime="$TMPDIR"
    else
        runtime="/tmp"
    fi

    actions=$(mktemp "$runtime/pj-actions.XXXXXX") || return 1

    # Security check: refuse to source from an unexpected location.
    case "$actions" in
        "$runtime"/*) ;;
        *)
            printf 'pj: invalid actions file location: %s\n' "$actions" >&2
            return 1
            ;;
    esac

    # Register cleanup. printf %q quotes the path into the trap at set time.
    # shellcheck disable=SC2064
    trap "rm -f $(printf '%q' "$actions")" RETURN

    PJ_SHELL=bash \
    PJ_ACTIONS_FILE="$actions" \
        command pj "$@"
    rc=$?

    if [ -s "$actions" ]; then
        # shellcheck disable=SC1090
        . "$actions"
    fi

    return "$rc"
}