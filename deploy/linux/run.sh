#!/usr/bin/env sh
set -eu
ROOT=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
UI_HOST=${CFC_TK_UI_HOST:-127.0.0.1}
UI_PORT=${CFC_TK_UI_PORT:-5173}
SERVER_PID=
UI_PID=
cleanup() {
    trap - EXIT INT TERM
    [ -z "$UI_PID" ] || kill "$UI_PID" 2>/dev/null || true
    [ -z "$SERVER_PID" ] || kill "$SERVER_PID" 2>/dev/null || true
    [ -z "$UI_PID" ] || wait "$UI_PID" 2>/dev/null || true
    [ -z "$SERVER_PID" ] || wait "$SERVER_PID" 2>/dev/null || true
}
trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM
if [ ! -x "$ROOT/server/bin/cfcbench-server" ] || [ ! -f "$ROOT/build/200.html" ]; then
    echo "Run ./install.sh first." >&2
    exit 1
fi
export PATH="$HOME/.local/bin:$PATH"
(cd "$ROOT/server" && exec ./bin/cfcbench-server) &
SERVER_PID=$!
sleep 1
if ! kill -0 "$SERVER_PID" 2>/dev/null; then
    echo "API failed to start; see the error above." >&2
    exit 1
fi
cd "$ROOT"
node node_modules/vite/bin/vite.js preview --host "$UI_HOST" --port "$UI_PORT" --strictPort &
UI_PID=$!
echo "CFCBench: http://$UI_HOST:$UI_PORT | API: http://localhost:8765"
echo "Press Ctrl+C to stop."
while kill -0 "$SERVER_PID" 2>/dev/null && kill -0 "$UI_PID" 2>/dev/null; do
    sleep 1
done
echo "A service exited. Stopping CFCBench." >&2
exit 1
