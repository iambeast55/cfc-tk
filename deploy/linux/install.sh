#!/usr/bin/env sh
set -eu
ROOT=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
cd "$ROOT"
for tool in node npm go python3; do
    command -v "$tool" >/dev/null 2>&1 || { echo "Missing dependency: $tool" >&2; exit 1; }
done
node -e 'if (Number(process.versions.node.split(".")[0]) < 22) { console.error("Node.js 22 or newer is required"); process.exit(1); }'
echo "Installing UI dependencies..."
npm ci
echo "Building UI..."
npm run build
echo "Building API..."
mkdir -p server/bin
(cd server && go build -buildvcs=false -o bin/cfcbench-server .)
chmod +x run.sh
echo "Installed. Start with ./run.sh"
