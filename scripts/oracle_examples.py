#!/usr/bin/env python3
"""Classify upstream examples for offline oracle execution.

This script intentionally defaults to classification only. Raw execution output
belongs under `.oracle/output` and must not be committed.
"""

from __future__ import annotations

import argparse
import json
from dataclasses import asdict, dataclass
from pathlib import Path


NETWORK_MARKERS = ("ArkFtpClient", "rcon", "RCON", "download_save_file", "upload_save_file")
MUTATION_MARKERS = (
    "store_db",
    "remove_",
    "modify_",
    "insert",
    "replace_owner",
    "change_owner",
    "force_grow",
)
EXPORT_MARKERS = ("export", "store_binary", "heatmap", "to_json", ".json", "draw_heatmap")
CLUSTER_MARKERS = ("cluster_data", "ClusterData", "ArkClusterData")


@dataclass(frozen=True)
class ExampleStatus:
    path: str
    category: str
    reason: str


def classify(path: Path, root: Path) -> ExampleStatus:
    text = path.read_text(encoding="utf-8", errors="replace")
    rel = path.relative_to(root).as_posix()
    lowered = text.lower()

    if "rcon_api" in rel or "rcon" in lowered:
        return ExampleStatus(rel, "network_skip", "RCON/live server behavior is out of scope")
    if rel.endswith("ex_ftp_config"):
        return ExampleStatus(rel, "network_skip", "FTP config sample")
    if any(marker in text for marker in CLUSTER_MARKERS):
        return ExampleStatus(rel, "local_cluster", "Local cluster-file parsing candidate")
    if any(marker in text for marker in MUTATION_MARKERS):
        return ExampleStatus(rel, "mutation_copy", "Mutates save data; run only on throwaway copies")
    if any(marker in text for marker in EXPORT_MARKERS):
        return ExampleStatus(rel, "export_output", "Offline-compatible but produces files/images")
    if any(marker in text for marker in NETWORK_MARKERS):
        return ExampleStatus(rel, "path_patch_readonly", "Patch acquisition code to local save path")
    return ExampleStatus(rel, "readonly", "No network or mutation marker detected")


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--examples", type=Path, default=Path(".oracle/upstream/examples"))
    parser.add_argument("--out", type=Path, default=Path(".oracle/output/example-classification.json"))
    args = parser.parse_args()

    statuses = [
        classify(path, args.examples)
        for path in sorted(args.examples.rglob("*.py"))
    ]
    args.out.parent.mkdir(parents=True, exist_ok=True)
    args.out.write_text(
        json.dumps([asdict(status) for status in statuses], indent=2, sort_keys=True),
        encoding="utf-8",
    )

    counts: dict[str, int] = {}
    for status in statuses:
        counts[status.category] = counts.get(status.category, 0) + 1
    for category, count in sorted(counts.items()):
        print(f"{category}: {count}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
