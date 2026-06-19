#!/usr/bin/env python3
"""Run privacy-safe oracle comparisons for implemented offline Go examples.

Detailed values and stdout are private and written under `.oracle/output`.
The committed markdown summary contains only case names and pass/fail/skip
statuses so it can be reviewed without leaking save-derived identifiers.
"""

from __future__ import annotations

import argparse
import json
import os
import re
import subprocess
import sys
from dataclasses import asdict, dataclass
from pathlib import Path
from typing import Any


@dataclass(frozen=True)
class CaseResult:
    name: str
    status: str
    detail: str


def run(cmd: list[str], cwd: Path, env: dict[str, str]) -> subprocess.CompletedProcess[str]:
    return subprocess.run(
        cmd,
        cwd=cwd,
        env=env,
        text=True,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        check=False,
    )


def parse_go_map_summary(output: str) -> dict[str, Any]:
    match = re.fullmatch(
        r"map=(?P<map>.*?) save_version=(?P<save_version>\d+) objects=(?P<objects>\d+) names=(?P<names>\d+)\n?",
        output,
    )
    if not match:
        raise ValueError("unexpected map_summary output")
    return {
        "map_name": match.group("map"),
        "save_version": int(match.group("save_version")),
        "object_count": int(match.group("objects")),
        "name_count": int(match.group("names")),
    }


def python_oracle(save_path: Path, upstream_src: Path) -> dict[str, Any]:
    sys.path.insert(0, str(upstream_src))
    from arkparse.saves.asa_save import AsaSave  # type: ignore

    save = AsaSave(save_path)
    try:
        classes = sorted(save.get_all_present_classes() or [])
        object_ids = save.save_connection.get_obj_uuids()
        return {
            "map_name": save.save_context.map_name,
            "save_version": save.save_context.save_version,
            "object_count": len(object_ids),
            "name_count": len(save.save_context.names),
            "classes": classes,
            "class_count": len(classes),
        }
    finally:
        save.close()


def compare(save_path: Path, repo_root: Path, upstream_src: Path) -> tuple[list[CaseResult], dict[str, Any]]:
    env = os.environ.copy()
    env.setdefault("PYTHONWARNINGS", "ignore")
    py = python_oracle(save_path, upstream_src)
    private: dict[str, Any] = {"save_path": str(save_path), "python": py, "go": {}}
    cases: list[CaseResult] = []

    go_map = run(["go", "run", "./examples/map_summary", str(save_path)], repo_root, env)
    private["go"]["map_summary"] = {
        "exit_code": go_map.returncode,
        "stdout": go_map.stdout,
        "stderr": go_map.stderr,
    }
    if go_map.returncode != 0:
        cases.append(CaseResult("map_summary", "fail", "Go example exited non-zero"))
    else:
        try:
            got = parse_go_map_summary(go_map.stdout)
            private["go"]["map_summary"]["parsed"] = got
            want = {k: py[k] for k in ("map_name", "save_version", "object_count", "name_count")}
            cases.append(CaseResult("map_summary", "pass" if got == want else "fail", "summary metrics compared"))
        except Exception as exc:  # noqa: BLE001 - private report captures details
            private["go"]["map_summary"]["parse_error"] = str(exc)
            cases.append(CaseResult("map_summary", "fail", "Go output could not be parsed"))

    go_classes = run(["go", "run", "./examples/object_classes", str(save_path)], repo_root, env)
    private["go"]["object_classes"] = {
        "exit_code": go_classes.returncode,
        "stdout": go_classes.stdout,
        "stderr": go_classes.stderr,
    }
    if go_classes.returncode != 0:
        cases.append(CaseResult("object_classes", "fail", "Go example exited non-zero"))
    else:
        got_classes = sorted(line for line in go_classes.stdout.splitlines() if line)
        private["go"]["object_classes"]["classes"] = got_classes
        cases.append(CaseResult("object_classes", "pass" if got_classes == py["classes"] else "fail", "class list compared"))

    return cases, private


def write_summary(path: Path, cases: list[CaseResult]) -> None:
    counts: dict[str, int] = {}
    for case in cases:
        counts[case.status] = counts.get(case.status, 0) + 1
    lines = [
        "# Oracle Comparison Summary",
        "",
        "This file is safe to commit. It records only aggregate status for the",
        "implemented offline Go example comparisons. Detailed oracle values, paths,",
        "class names, stdout, and stderr stay in `.oracle/output/oracle-comparison.json`.",
        "",
        "## Case Status",
        "",
    ]
    for case in cases:
        lines.append(f"- `{case.name}`: `{case.status}` ({case.detail})")
    lines.extend(["", "## Counts", ""])
    for status, count in sorted(counts.items()):
        lines.append(f"- `{status}`: {count}")
    lines.append("")
    path.write_text("\n".join(lines), encoding="utf-8")


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--save", type=Path, default=None)
    parser.add_argument("--upstream-src", type=Path, default=Path(".oracle/upstream/src"))
    parser.add_argument("--out", type=Path, default=Path(".oracle/output/oracle-comparison.json"))
    parser.add_argument("--summary", type=Path, default=Path("docs/oracle-comparison-summary.md"))
    args = parser.parse_args()

    repo_root = Path.cwd()
    save_arg = args.save or os.environ.get("ARK_ORACLE_SAVE")
    if not save_arg:
        raise SystemExit("missing --save or ARK_ORACLE_SAVE")
    save_path = Path(save_arg).expanduser().resolve()
    if not save_path.exists():
        raise SystemExit("oracle save does not exist")

    cases, private = compare(save_path, repo_root, args.upstream_src)
    args.out.parent.mkdir(parents=True, exist_ok=True)
    args.out.write_text(json.dumps({"cases": [asdict(c) for c in cases], **private}, indent=2, sort_keys=True), encoding="utf-8")
    write_summary(args.summary, cases)

    for status, count in sorted({c.status: sum(1 for item in cases if item.status == c.status) for c in cases}.items()):
        print(f"{status}: {count}")
    return 0 if all(case.status == "pass" for case in cases) else 1


if __name__ == "__main__":
    raise SystemExit(main())
