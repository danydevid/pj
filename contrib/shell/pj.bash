#!/usr/bin/bash

export PJ_SHELL="bash"
export PJ_ACTIONS_FILE=$(mktemp "${XDG_RUNTIME_DIR:-/tmp}/pj-actions-XXXXXX")

bin/pj $@
