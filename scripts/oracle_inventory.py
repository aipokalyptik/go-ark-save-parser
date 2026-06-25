#!/usr/bin/env python3
"""Inventory private ARK backup data without committing sensitive details.

The private manifest contains relative paths and hashes and is written under
`.oracle/`, which is ignored by git. The committed summary contains only counts
and coarse size ranges.
"""

from __future__ import annotations

import argparse
import hashlib
import json
from collections import Counter
from dataclasses import asdict, dataclass
from pathlib import Path


KNOWN_EXTENSIONS = {
    ".ark": "map_save",
    ".arkprofile": "player_profile",
    ".arktribe": "tribe_save",
    ".arktributetribe": "tribute_tribe",
    ".arktributetribetribe": "tribute_tribe",
}

CLUSTER_MARKERS = {
    ".arktribute": "local_cluster",
    ".arktributetribe": "local_cluster",
    ".arktributetribetribe": "local_cluster",
}


@dataclass(frozen=True)
class Entry:
    kind: str
    extension: str
    relative_path: str
    size_bytes: int
    sha256: str


def sha256(path: Path) -> str:
    h = hashlib.sha256()
    with path.open("rb") as f:
        for chunk in iter(lambda: f.read(1024 * 1024), b""):
            h.update(chunk)
    return h.hexdigest()


def classify(path: Path) -> str | None:
    name = path.name.lower()
    suffix = path.suffix.lower()
    if suffix in KNOWN_EXTENSIONS:
        return KNOWN_EXTENSIONS[suffix]
    if suffix in CLUSTER_MARKERS:
        return CLUSTER_MARKERS[suffix]
    if "cluster" in name and path.is_file():
        return "local_cluster_candidate"
    return None


def size_bucket(size: int) -> str:
    mib = size / (1024 * 1024)
    if mib < 1:
        return "<1MiB"
    if mib < 10:
        return "1-10MiB"
    if mib < 100:
        return "10-100MiB"
    if mib < 1024:
        return "100MiB-1GiB"
    return ">=1GiB"


def build_inventory(root: Path) -> list[Entry]:
    entries: list[Entry] = []
    for path in sorted(p for p in root.rglob("*") if p.is_file()):
        kind = classify(path)
        if kind is None:
            continue
        entries.append(
            Entry(
                kind=kind,
                extension=path.suffix.lower(),
                relative_path=path.relative_to(root).as_posix(),
                size_bytes=path.stat().st_size,
                sha256=sha256(path),
            )
        )
    return entries


def write_summary(entries: list[Entry], summary_path: Path) -> None:
    by_kind = Counter(e.kind for e in entries)
    by_extension = Counter(e.extension or "<none>" for e in entries)
    by_size = Counter(size_bucket(e.size_bytes) for e in entries)
    total_bytes = sum(e.size_bytes for e in entries)

    lines = [
        "# Oracle Backup Summary",
        "",
        "This file is safe to commit. It intentionally contains only aggregate",
        "counts for the private oracle backup and does not include file paths,",
        "hashes, player names, tribe names, object identifiers, or raw output.",
        "",
        "## Source",
        "",
        "- Tarball: `~/Downloads/SavedArks.tar.bz2`",
        "- Private extraction root: `.oracle/data`",
        "- Private manifest: `.oracle/manifest.json`",
        "",
        "## Counts By Kind",
        "",
    ]

    if by_kind:
        lines.extend(f"- `{kind}`: {count}" for kind, count in sorted(by_kind.items()))
    else:
        lines.append("- No ARK save-like files were detected.")

    lines.extend(["", "## Counts By Extension", ""])
    lines.extend(f"- `{ext}`: {count}" for ext, count in sorted(by_extension.items()))

    lines.extend(["", "## Size Buckets", ""])
    lines.extend(f"- `{bucket}`: {count}" for bucket, count in sorted(by_size.items()))

    lines.extend(
        [
            "",
            "## Aggregate Size",
            "",
            f"- Total detected save-like bytes: {total_bytes}",
            "",
        ]
    )

    summary_path.parent.mkdir(parents=True, exist_ok=True)
    summary_path.write_text("\n".join(lines), encoding="utf-8")


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--root", type=Path, default=Path(".oracle/data"))
    parser.add_argument("--manifest", type=Path, default=Path(".oracle/manifest.json"))
    parser.add_argument("--summary", type=Path, default=Path("docs/oracle-summary.md"))
    args = parser.parse_args()

    entries = build_inventory(args.root)
    args.manifest.parent.mkdir(parents=True, exist_ok=True)
    args.manifest.write_text(
        json.dumps(
            {
                "root": str(args.root),
                "entries": [asdict(e) for e in entries],
            },
            indent=2,
            sort_keys=True,
        ),
        encoding="utf-8",
    )
    write_summary(entries, args.summary)
    print(f"wrote private manifest with {len(entries)} entries")
    print(f"wrote sanitized summary to {args.summary}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
