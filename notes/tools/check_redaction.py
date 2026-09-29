#!/usr/bin/env python3
"""Check public files before committing them to the course repository.

The optional local map adds project-specific forbidden terms. Generic checks run
on a fresh clone too, where notes/private/redaction-map.json is unavailable.
"""

import json
import re
import subprocess
import sys
from pathlib import Path


MAP_PATH = Path("notes/private/redaction-map.json")
PRIVATE_PATTERNS = {
    "email": re.compile(r"(?<![\w.+-])[\w.+-]+@[\w.-]+\.[A-Za-z]{2,}"),
    "home path": re.compile(r"/(?:Users|home)/[A-Za-z0-9_.-]+/"),
    "temporary path": re.compile(r"/(?:tmp|var/folders|mnt/data)/[^\s`\"'<>]+"),
    "private IP": re.compile(
        r"(?<!\d)(?:10\.(?:\d{1,3}\.){2}\d{1,3}|"
        r"192\.168\.(?:\d{1,3}\.)\d{1,3}|"
        r"172\.(?:1[6-9]|2\d|3[01])\.(?:\d{1,3}\.)\d{1,3})(?!\d)"
    ),
    "UUID": re.compile(r"\b[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}\b", re.I),
}


def public_paths():
    output = subprocess.run(
        ["git", "ls-files", "--cached", "--others", "--exclude-standard", "-z"],
        capture_output=True, check=True,
    ).stdout
    return [Path(p.decode()) for p in output.split(b"\0") if p]


def main():
    words = []
    if MAP_PATH.exists():
        mapping = json.loads(MAP_PATH.read_text(encoding="utf-8"))
        words = [w.casefold() for w in mapping.get("forbidden", []) if w]

    failures = []
    for path in public_paths():
        if not path.is_file():
            continue
        try:
            content = path.read_text(encoding="utf-8")
        except UnicodeDecodeError:
            continue
        problems = []
        lowered = re.sub(r"[A-Za-z0-9+/=]{200,}", "", content).casefold()
        if any(word in lowered for word in words):
            problems.append("local forbidden term")
        if path.parts[:2] == ("notes", "sessions"):
            problems += [name for name, pattern in PRIVATE_PATTERNS.items() if pattern.search(content)]
        if problems:
            failures.append(f"{path}: {', '.join(problems)}")

    if failures:
        print("Public-content check failed:")
        print("\n".join(failures))
        return 1
    print("Public-content check passed" + (" (local map included)" if words else " (generic checks only)"))
    return 0


if __name__ == "__main__":
    sys.exit(main())
