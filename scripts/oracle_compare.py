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
import struct
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


def parse_key_value_lines(output: str) -> dict[str, Any]:
    values: dict[str, Any] = {}
    for line in output.splitlines():
        for part in line.split():
            if "=" not in part:
                continue
            key, value = part.split("=", 1)
            try:
                if "." in value:
                    values[key] = float(value)
                else:
                    values[key] = int(value)
            except ValueError:
                values[key] = value
    return values


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


def python_local_profiles_oracle(save_path: Path, repo_root: Path, upstream_src: Path) -> dict[str, Any]:
    sys.path.insert(0, str(upstream_src))
    from arkparse.api.player_api import PlayerApi  # type: ignore
    from arkparse.saves.asa_save import AsaSave  # type: ignore

    save_dir = save_path.parent
    empty_cluster_dir = repo_root / ".oracle" / "output" / "empty-cluster"
    empty_cluster_dir.mkdir(parents=True, exist_ok=True)
    save = AsaSave(save_path)
    try:
        player_api = PlayerApi(
            save,
            no_pawns=True,
            bypass_inventory=True,
            force_legacy_store=True,
            cluster_data_dir=empty_cluster_dir,
        )
        return {
            "profiles": len(player_api.profile_paths),
            "tribes": len(player_api.tribe_paths),
            "clusters": len([path for path in save_dir.iterdir() if path.is_file() and path.suffix == "" and not path.name.startswith(".")]),
            "tributes": len(list(save_dir.glob("*.arktributetribe"))) + len(list(save_dir.glob("*.arktributetribetribe"))),
            "parsed_players": len(player_api.players),
            "parsed_tribes": len(player_api.tribes),
            "tribe_player_links": sum(len(players) for players in player_api.tribe_to_player_map.values()),
        }
    finally:
        save.close()


def python_dino_filter_oracle(save_path: Path, upstream_src: Path) -> dict[str, Any]:
    sys.path.insert(0, str(upstream_src))
    from arkparse.api.dino_api import DinoApi  # type: ignore
    from arkparse.object_model.dinos.tamed_dino import TamedDino  # type: ignore
    from arkparse.saves.asa_save import AsaSave  # type: ignore

    save = AsaSave(save_path)
    try:
        dino_api = DinoApi(save)
        dinos = dino_api.get_all(include_cryos=False, max_workers=1, bypass_inventory=True)
        tamed = sum(isinstance(dino, TamedDino) for dino in dinos.values())
        return {
            "dinos": len(dinos),
            "tamed": tamed,
            "wild": len(dinos) - tamed,
            "classes": len({dino.object.blueprint for dino in dinos.values()}),
        }
    finally:
        save.close()


def python_property_filter_oracle(save_path: Path, upstream_src: Path) -> dict[str, Any]:
    sys.path.insert(0, str(upstream_src))
    from arkparse.parsing import GameObjectReaderConfiguration  # type: ignore
    from arkparse.saves.asa_save import AsaSave  # type: ignore

    save = AsaSave(save_path)
    try:
        config = GameObjectReaderConfiguration(property_names=["TamerString", "Health"])
        objects = save.get_game_objects(config)
        return {
            "objects": len(objects),
            "classes": len({obj.blueprint for obj in objects.values()}),
        }
    finally:
        save.close()


def python_stackable_count_oracle(save_path: Path, upstream_src: Path) -> dict[str, Any]:
    sys.path.insert(0, str(upstream_src))
    from arkparse.api.stackable_api import StackableApi  # type: ignore
    from arkparse.saves.asa_save import AsaSave  # type: ignore

    save = AsaSave(save_path)
    try:
        blueprint = next(
            blueprint
            for blueprint in sorted(save.get_all_present_classes() or [])
            if StackableApi.is_applicable_bp(blueprint)
        )
        api = StackableApi(save)
        items = api.get_by_class(StackableApi.Classes.RESOURCE, [blueprint])
        return {
            "blueprint": blueprint,
            "items": len(items),
            "total": api.get_count(items),
        }
    finally:
        save.close()


def python_cluster_data_oracle(upstream_src: Path) -> dict[str, Any] | None:
    sys.path.insert(0, str(upstream_src))
    from arkparse.object_model.cluster_data.ark_cluster_data import ClusterData  # type: ignore

    cluster_dir = upstream_src.parent / "examples" / "player_api" / "cluster_data"
    files = sorted(path for path in cluster_dir.iterdir() if path.is_file())
    if not files:
        return None
    cluster_file = files[0]
    cluster = ClusterData(cluster_file.parent, cluster_file.name)
    return {
        "path": str(cluster_file),
        "id": cluster_file.name,
        "items": len(cluster.items),
        "dinos": len(cluster.dinos),
    }


def python_local_tribute_oracle(save_path: Path) -> dict[str, Any]:
    save_dir = save_path.parent
    files = sorted(
        path
        for path in save_dir.iterdir()
        if path.is_file() and path.suffix in {".arktributetribe", ".arktributetribetribe"}
    )
    player_ids = 0
    tribe_ids = 0
    for path in files:
        raw = path.read_bytes()
        offset = 0
        for label in ("player_data_ids", "tribe_data_ids"):
            if offset + 4 > len(raw):
                raise ValueError(f"{path.name}: missing {label} count")
            count = struct.unpack_from("<i", raw, offset)[0]
            offset += 4
            if count < 0:
                raise ValueError(f"{path.name}: negative {label} count")
            byte_count = count * 8
            if offset + byte_count > len(raw):
                raise ValueError(f"{path.name}: {label} count exceeds file size")
            if label == "player_data_ids":
                player_ids += count
            else:
                tribe_ids += count
            offset += byte_count
        if offset != len(raw):
            raise ValueError(f"{path.name}: trailing bytes")
    return {
        "tribute_files": len(files),
        "player_data_ids": player_ids,
        "tribe_data_ids": tribe_ids,
    }


def compare(save_path: Path, repo_root: Path, upstream_src: Path) -> tuple[list[CaseResult], dict[str, Any]]:
    env = os.environ.copy()
    env.setdefault("PYTHONWARNINGS", "ignore")
    py = python_oracle(save_path, upstream_src)
    py_local_profiles = python_local_profiles_oracle(save_path, repo_root, upstream_src)
    py_dino_filter = python_dino_filter_oracle(save_path, upstream_src)
    py_property_filter = python_property_filter_oracle(save_path, upstream_src)
    py_stackable_count = python_stackable_count_oracle(save_path, upstream_src)
    py_cluster_data = python_cluster_data_oracle(upstream_src)
    py_local_tribute = python_local_tribute_oracle(save_path)
    private: dict[str, Any] = {
        "save_path": str(save_path),
        "python": py,
        "python_local_profiles": py_local_profiles,
        "python_dino_filter": py_dino_filter,
        "python_property_filter": py_property_filter,
        "python_stackable_count": py_stackable_count,
        "python_cluster_data": py_cluster_data,
        "python_local_tribute": py_local_tribute,
        "go": {},
    }
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

    export_json_path = repo_root / ".oracle" / "output" / "export-save-info.json"
    go_export_json = run(["go", "run", "./cmd/arksave", "export-json", str(save_path), str(export_json_path)], repo_root, env)
    private["go"]["export_json"] = {
        "exit_code": go_export_json.returncode,
        "stdout": go_export_json.stdout,
        "stderr": go_export_json.stderr,
        "output": str(export_json_path),
    }
    if go_export_json.returncode != 0:
        cases.append(CaseResult("export_json", "fail", "Go CLI exited non-zero"))
    else:
        try:
            got_export = json.loads(export_json_path.read_text(encoding="utf-8"))
            private["go"]["export_json"]["parsed"] = got_export
            got_classes = sorted({item["class_name"] for item in got_export.get("objects", [])})
            want = {k: py[k] for k in ("map_name", "save_version", "object_count", "name_count")}
            got = {k: got_export[k] for k in want}
            status = "pass" if got == want and got_classes == py["classes"] else "fail"
            cases.append(CaseResult("export_json", status, "save-info JSON metrics and class list compared"))
        except Exception as exc:  # noqa: BLE001 - private report captures details
            private["go"]["export_json"]["parse_error"] = str(exc)
            cases.append(CaseResult("export_json", "fail", "Go JSON export could not be parsed"))

    go_local_profiles = run(["go", "run", "./examples/local_profiles", str(save_path.parent)], repo_root, env)
    private["go"]["local_profiles"] = {
        "exit_code": go_local_profiles.returncode,
        "stdout": go_local_profiles.stdout,
        "stderr": go_local_profiles.stderr,
    }
    if go_local_profiles.returncode != 0:
        cases.append(CaseResult("local_profiles", "fail", "Go example exited non-zero"))
    else:
        got = parse_key_value_lines(go_local_profiles.stdout)
        private["go"]["local_profiles"]["parsed"] = got
        want = {key: py_local_profiles[key] for key in ("profiles", "tribes", "clusters", "tributes", "parsed_players", "parsed_tribes", "tribe_player_links")}
        cases.append(CaseResult("local_profiles", "pass" if {key: got.get(key) for key in want} == want else "fail", "local profile and tribe aggregate counts compared"))

    go_dino_filter = run(["go", "run", "./examples/dino_filter", str(save_path)], repo_root, env)
    private["go"]["dino_filter"] = {
        "exit_code": go_dino_filter.returncode,
        "stdout": go_dino_filter.stdout,
        "stderr": go_dino_filter.stderr,
    }
    if go_dino_filter.returncode != 0:
        cases.append(CaseResult("dino_filter", "fail", "Go example exited non-zero"))
    else:
        got = parse_key_value_lines(go_dino_filter.stdout)
        private["go"]["dino_filter"]["parsed"] = got
        want = {key: py_dino_filter[key] for key in ("dinos", "tamed", "wild", "classes")}
        cases.append(CaseResult("dino_filter", "pass" if {key: got.get(key) for key in want} == want else "fail", "dino aggregate counts compared"))

    go_property_filter = run(["go", "run", "./examples/property_filter", str(save_path), "TamerString", "Health"], repo_root, env)
    private["go"]["property_filter"] = {
        "exit_code": go_property_filter.returncode,
        "stdout": go_property_filter.stdout,
        "stderr": go_property_filter.stderr,
    }
    if go_property_filter.returncode != 0:
        cases.append(CaseResult("property_filter", "fail", "Go example exited non-zero"))
    else:
        got = parse_key_value_lines(go_property_filter.stdout)
        private["go"]["property_filter"]["parsed"] = got
        cases.append(CaseResult("property_filter", "pass" if {key: got.get(key) for key in py_property_filter} == py_property_filter else "fail", "property-name filtered object counts compared"))

    go_stackable_count = run(["go", "run", "./examples/stackable_count", str(save_path), py_stackable_count["blueprint"]], repo_root, env)
    private["go"]["stackable_count"] = {
        "exit_code": go_stackable_count.returncode,
        "stdout": go_stackable_count.stdout,
        "stderr": go_stackable_count.stderr,
    }
    if go_stackable_count.returncode != 0:
        cases.append(CaseResult("stackable_count", "fail", "Go example exited non-zero"))
    else:
        got = parse_key_value_lines(go_stackable_count.stdout)
        private["go"]["stackable_count"]["parsed"] = got
        want = {key: py_stackable_count[key] for key in ("items", "total")}
        cases.append(CaseResult("stackable_count", "pass" if {key: got.get(key) for key in want} == want else "fail", "stackable item count and total quantity compared"))

    domain_dinos_path = repo_root / ".oracle" / "output" / "export-domain-dinos.json"
    go_domain_dinos = run(["go", "run", "./cmd/arksave", "export-domain-json", str(save_path), "dinos", str(domain_dinos_path)], repo_root, env)
    private["go"]["domain_json_dinos"] = {
        "exit_code": go_domain_dinos.returncode,
        "stdout": go_domain_dinos.stdout,
        "stderr": go_domain_dinos.stderr,
        "output": str(domain_dinos_path),
    }
    if go_domain_dinos.returncode != 0:
        cases.append(CaseResult("domain_json_dinos", "fail", "Go CLI exited non-zero"))
    else:
        try:
            got_export = json.loads(domain_dinos_path.read_text(encoding="utf-8"))
            items = got_export.get("items", [])
            got = {
                "dinos": got_export.get("count"),
                "tamed": sum(1 for item in items if item.get("is_tamed")),
                "wild": sum(1 for item in items if not item.get("is_tamed")),
                "classes": len({item.get("blueprint") for item in items}),
            }
            private["go"]["domain_json_dinos"]["parsed"] = got
            want = {key: py_dino_filter[key] for key in ("dinos", "tamed", "wild", "classes")}
            cases.append(CaseResult("domain_json_dinos", "pass" if got == want else "fail", "dino domain JSON aggregate counts compared"))
        except Exception as exc:  # noqa: BLE001 - private report captures details
            private["go"]["domain_json_dinos"]["parse_error"] = str(exc)
            cases.append(CaseResult("domain_json_dinos", "fail", "Go dino domain JSON could not be parsed"))

    if py_cluster_data is None:
        cases.append(CaseResult("cluster_json", "skip", "no upstream local cluster fixture found"))
    else:
        go_cluster_json = run(["go", "run", "./examples/cluster_json", py_cluster_data["path"]], repo_root, env)
        private["go"]["cluster_json"] = {
            "exit_code": go_cluster_json.returncode,
            "stdout": go_cluster_json.stdout,
            "stderr": go_cluster_json.stderr,
        }
        if go_cluster_json.returncode != 0:
            cases.append(CaseResult("cluster_json", "fail", "Go example exited non-zero"))
        else:
            try:
                got_cluster = json.loads(go_cluster_json.stdout)
                private["go"]["cluster_json"]["parsed"] = got_cluster
                want = {
                    "id": py_cluster_data["id"],
                    "item_count": py_cluster_data["items"],
                    "dino_count": py_cluster_data["dinos"],
                }
                got = {key: got_cluster[key] for key in want}
                cases.append(CaseResult("cluster_json", "pass" if got == want else "fail", "local cluster upload counts compared"))
            except Exception as exc:  # noqa: BLE001 - private report captures details
                private["go"]["cluster_json"]["parse_error"] = str(exc)
                cases.append(CaseResult("cluster_json", "fail", "Go cluster JSON could not be parsed"))

    go_local_tribute = run(["go", "run", "./examples/local_tribute", str(save_path.parent)], repo_root, env)
    private["go"]["local_tribute"] = {
        "exit_code": go_local_tribute.returncode,
        "stdout": go_local_tribute.stdout,
        "stderr": go_local_tribute.stderr,
    }
    if go_local_tribute.returncode != 0:
        cases.append(CaseResult("local_tribute", "fail", "Go example exited non-zero"))
    else:
        got = parse_key_value_lines(go_local_tribute.stdout)
        private["go"]["local_tribute"]["parsed"] = got
        cases.append(CaseResult("local_tribute", "pass" if {key: got.get(key) for key in py_local_tribute} == py_local_tribute else "fail", "local tribute aggregate counts compared"))

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
