#!/usr/bin/env python3
"""Run narrow private oracle probes for concrete Go parity blockers.

Detailed probe values are private and written under `.oracle/output`. Stdout is
limited to status metadata so it can be pasted into notes without leaking save
identifiers, player names, locations, or object UUIDs.
"""

from __future__ import annotations

import argparse
import contextlib
import json
import os
import sys
import traceback
from dataclasses import asdict, dataclass
from pathlib import Path
from typing import Any, Callable


@dataclass(frozen=True)
class ProbeResult:
    name: str
    status: str
    detail: str
    output: str


def default_output_path(repo_root: Path, probe_name: str) -> Path:
    return repo_root / ".oracle" / "output" / f"oracle-probe-{probe_name.replace('_', '-')}.json"


def summary_rows(results: list[ProbeResult]) -> list[dict[str, str]]:
    return [
        {
            "name": result.name,
            "status": result.status,
            "detail": result.detail,
        }
        for result in results
    ]


def upstream_src(repo_root: Path) -> Path:
    return repo_root / ".oracle" / "upstream" / "src"


def write_private_json(path: Path, payload: dict[str, Any]) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(json.dumps(payload, indent=2, sort_keys=True), encoding="utf-8")


def probe_equipment_rank(save_path: Path, repo_root: Path) -> dict[str, Any]:
    sys.path.insert(0, str(upstream_src(repo_root)))
    import oracle_compare  # type: ignore

    return oracle_compare.python_equipment_rank_oracle(save_path, upstream_src(repo_root))


def probe_dino_cryopod_location(save_path: Path, repo_root: Path) -> dict[str, Any]:
    sys.path.insert(0, str(upstream_src(repo_root)))
    from arkparse.api.dino_api import DinoApi  # type: ignore
    from arkparse.saves.asa_save import AsaSave  # type: ignore

    save = AsaSave(save_path)
    try:
        dinos = DinoApi(save).get_all_in_cryopod()
        with_locations = sum(1 for dino in dinos.values() if getattr(dino, "location", None) is not None)
        with_cryopod = sum(1 for dino in dinos.values() if getattr(dino, "cryopod", None) is not None)
        owner_teams = sorted(
            {
                getattr(getattr(dino, "owner", None), "target_team", None)
                for dino in dinos.values()
                if getattr(getattr(dino, "owner", None), "target_team", None) is not None
            }
        )
        return {
            "cryopodded": len(dinos),
            "with_locations": with_locations,
            "with_cryopod": with_cryopod,
            "owner_team_count": len(owner_teams),
        }
    finally:
        save.close()


PROBES: dict[str, Callable[[Path, Path], dict[str, Any]]] = {
    "equipment_rank": probe_equipment_rank,
    "dino_cryopod_location": probe_dino_cryopod_location,
}


def run_probe(name: str, save_path: Path, repo_root: Path) -> ProbeResult:
    output = default_output_path(repo_root, name)
    probe = PROBES[name]
    try:
        payload = probe(save_path, repo_root)
    except Exception as exc:  # pragma: no cover - exercised by private saves.
        write_private_json(
            output,
            {
                "probe": name,
                "status": "error",
                "error_type": type(exc).__name__,
                "error": str(exc),
                "traceback": traceback.format_exc(),
            },
        )
        return ProbeResult(name=name, status="error", detail=type(exc).__name__, output=str(output))

    write_private_json(
        output,
        {
            "probe": name,
            "status": "pass",
            "payload": payload,
        },
    )
    return ProbeResult(name=name, status="pass", detail="private output written", output=str(output))


def parse_args(argv: list[str]) -> argparse.Namespace:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--probe", choices=sorted(PROBES) + ["all"], default="all")
    parser.add_argument("--save", default=os.environ.get("ARK_ORACLE_SAVE"))
    return parser.parse_args(argv)


def main(argv: list[str] | None = None) -> int:
    args = parse_args(argv or sys.argv[1:])
    if not args.save:
        print("ARK_ORACLE_SAVE or --save is required", file=sys.stderr)
        return 2

    repo_root = Path.cwd()
    save_path = Path(args.save)
    names = sorted(PROBES) if args.probe == "all" else [args.probe]
    results = [run_probe(name, save_path, repo_root) for name in names]
    print(json.dumps(summary_rows(results), indent=2, sort_keys=True))
    return 0


if __name__ == "__main__":
    with contextlib.suppress(KeyboardInterrupt):
        raise SystemExit(main())
    raise SystemExit(130)
