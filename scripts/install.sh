#!/usr/bin/env bash
# scripts/install.sh — install pj manager + plugins + shell wrappers
#
# Manager   -> $BIN_DIR/pj
# Plugins   -> $PLUGIN_DIR/pj-*
# Wrappers  -> $XDG_CONFIG_HOME/pj/init.{bash,zsh,fish}
#
# Overrides:
#     PREFIX       default: $HOME/.local
#     BIN_DIR      default: $PREFIX/bin
#     PLUGIN_DIR   default: $HOME/.pj/bin
#     CONFIG_DIR   default: ${XDG_CONFIG_HOME:-$HOME/.config}/pj
#
# Requires: bin/ populated (run scripts/build.sh first).

set -euo pipefail

cd "$(dirname "$0")/.."

PREFIX="${PREFIX:-$HOME/.local}"
BIN_DIR="${BIN_DIR:-$PREFIX/bin}"
PLUGIN_DIR="${PLUGIN_DIR:-$HOME/.pj/bin}"
CONFIG_DIR="${CONFIG_DIR:-${XDG_CONFIG_HOME:-$HOME/.config}/pj}"

DRY_RUN=0
for arg in "$@"; do
    case "$arg" in
        -n|--dry-run) DRY_RUN=1 ;;
        -h|--help)
            sed -n '2,12p' "$0"
            exit 0
            ;;
        *)
            echo "install: unknown flag: $arg" >&2
            exit 2
            ;;
    esac
done

run() {
    if [ "$DRY_RUN" -eq 1 ]; then
        printf '  would: %s\n' "$*"
    else
        "$@"
    fi
}

if [ ! -d bin ] || [ -z "$(ls -A bin 2>/dev/null)" ]; then
    echo "install: bin/ is empty — run ./scripts/build.sh first" >&2
    exit 1
fi
if [ ! -f bin/pj ]; then
    echo "install: bin/pj not found — run ./scripts/build.sh first" >&2
    exit 1
fi

echo "pj install"
[ "$DRY_RUN" -eq 1 ] && echo "  (dry run — no changes will be made)"
echo "  manager  -> $BIN_DIR"
echo "  plugins  -> $PLUGIN_DIR"
echo "  wrappers -> $CONFIG_DIR"
echo

run mkdir -p "$BIN_DIR"
run chmod 0755 "$BIN_DIR"
run mkdir -p "$PLUGIN_DIR"
run chmod 0700 "$PLUGIN_DIR"
run mkdir -p "$CONFIG_DIR"
run chmod 0700 "$CONFIG_DIR"

# Manager
run install -m 0755 bin/pj "$BIN_DIR/pj"
printf '  %s\n' "$BIN_DIR/pj"

# Plugins
shopt -s nullglob
for f in bin/pj-*; do
    [ -f "$f" ] || continue
    name=$(basename "$f")
    run install -m 0755 "$f" "$PLUGIN_DIR/$name"
    printf '  %s\n' "$PLUGIN_DIR/$name"
done
shopt -u nullglob

# Shell wrappers — copy only if source exists and is newer or missing.
for w in init.bash init.zsh init.fish; do
    src="contrib/shell/$w"
    dst="$CONFIG_DIR/$w"
    [ -f "$src" ] || continue

    if [ "$DRY_RUN" -eq 1 ]; then
        echo "  would install $dst"
        continue
    fi

    if [ -f "$dst" ] && cmp -s "$src" "$dst"; then
        printf '  %s (unchanged)\n' "$dst"
        continue
    fi
    install -m 0644 "$src" "$dst"
    printf '  %s\n' "$dst"
done

# PATH hint
case ":$PATH:" in
*":$BIN_DIR:"*) ;;
*)
    echo
    echo "NOTE: $BIN_DIR is not on your PATH."
    echo "Add to your shell rc (bash/zsh):"
    echo "    export PATH=\"$BIN_DIR:\$PATH\""
    echo "Or for fish:"
    echo "    fish_add_path $BIN_DIR"
    ;;
esac

# Enable-the-wrapper hint
echo
echo "Enable cd-on-jump by adding ONE of these to your shell rc:"
echo "    bash: source $CONFIG_DIR/init.bash"
echo "    zsh:  source $CONFIG_DIR/init.zsh"
echo "    fish: source $CONFIG_DIR/init.fish"

# Post-install verification
if [ "$DRY_RUN" -eq 0 ]; then
    if "$BIN_DIR/pj" --version >/dev/null 2>&1; then
        echo
        echo "verify: $("$BIN_DIR/pj" --version)"
    else
        echo "verify: FAILED — pj --version did not run" >&2
        exit 1
    fi
fi

echo
echo "done. next: pj init && pj list"