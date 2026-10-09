#!/usr/bin/env python3
"""Check public files before committing them to the course repository.

The optional local map adds project-specific forbidden terms. Generic checks run
on a fresh clone too, where notes/private/redaction-map.json is unavailable.
"""

import json
import subprocess
import sys
from pathlib import Path
from public_redaction import IDENTIFIER_FIELDS, PATTERNS, SECRET_FIELDS, strings


MAP_PATH = Path("notes/private/redaction-map.json")
PRIVATE_PATTERNS = PATTERNS


def public_paths():
    output = subprocess.run(
        ["git", "ls-files", "--cached", "--others", "--exclude-standard", "-z"],
        capture_output=True, check=True,
    ).stdout
    return [Path(p.decode()) for p in output.split(b"\0") if p]


def metadata_findings(value):
    findings = set()
    if isinstance(value, dict):
        for key, item in value.items():
            if isinstance(item, str) and item and not item.startswith(("〔", "[omitted]", "[已")):
                if key.lower() in IDENTIFIER_FIELDS:
                    findings.add("unmasked metadata ID")
                if key.lower() in SECRET_FIELDS:
                    findings.add("unmasked credential/signature field")
            findings.update(metadata_findings(item))
    elif isinstance(value, list):
        for item in value:
            findings.update(metadata_findings(item))
    return findings


def main():
    words = []
    if MAP_PATH.exists():
        mapping = json.loads(MAP_PATH.read_text(encoding="utf-8"))
        words = [w.casefold() for w in mapping.get("forbidden", []) if w]

    failures = []
    for path in public_paths():
        if path.parts[:2] == ("notes", "private"):
            failures.append(f"{path}: private originals must not be tracked")
            continue
        if not path.is_file():
            continue
        try:
            content = path.read_text(encoding="utf-8")
        except UnicodeDecodeError:
            continue
        problems = []
        lowered = content.casefold()
        if any(word in lowered for word in words):
            problems.append("local forbidden term")
        if path.parts[:2] in (("notes", "sessions"), ("notes", "raw")):
            problems += [name for name, pattern in PRIVATE_PATTERNS.items() if pattern.search(content)]
            if path.suffix == ".jsonl":
                try:
                    # Decode escaped JSON strings before checking embedded tool content.
                    events = [json.loads(line) for line in content.splitlines() if line.strip()]
                    problems += sorted(metadata_findings(events))
                    content = "\n".join(text for event in events for text in strings(event))
                    problems += [name for name, pattern in PRIVATE_PATTERNS.items() if pattern.search(content)]
                    if any(word in content.casefold() for word in words):
                        problems.append("escaped local forbidden term")
                except json.JSONDecodeError:
                    problems.append("invalid JSONL")
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
