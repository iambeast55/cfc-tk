"""Build a Linux source deployment archive using an explicit file allowlist."""

import hashlib
import io
import json
from pathlib import Path
import tarfile


def main():
    root = Path(__file__).resolve().parent.parent
    version = json.loads((root / "package.json").read_text())["version"]
    name = f"cfcbench-linux-{version}"
    output = root / "dist" / f"{name}.tar.gz"
    output.parent.mkdir(exist_ok=True)
    files = {}
    for filename in ("package.json", "package-lock.json", "svelte.config.js",
                     "vite.config.ts", "tsconfig.json", "components.json", ".npmrc"):
        path = root / filename
        if path.is_file():
            files[filename] = path
    for directory in ("src", "static"):
        for path in (root / directory).rglob("*"):
            if path.is_file() and not path.is_symlink():
                files[path.relative_to(root).as_posix()] = path
    for path in (root / "server").iterdir():
        if path.is_file() and (path.suffix == ".go" or path.name in ("go.mod", "go.sum")):
            files[f"server/{path.name}"] = path
    for filename in ("install.sh", "run.sh", "README.md"):
        files[filename] = root / "deploy" / "linux" / filename
    with tarfile.open(output, "w:gz") as archive:
        for relative, path in sorted(files.items()):
            data = path.read_bytes()
            if path.suffix in (".sh", ".go", ".js", ".ts", ".svelte", ".css", ".md", ".json"):
                data = data.replace(b"\r\n", b"\n")
            info = tarfile.TarInfo(f"{name}/{relative}")
            info.size = len(data)
            info.mode = 0o755 if relative.endswith(".sh") else 0o644
            archive.addfile(info, io.BytesIO(data))
    digest = hashlib.sha256(output.read_bytes()).hexdigest()
    output.with_suffix(output.suffix + ".sha256").write_text(
        f"{digest}  {output.name}\n", encoding="ascii")
    print(f"Created {output} ({len(files)} files)")


if __name__ == "__main__":
    main()
