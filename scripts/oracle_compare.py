#!/usr/bin/env python3
"""Run privacy-safe oracle comparisons for implemented offline Go examples.

Detailed values and stdout are private and written under `.oracle/output`.
The committed markdown summary contains only case names and pass/fail/skip
statuses so it can be reviewed without leaking save-derived identifiers.
"""

from __future__ import annotations

import argparse
import contextlib
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


def oracle_env(repo_root: Path) -> dict[str, str]:
    env = os.environ.copy()
    env.setdefault("PYTHONWARNINGS", "ignore")
    env.setdefault("GOCACHE", str(repo_root / ".cache" / "go-build"))
    env.setdefault("XDG_CACHE_HOME", str(repo_root / ".cache" / "xdg"))
    os.environ.setdefault("PYTHONWARNINGS", env["PYTHONWARNINGS"])
    os.environ.setdefault("GOCACHE", env["GOCACHE"])
    os.environ.setdefault("XDG_CACHE_HOME", env["XDG_CACHE_HOME"])
    return env


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
                if value == "true":
                    values[key] = True
                elif value == "false":
                    values[key] = False
                elif "." in value:
                    values[key] = float(value)
                else:
                    values[key] = int(value)
            except ValueError:
                values[key] = value
    return values


def parse_go_player_inventory(output: str) -> dict[str, Any]:
    values = parse_key_value_lines(output)
    if "player" not in values or "items" not in values:
        raise ValueError("unexpected player_inventory output")
    values["has_location"] = "location=(" in output
    return values


def parse_go_dino_best_stat(output: str) -> dict[str, Any] | None:
    if output.strip() == "no_match":
        return None
    match = re.fullmatch(
        r"uuid=(?P<uuid>\S+) blueprint=\"(?P<blueprint>.*?)\" stat=(?P<stat>\S+) points=(?P<points>-?\d+) level=(?P<level>-?\d+)\n?",
        output,
    )
    if not match:
        raise ValueError("unexpected dino_best_stat output")
    return {
        "uuid": match.group("uuid"),
        "blueprint": match.group("blueprint"),
        "stat": match.group("stat"),
        "points": int(match.group("points")),
        "level": int(match.group("level")),
    }


def parse_go_dino_most_mutated(output: str) -> dict[str, Any] | None:
    if output.strip() == "no_match":
        return None
    values = parse_key_value_lines(output)
    if values.get("has_most_mutated") != 1 or "mutations" not in values:
        raise ValueError("unexpected dino_most_mutated output")
    return {
        "has_most_mutated": values["has_most_mutated"],
        "mutations": values["mutations"],
        "level": values.get("level", 0),
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


def python_object_summary_oracle(save_path: Path, upstream_src: Path) -> dict[str, Any] | None:
    sys.path.insert(0, str(upstream_src))
    from arkparse.saves.asa_save import AsaSave  # type: ignore

    save = AsaSave(save_path)
    try:
        for obj_uuid in save.save_connection.get_obj_uuids():
            try:
                obj = save.get_game_object_by_id(obj_uuid)
            except Exception:
                continue
            if obj is None:
                continue
            raw = save.save_connection.get_game_obj_binary(obj_uuid)
            return {
                "uuid": str(obj_uuid),
                "has_object": 1,
                "bytes": len(raw),
                "properties": len(obj.properties),
            }
        return None
    finally:
        save.close()


STORAGE_CLASS_SUBSTRINGS = [
    "StorageBox_Large_C",
    "BP_DedicatedStorage_C",
    "StorageBox_ChemBench_C",
    "StorageBox_IndustrialGrinder_C",
    "StorageBox_Fabricator_C",
    "StorageBox_TekGenerator_C",
    "StorageBox_Huge_C",
    "StorageBox_AnvilBench_C",
    "StorageBox_TekReplicator_C",
    "StorageBox_Small_C",
    "PrimalItemStructure_LibraryStorage_C",
]


def python_class_lookup_oracle(save_path: Path, upstream_src: Path) -> dict[str, Any]:
    sys.path.insert(0, str(upstream_src))
    from arkparse.api.structure_api import StructureApi  # type: ignore
    from arkparse.parsing import GameObjectReaderConfiguration  # type: ignore
    from arkparse.saves.asa_save import AsaSave  # type: ignore

    with open(os.devnull, "w", encoding="utf-8") as devnull:
        with contextlib.redirect_stdout(devnull), contextlib.redirect_stderr(devnull):
            save = AsaSave(save_path)
            try:
                config = GameObjectReaderConfiguration(
                    blueprint_name_filter=lambda name: name is not None and any(cls in name for cls in STORAGE_CLASS_SUBSTRINGS)
                )
                structures = StructureApi(save).get_all(config, max_workers=1)
                return {
                    "objects": len(structures),
                    "classes": len({structure.object.blueprint for structure in structures.values()}),
                }
            finally:
                save.close()


def python_class_property_summary_oracle(save_path: Path, upstream_src: Path) -> dict[str, Any] | None:
    sys.path.insert(0, str(upstream_src))
    from arkparse.parsing import GameObjectReaderConfiguration  # type: ignore
    from arkparse.saves.asa_save import AsaSave  # type: ignore

    with open(os.devnull, "w", encoding="utf-8") as devnull:
        with contextlib.redirect_stdout(devnull), contextlib.redirect_stderr(devnull):
            save = AsaSave(save_path)
            try:
                candidates: dict[str, int] = {}
                for class_name in save.get_all_present_classes() or []:
                    if class_name is None:
                        continue
                    candidates[class_name] = candidates.get(class_name, 0) + 1
                for class_name, _ in sorted(candidates.items(), key=lambda item: (item[1], item[0])):
                    config = GameObjectReaderConfiguration(
                        blueprint_name_filter=lambda name, selected=class_name: name == selected
                    )
                    try:
                        objects = save.get_game_objects(config)
                    except Exception:
                        continue
                    if not objects:
                        continue
                    properties = set()
                    for obj in objects.values():
                        properties.update(obj.property_names)
                    return {
                        "class_substring": class_name,
                        "objects": len(objects),
                        "properties": len(properties),
                        "faults": 0,
                    }
                return None
            finally:
                save.close()


def python_player_inventory_oracle(save_path: Path, repo_root: Path, upstream_src: Path) -> dict[str, Any] | None:
    sys.path.insert(0, str(upstream_src))
    from arkparse.api.player_api import PlayerApi  # type: ignore
    from arkparse.saves.asa_save import AsaSave  # type: ignore

    empty_cluster_dir = repo_root / ".oracle" / "output" / "empty-cluster"
    empty_cluster_dir.mkdir(parents=True, exist_ok=True)
    save = AsaSave(save_path)
    try:
        player_api = PlayerApi(
            save,
            force_legacy_store=True,
            cluster_data_dir=empty_cluster_dir,
        )
        for player in player_api.players:
            inventory = player_api.get_player_inventory(player, save)
            if inventory is None:
                continue
            item_count = getattr(inventory, "number_of_items")
            if callable(item_count):
                item_count = item_count()
            return {
                "player_id": player.id_,
                "items": int(item_count),
                "has_location": player.location is not None,
            }
        return None
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
            "total_deaths": sum(player.nr_of_deaths for player in player_api.players),
            "highest_level": max((player.stats.level for player in player_api.players), default=0),
            "unlocked_engrams": len({engram.value for player in player_api.players for engram in player.stats.engrams}),
        }
    finally:
        save.close()


def python_player_tribe_links_oracle(save_path: Path, repo_root: Path, upstream_src: Path) -> dict[str, Any]:
    sys.path.insert(0, str(upstream_src))
    from arkparse.api.player_api import PlayerApi  # type: ignore
    from arkparse.saves.asa_save import AsaSave  # type: ignore

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
        player_by_id = {player.id_: player for player in player_api.players}
        tribe_ids = {tribe.tribe_id for tribe in player_api.tribes}
        active_links = 0
        inactive_members = 0
        tribes_with_inactive = 0
        tribes_without_active = 0
        for tribe in player_api.tribes:
            active = len(player_api.tribe_to_player_map.get(tribe.tribe_id, []))
            active_links += active
            if active == 0:
                tribes_without_active += 1
            inactive = sum(1 for member_id in tribe.member_ids if member_id not in player_by_id)
            inactive_members += inactive
            if inactive > 0:
                tribes_with_inactive += 1
        return {
            "players": len(player_api.players),
            "tribes": len(player_api.tribes),
            "active_links": active_links,
            "inactive_members": inactive_members,
            "players_without_tribe": sum(1 for player in player_api.players if player.tribe not in tribe_ids),
            "tribes_with_inactive": tribes_with_inactive,
            "tribes_without_active": tribes_without_active,
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


def python_dino_best_stat_no_cryos_oracle(save_path: Path, upstream_src: Path) -> dict[str, Any] | None:
    sys.path.insert(0, str(upstream_src))
    from arkparse.api.dino_api import DinoApi  # type: ignore
    from arkparse.saves.asa_save import AsaSave  # type: ignore

    with open(os.devnull, "w", encoding="utf-8") as devnull:
        with contextlib.redirect_stdout(devnull), contextlib.redirect_stderr(devnull):
            save = AsaSave(save_path)
            try:
                dino_api = DinoApi(save)
                dinos = dino_api.get_all(include_cryos=False, max_workers=1, bypass_inventory=True)
                best_uuid = None
                best_dino = None
                best_stat = None
                best_value = None
                for uuid, dino in dinos.items():
                    stat, value = dino.stats.get_highest_stat()
                    if best_value is None or value > best_value:
                        best_uuid = uuid
                        best_dino = dino
                        best_stat = stat
                        best_value = value
                if best_dino is None:
                    return None
                return {
                    "uuid": str(best_uuid),
                    "blueprint": best_dino.object.blueprint,
                    "stat": best_stat.name.lower() if best_stat is not None else "unknown",
                    "points": int(best_value),
                    "level": int(best_dino.stats.current_level),
                }
            finally:
                save.close()


def python_dino_best_base_stat_oracle(save_path: Path, upstream_src: Path) -> dict[str, Any] | None:
    sys.path.insert(0, str(upstream_src))
    from arkparse.api.dino_api import DinoApi  # type: ignore
    from arkparse.enums import ArkStat  # type: ignore
    from arkparse.object_model.dinos.tamed_dino import TamedDino  # type: ignore
    from arkparse.saves.asa_save import AsaSave  # type: ignore

    with open(os.devnull, "w", encoding="utf-8") as devnull:
        with contextlib.redirect_stdout(devnull), contextlib.redirect_stderr(devnull):
            save = AsaSave(save_path)
            try:
                dino_api = DinoApi(save)
                dinos = dino_api.get_all(include_cryos=False, include_wild=False, include_tamed=True, max_workers=1, bypass_inventory=True)
                by_class: dict[str, list[Any]] = {}
                for dino in dinos.values():
                    if not isinstance(dino, TamedDino) or dino.stats is None:
                        continue
                    by_class.setdefault(dino.object.blueprint, []).append(dino)
                if not by_class:
                    return None
                blueprint, candidates = max(by_class.items(), key=lambda item: (len(item[1]), item[0]))
                best = max(candidates, key=lambda dino: dino.stats.get(ArkStat.WEIGHT, base=True))
                return {
                    "blueprint": blueprint,
                    "stat": "weight",
                    "has_result": 1,
                    "points": int(best.stats.get(ArkStat.WEIGHT, base=True)),
                    "level": int(best.stats.current_level),
                }
            finally:
                save.close()


def python_dino_most_mutated_oracle(save_path: Path, upstream_src: Path) -> dict[str, Any] | None:
    sys.path.insert(0, str(upstream_src))
    from arkparse.api.dino_api import DinoApi  # type: ignore
    from arkparse.saves.asa_save import AsaSave  # type: ignore

    with open(os.devnull, "w", encoding="utf-8") as devnull:
        with contextlib.redirect_stdout(devnull), contextlib.redirect_stderr(devnull):
            save = AsaSave(save_path)
            try:
                dino_api = DinoApi(save)
                dinos = dino_api.get_all_tamed()
                most_mutated = None
                best_total = None
                for dino in dinos.values():
                    total = int(dino.stats.get_total_mutations())
                    if most_mutated is None or total > best_total:
                        most_mutated = dino
                        best_total = total
                if most_mutated is None:
                    return None
                return {
                    "has_most_mutated": 1,
                    "mutations": int(best_total),
                    "level": int(most_mutated.stats.current_level),
                }
            finally:
                save.close()


def python_dino_babies_oracle(save_path: Path, upstream_src: Path) -> dict[str, Any]:
    sys.path.insert(0, str(upstream_src))
    from arkparse.api.dino_api import DinoApi  # type: ignore
    from arkparse.object_model.dinos import TamedBaby  # type: ignore
    from arkparse.saves.asa_save import AsaSave  # type: ignore

    with open(os.devnull, "w", encoding="utf-8") as devnull:
        with contextlib.redirect_stdout(devnull), contextlib.redirect_stderr(devnull):
            save = AsaSave(save_path)
            try:
                dino_api = DinoApi(save)
                babies = dino_api.get_all_babies(include_wild=True)
                tamed = sum(1 for dino in babies.values() if isinstance(dino, TamedBaby))
                return {
                    "wild_babies": len(babies) - tamed,
                    "tamed_babies": tamed,
                }
            finally:
                save.close()


def python_dino_wild_tamables_oracle(save_path: Path, upstream_src: Path) -> dict[str, Any]:
    sys.path.insert(0, str(upstream_src))
    from arkparse.api.dino_api import DinoApi  # type: ignore
    from arkparse.saves.asa_save import AsaSave  # type: ignore

    with open(os.devnull, "w", encoding="utf-8") as devnull:
        with contextlib.redirect_stdout(devnull), contextlib.redirect_stderr(devnull):
            save = AsaSave(save_path)
            try:
                dino_api = DinoApi(save)
                wild = dino_api.get_all_wild()
                tamables = dino_api.get_all_wild_tamables()
                return {
                    "wild_dinos": len(wild),
                    "wild_tamables": len(tamables),
                }
            finally:
                save.close()


def python_dino_wild_tamed_oracle(save_path: Path, upstream_src: Path) -> dict[str, Any]:
    sys.path.insert(0, str(upstream_src))
    from arkparse.api.dino_api import DinoApi  # type: ignore
    from arkparse.helpers.dino.is_wild_tamed import is_wild_tamed  # type: ignore
    from arkparse.saves.asa_save import AsaSave  # type: ignore

    with open(os.devnull, "w", encoding="utf-8") as devnull:
        with contextlib.redirect_stdout(devnull), contextlib.redirect_stderr(devnull):
            save = AsaSave(save_path)
            try:
                dinos = DinoApi(save).get_all_tamed(include_cryopodded=False)
                wild_tamed = [dino for dino in dinos.values() if is_wild_tamed(dino)]
                return {
                    "wild_tamed": len(wild_tamed),
                    "max_level": max((int(getattr(dino.stats, "current_level", 0)) for dino in wild_tamed), default=0),
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


def python_stackable_owned_by_oracle(save_path: Path, upstream_src: Path) -> dict[str, Any] | None:
    sys.path.insert(0, str(upstream_src))
    from arkparse.api.stackable_api import StackableApi  # type: ignore
    from arkparse.api.structure_api import StructureApi  # type: ignore
    from arkparse.classes.equipment import Ammo  # type: ignore
    from arkparse.saves.asa_save import AsaSave  # type: ignore

    blueprint = Ammo.advanced_rifle_bullet
    save = AsaSave(save_path)
    try:
        stackable_api = StackableApi(save)
        structure_api = StructureApi(save)
        containers: dict[Any, Any] = {}
        by_blueprint_tribe: dict[tuple[str, int], dict[str, int]] = {}

        def add_stack(stack: Any, forced_blueprint: str | None = None) -> None:
            owner_inv_uuid = getattr(stack, "owner_inv_uuid", None)
            if owner_inv_uuid is None:
                return
            if owner_inv_uuid not in containers:
                containers[owner_inv_uuid] = structure_api.get_container_of_inventory(owner_inv_uuid)
            container = containers[owner_inv_uuid]
            if container is None:
                return
            raw_tribe_id = container.owner.tribe_id
            if raw_tribe_id is None:
                return
            tribe_id = int(raw_tribe_id)
            if tribe_id == 0:
                return
            stack_blueprint = forced_blueprint or getattr(getattr(stack, "object", None), "blueprint", "")
            if not stack_blueprint:
                return
            values = by_blueprint_tribe.setdefault((stack_blueprint, tribe_id), {"items": 0, "total": 0})
            values["items"] += 1
            values["total"] += int(stack.quantity)

        for stack in stackable_api.get_by_class(StackableApi.Classes.AMMO, blueprint).values():
            add_stack(stack, blueprint)
        if not by_blueprint_tribe:
            for stack in stackable_api.get_all(StackableApi.Classes.RESOURCE, max_workers=1).values():
                add_stack(stack)
        if not by_blueprint_tribe:
            return None
        (blueprint, tribe_id), values = max(
            by_blueprint_tribe.items(),
            key=lambda item: (item[1]["total"], item[1]["items"], item[0][0], -item[0][1]),
        )
        return {
            "blueprint": blueprint,
            "tribe_id": tribe_id,
            "items": values["items"],
            "total": values["total"],
        }
    finally:
        save.close()


def python_domain_stackables_oracle(save_path: Path, upstream_src: Path) -> dict[str, Any]:
    sys.path.insert(0, str(upstream_src))
    from arkparse.api.stackable_api import StackableApi  # type: ignore
    from arkparse.saves.asa_save import AsaSave  # type: ignore

    save = AsaSave(save_path)
    try:
        api = StackableApi(save)
        resources = api.get_all(StackableApi.Classes.RESOURCE, max_workers=1)
        ammo = api.get_all(StackableApi.Classes.AMMO, max_workers=1)
        items = {**resources, **ammo}
        return {
            "stackables": len(items),
            "total_quantity": sum(int(item.quantity) for item in items.values()),
            "classes": len({item.object.blueprint for item in items.values()}),
            "owned": sum(1 for item in items.values() if getattr(item, "owner_inv_uuid", None) is not None),
        }
    finally:
        save.close()


def python_equipment_longneck_blueprint_oracle(save_path: Path, upstream_src: Path) -> dict[str, Any]:
    sys.path.insert(0, str(upstream_src))
    from arkparse.api.equipment_api import EquipmentApi  # type: ignore
    from arkparse.classes.equipment import Weapons  # type: ignore
    from arkparse.saves.asa_save import AsaSave  # type: ignore

    blueprint = Weapons.advanced.longneck
    save = AsaSave(save_path)
    try:
        api = EquipmentApi(save)
        weapons = api.get_filtered(
            EquipmentApi.Classes.WEAPON,
            classes=[blueprint],
            only_blueprints=True,
        )
        return {
            "blueprint": blueprint,
            "longneck_bp_count": len(weapons),
            "max_damage": max((weapon.damage for weapon in weapons.values()), default=None),
        }
    finally:
        save.close()


def python_equipment_best_oracle(save_path: Path, upstream_src: Path) -> dict[str, Any]:
    sys.path.insert(0, str(upstream_src))
    from arkparse.api.equipment_api import EquipmentApi  # type: ignore
    from arkparse.saves.asa_save import AsaSave  # type: ignore

    save = AsaSave(save_path)
    try:
        api = EquipmentApi(save)
        weapons = api.get_filtered(EquipmentApi.Classes.WEAPON, no_bluepints=True)
        armor = api.get_filtered(EquipmentApi.Classes.ARMOR, no_bluepints=True)
        result: dict[str, Any] = {}
        if weapons:
            best_weapon = max(weapons.values(), key=lambda item: item.damage)
            result.update(
                {
                    "weapon_damage": float(f"{best_weapon.damage:.1f}"),
                    "weapon": best_weapon.get_short_name(),
                    "weapon_crafted": bool(best_weapon.crafter),
                }
            )
        else:
            result["weapon"] = "no_match"
        if armor:
            best_armor = max(armor.values(), key=lambda item: item.durability)
            result.update(
                {
                    "armor_durability": float(f"{best_armor.durability:.1f}"),
                    "armor": best_armor.get_short_name(),
                    "armor_crafted": bool(best_armor.crafter),
                }
            )
        else:
            result["armor"] = "no_match"
        return result
    finally:
        save.close()


def python_equipment_ascendant_weapon_bps_oracle(save_path: Path, upstream_src: Path) -> dict[str, Any]:
    sys.path.insert(0, str(upstream_src))
    from arkparse.api.equipment_api import EquipmentApi  # type: ignore
    from arkparse.enums import ArkItemQuality  # type: ignore
    from arkparse.saves.asa_save import AsaSave  # type: ignore

    save = AsaSave(save_path)
    try:
        api = EquipmentApi(save)
        weapons = api.get_filtered(
            EquipmentApi.Classes.WEAPON,
            minimum_quality=ArkItemQuality.ASCENDANT,
            only_blueprints=True,
        )
        return {
            "items": len(weapons),
            "max_damage": max((float(f"{weapon.damage:.1f}") for weapon in weapons.values()), default=0.0),
        }
    finally:
        save.close()


def python_equipment_saddles_oracle(save_path: Path, upstream_src: Path) -> dict[str, Any]:
    sys.path.insert(0, str(upstream_src))
    from arkparse.api.equipment_api import EquipmentApi  # type: ignore
    from arkparse.saves.asa_save import AsaSave  # type: ignore

    save = AsaSave(save_path)
    try:
        saddles = EquipmentApi(save).get_saddles()
        return {
            "item_saddles": len(saddles),
        }
    finally:
        save.close()


def python_equipment_owned_by_oracle(save_path: Path, upstream_src: Path) -> dict[str, Any]:
    sys.path.insert(0, str(upstream_src))
    from arkparse.api.equipment_api import EquipmentApi  # type: ignore
    from arkparse.api.structure_api import StructureApi  # type: ignore
    from arkparse.classes.equipment import Weapons  # type: ignore
    from arkparse.saves.asa_save import AsaSave  # type: ignore

    save = AsaSave(save_path)
    try:
        equipment_api = EquipmentApi(save)
        structure_api = StructureApi(save)
        candidates: dict[tuple[str, int], list[Any]] = {}
        containers: dict[Any, Any] = {}
        for blueprint in Weapons.advanced.all_bps:
            weapons = equipment_api.get_filtered(
                EquipmentApi.Classes.WEAPON,
                classes=[blueprint],
                only_blueprints=True,
            )
            for weapon in weapons.values():
                owner_inv_uuid = getattr(weapon, "owner_inv_uuid", None)
                if owner_inv_uuid is None:
                    continue
                if owner_inv_uuid not in containers:
                    containers[owner_inv_uuid] = structure_api.get_container_of_inventory(owner_inv_uuid)
                container = containers[owner_inv_uuid]
                if container is None or getattr(container, "owner", None) is None:
                    continue
                tribe_id = getattr(container.owner, "tribe_id", None)
                if tribe_id is None:
                    continue
                candidates.setdefault((blueprint, int(tribe_id)), []).append(weapon)
        if not candidates:
            return {"blueprint": "", "tribe_id": 0, "items": 0, "max_damage": 0.0}
        (blueprint, tribe_id), weapons = max(
            candidates.items(),
            key=lambda item: (len(item[1]), max((weapon.damage for weapon in item[1]), default=0.0), item[0][0], item[0][1]),
        )
        return {
            "blueprint": blueprint,
            "tribe_id": tribe_id,
            "items": len(weapons),
            "max_damage": max((float(f"{weapon.damage:.1f}") for weapon in weapons), default=0.0),
        }
    finally:
        save.close()


def python_structure_owner_count_oracle(save_path: Path, upstream_src: Path) -> dict[str, Any] | None:
    sys.path.insert(0, str(upstream_src))
    from arkparse.api.structure_api import StructureApi  # type: ignore
    from arkparse.saves.asa_save import AsaSave  # type: ignore

    with open(os.devnull, "w", encoding="utf-8") as devnull:
        with contextlib.redirect_stdout(devnull), contextlib.redirect_stderr(devnull):
            save = AsaSave(save_path)
            try:
                structures = StructureApi(save).get_all(max_workers=1)
                counts: dict[int, int] = {}
                for structure in structures.values():
                    tribe_id = getattr(getattr(structure, "owner", None), "tribe_id", None)
                    if tribe_id is None or int(tribe_id) == 0:
                        continue
                    counts[int(tribe_id)] = counts.get(int(tribe_id), 0) + 1
                if not counts:
                    return None
                tribe_id, count = max(counts.items(), key=lambda item: (item[1], item[0]))
                return {
                    "tribe_id": tribe_id,
                    "structures": count,
                }
            finally:
                save.close()


def python_base_components_oracle(save_path: Path, upstream_src: Path) -> dict[str, Any]:
    sys.path.insert(0, str(upstream_src))
    from arkparse.api.structure_api import StructureApi  # type: ignore
    from arkparse.saves.asa_save import AsaSave  # type: ignore

    save = AsaSave(save_path)
    try:
        structures = StructureApi(save).get_all(max_workers=1)
        remaining = set(structures.keys())
        sizes: list[int] = []
        for start in sorted(structures.keys(), key=str):
            if start not in remaining:
                continue
            remaining.remove(start)
            stack = [start]
            size = 0
            while stack:
                current = stack.pop()
                size += 1
                for linked in structures[current].linked_structure_uuids:
                    if linked in remaining:
                        remaining.remove(linked)
                        stack.append(linked)
            sizes.append(size)
        return {
            "bases": len(sizes),
            "total_structures": sum(sizes),
            "largest": max(sizes, default=0),
            "min10": sum(1 for size in sizes if size >= 10),
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


def normalize_blueprint(value: str) -> str:
    if value.startswith("Blueprint'") and value.endswith("'"):
        return value[len("Blueprint'") : -1]
    return value


def compare(save_path: Path, repo_root: Path, upstream_src: Path) -> tuple[list[CaseResult], dict[str, Any]]:
    env = oracle_env(repo_root)
    py = python_oracle(save_path, upstream_src)
    py_object_summary = python_object_summary_oracle(save_path, upstream_src)
    py_class_lookup = python_class_lookup_oracle(save_path, upstream_src)
    py_class_property_summary = python_class_property_summary_oracle(save_path, upstream_src)
    py_local_profiles = python_local_profiles_oracle(save_path, repo_root, upstream_src)
    py_player_tribe_links = python_player_tribe_links_oracle(save_path, repo_root, upstream_src)
    py_dino_filter = python_dino_filter_oracle(save_path, upstream_src)
    py_dino_best_stat_no_cryos = python_dino_best_stat_no_cryos_oracle(save_path, upstream_src)
    py_dino_best_base_stat = python_dino_best_base_stat_oracle(save_path, upstream_src)
    py_dino_most_mutated = python_dino_most_mutated_oracle(save_path, upstream_src)
    py_dino_babies = python_dino_babies_oracle(save_path, upstream_src)
    py_dino_wild_tamables = python_dino_wild_tamables_oracle(save_path, upstream_src)
    py_dino_wild_tamed = python_dino_wild_tamed_oracle(save_path, upstream_src)
    py_property_filter = python_property_filter_oracle(save_path, upstream_src)
    py_stackable_count = python_stackable_count_oracle(save_path, upstream_src)
    py_stackable_owned_by = python_stackable_owned_by_oracle(save_path, upstream_src)
    py_domain_stackables = python_domain_stackables_oracle(save_path, upstream_src)
    py_equipment_longneck_blueprint = python_equipment_longneck_blueprint_oracle(save_path, upstream_src)
    py_equipment_best = python_equipment_best_oracle(save_path, upstream_src)
    py_equipment_ascendant_weapon_bps = python_equipment_ascendant_weapon_bps_oracle(save_path, upstream_src)
    py_equipment_saddles = python_equipment_saddles_oracle(save_path, upstream_src)
    py_equipment_owned_by = python_equipment_owned_by_oracle(save_path, upstream_src)
    py_structure_owner_count = python_structure_owner_count_oracle(save_path, upstream_src)
    py_base_components = python_base_components_oracle(save_path, upstream_src)
    py_player_inventory = python_player_inventory_oracle(save_path, repo_root, upstream_src)
    py_cluster_data = python_cluster_data_oracle(upstream_src)
    py_local_tribute = python_local_tribute_oracle(save_path)
    private: dict[str, Any] = {
        "save_path": str(save_path),
        "python": py,
        "python_object_summary": py_object_summary,
        "python_class_lookup": py_class_lookup,
        "python_class_property_summary": py_class_property_summary,
        "python_local_profiles": py_local_profiles,
        "python_player_tribe_links": py_player_tribe_links,
        "python_dino_filter": py_dino_filter,
        "python_dino_best_stat_no_cryos": py_dino_best_stat_no_cryos,
        "python_dino_best_base_stat": py_dino_best_base_stat,
        "python_dino_most_mutated": py_dino_most_mutated,
        "python_dino_babies": py_dino_babies,
        "python_dino_wild_tamables": py_dino_wild_tamables,
        "python_dino_wild_tamed": py_dino_wild_tamed,
        "python_property_filter": py_property_filter,
        "python_stackable_count": py_stackable_count,
        "python_stackable_owned_by": py_stackable_owned_by,
        "python_domain_stackables": py_domain_stackables,
        "python_equipment_longneck_blueprint": py_equipment_longneck_blueprint,
        "python_equipment_best": py_equipment_best,
        "python_equipment_ascendant_weapon_bps": py_equipment_ascendant_weapon_bps,
        "python_equipment_saddles": py_equipment_saddles,
        "python_equipment_owned_by": py_equipment_owned_by,
        "python_structure_owner_count": py_structure_owner_count,
        "python_base_components": py_base_components,
        "python_player_inventory": py_player_inventory,
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

    if py_object_summary is None:
        cases.append(CaseResult("object_summary", "skip", "oracle save has no parseable object candidate"))
    else:
        go_object_summary = run(["go", "run", "./examples/object_summary", str(save_path), str(py_object_summary["uuid"])], repo_root, env)
        private["go"]["object_summary"] = {
            "exit_code": go_object_summary.returncode,
            "stdout": go_object_summary.stdout,
            "stderr": go_object_summary.stderr,
        }
        if go_object_summary.returncode != 0:
            cases.append(CaseResult("object_summary", "fail", "Go example exited non-zero"))
        else:
            got = parse_key_value_lines(go_object_summary.stdout)
            private["go"]["object_summary"]["parsed"] = got
            want = {key: py_object_summary[key] for key in ("has_object", "bytes", "properties")}
            cases.append(CaseResult("object_summary", "pass" if {key: got.get(key) for key in want} == want else "fail", "object-by-UUID byte and property counts compared"))

    go_class_lookup = run(["go", "run", "./examples/class_lookup", str(save_path), *STORAGE_CLASS_SUBSTRINGS], repo_root, env)
    private["go"]["class_lookup"] = {
        "exit_code": go_class_lookup.returncode,
        "stdout": go_class_lookup.stdout,
        "stderr": go_class_lookup.stderr,
    }
    if go_class_lookup.returncode != 0:
        cases.append(CaseResult("class_lookup", "fail", "Go example exited non-zero"))
    else:
        got = parse_key_value_lines(go_class_lookup.stdout)
        private["go"]["class_lookup"]["parsed"] = got
        cases.append(CaseResult("class_lookup", "pass" if {key: got.get(key) for key in py_class_lookup} == py_class_lookup else "fail", "storage class substring structure counts compared"))

    if py_class_property_summary is None:
        cases.append(CaseResult("class_property_summary", "skip", "oracle save has no parseable class candidate"))
    else:
        go_class_property_summary = run([
            "go",
            "run",
            "./examples/class_property_summary",
            str(save_path),
            str(py_class_property_summary["class_substring"]),
        ], repo_root, env)
        private["go"]["class_property_summary"] = {
            "exit_code": go_class_property_summary.returncode,
            "stdout": go_class_property_summary.stdout,
            "stderr": go_class_property_summary.stderr,
        }
        if go_class_property_summary.returncode != 0:
            cases.append(CaseResult("class_property_summary", "fail", "Go example exited non-zero"))
        else:
            got = parse_key_value_lines(go_class_property_summary.stdout)
            private["go"]["class_property_summary"]["parsed"] = got
            want = {key: py_class_property_summary[key] for key in ("objects", "properties", "faults")}
            cases.append(CaseResult("class_property_summary", "pass" if {key: got.get(key) for key in want} == want else "fail", "class property-name aggregate compared"))

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
        want = {key: py_local_profiles[key] for key in ("total_deaths", "unlocked_engrams")}
        cases.append(CaseResult("local_profile_player_aggregates", "pass" if {key: got.get(key) for key in want} == want else "fail", "local player death and unlocked engram aggregates compared"))

    go_player_all = run(["go", "run", "./examples/player_all", str(save_path)], repo_root, env)
    private["go"]["player_all"] = {
        "exit_code": go_player_all.returncode,
        "stdout": go_player_all.stdout,
        "stderr": go_player_all.stderr,
    }
    if go_player_all.returncode != 0:
        cases.append(CaseResult("player_all", "fail", "Go example exited non-zero"))
    else:
        got = parse_key_value_lines(go_player_all.stdout)
        private["go"]["player_all"]["parsed"] = got
        want = {
            "players": py_local_profiles["parsed_players"],
            "tribes": py_local_profiles["parsed_tribes"],
            "highest_level": py_local_profiles["highest_level"],
            "total_deaths": py_local_profiles["total_deaths"],
            "unlocked_engrams": py_local_profiles["unlocked_engrams"],
        }
        cases.append(CaseResult("player_all", "pass" if {key: got.get(key) for key in want} == want else "fail", "save path player aggregate fallback compared"))

    go_player_tribe_links = run(["go", "run", "./examples/player_tribe_links", str(save_path)], repo_root, env)
    private["go"]["player_tribe_links"] = {
        "exit_code": go_player_tribe_links.returncode,
        "stdout": go_player_tribe_links.stdout,
        "stderr": go_player_tribe_links.stderr,
    }
    if go_player_tribe_links.returncode != 0:
        cases.append(CaseResult("player_tribe_links", "fail", "Go example exited non-zero"))
    else:
        got = parse_key_value_lines(go_player_tribe_links.stdout)
        private["go"]["player_tribe_links"]["parsed"] = got
        cases.append(CaseResult("player_tribe_links", "pass" if {key: got.get(key) for key in py_player_tribe_links} == py_player_tribe_links else "fail", "player tribe active and inactive relation aggregates compared"))

    if py_player_inventory is None:
        cases.append(CaseResult("player_inventory", "skip", "oracle save has no player inventory candidate"))
    else:
        go_player_inventory = run(["go", "run", "./examples/player_inventory", str(save_path), str(py_player_inventory["player_id"])], repo_root, env)
        private["go"]["player_inventory"] = {
            "exit_code": go_player_inventory.returncode,
            "stdout": go_player_inventory.stdout,
            "stderr": go_player_inventory.stderr,
        }
        if go_player_inventory.returncode != 0:
            cases.append(CaseResult("player_inventory", "fail", "Go example exited non-zero"))
        else:
            try:
                got = parse_go_player_inventory(go_player_inventory.stdout)
                private["go"]["player_inventory"]["parsed"] = got
                want = {key: py_player_inventory[key] for key in ("items", "has_location")}
                cases.append(CaseResult("player_inventory", "pass" if {key: got.get(key) for key in want} == want else "fail", "player inventory item count and location presence compared"))
            except Exception as exc:  # noqa: BLE001 - private report captures details
                private["go"]["player_inventory"]["parse_error"] = str(exc)
                cases.append(CaseResult("player_inventory", "fail", "Go player inventory output could not be parsed"))

    go_dino_filter = run(["go", "run", "./examples/dino_filter", "--no-cryos", str(save_path)], repo_root, env)
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
        want["cryopodded"] = 0
        cases.append(CaseResult("dino_filter", "pass" if {key: got.get(key) for key in want} == want else "fail", "dino aggregate counts compared"))

    if py_dino_best_stat_no_cryos is None:
        cases.append(CaseResult("dino_best_stat_no_cryos", "skip", "oracle save has no non-cryopod stat-bearing dino candidate"))
    else:
        go_dino_best_stat = run(["go", "run", "./examples/dino_best_stat", "--no-cryos", str(save_path)], repo_root, env)
        private["go"]["dino_best_stat_no_cryos"] = {
            "exit_code": go_dino_best_stat.returncode,
            "stdout": go_dino_best_stat.stdout,
            "stderr": go_dino_best_stat.stderr,
        }
        if go_dino_best_stat.returncode != 0:
            cases.append(CaseResult("dino_best_stat_no_cryos", "fail", "Go example exited non-zero"))
        else:
            try:
                got = parse_go_dino_best_stat(go_dino_best_stat.stdout)
                private["go"]["dino_best_stat_no_cryos"]["parsed"] = got
                cases.append(CaseResult("dino_best_stat_no_cryos", "pass" if got == py_dino_best_stat_no_cryos else "fail", "best stat dino without cryopods compared"))
            except Exception as exc:  # noqa: BLE001 - private report captures details
                private["go"]["dino_best_stat_no_cryos"]["parse_error"] = str(exc)
                cases.append(CaseResult("dino_best_stat_no_cryos", "fail", "Go dino_best_stat output could not be parsed"))

    if py_dino_best_base_stat is None:
        cases.append(CaseResult("dino_best_base_stat", "skip", "oracle save has no direct tamed stat-bearing dino candidate"))
    else:
        go_dino_best_base_stat = run([
            "go",
            "run",
            "./examples/dino_best_base_stat",
            str(save_path),
            str(py_dino_best_base_stat["blueprint"]),
            str(py_dino_best_base_stat["stat"]),
        ], repo_root, env)
        private["go"]["dino_best_base_stat"] = {
            "exit_code": go_dino_best_base_stat.returncode,
            "stdout": go_dino_best_base_stat.stdout,
            "stderr": go_dino_best_base_stat.stderr,
        }
        if go_dino_best_base_stat.returncode != 0:
            cases.append(CaseResult("dino_best_base_stat", "fail", "Go example exited non-zero"))
        else:
            got = parse_key_value_lines(go_dino_best_base_stat.stdout)
            private["go"]["dino_best_base_stat"]["parsed"] = got
            want = {key: py_dino_best_base_stat[key] for key in ("has_result", "stat", "points", "level")}
            cases.append(CaseResult("dino_best_base_stat", "pass" if {key: got.get(key) for key in want} == want else "fail", "class-filtered tamed base stat dino compared"))

    if py_dino_most_mutated is None:
        cases.append(CaseResult("dino_most_mutated", "skip", "oracle save has no tamed mutation-bearing dino candidate"))
    else:
        go_dino_most_mutated = run(["go", "run", "./examples/dino_most_mutated", str(save_path)], repo_root, env)
        private["go"]["dino_most_mutated"] = {
            "exit_code": go_dino_most_mutated.returncode,
            "stdout": go_dino_most_mutated.stdout,
            "stderr": go_dino_most_mutated.stderr,
        }
        if go_dino_most_mutated.returncode != 0:
            cases.append(CaseResult("dino_most_mutated", "fail", "Go example exited non-zero"))
        else:
            try:
                got = parse_go_dino_most_mutated(go_dino_most_mutated.stdout)
                private["go"]["dino_most_mutated"]["parsed"] = got
                cases.append(CaseResult("dino_most_mutated", "pass" if got == py_dino_most_mutated else "fail", "most mutated tamed dino aggregate compared"))
            except Exception as exc:  # noqa: BLE001 - private report captures details
                private["go"]["dino_most_mutated"]["parse_error"] = str(exc)
                cases.append(CaseResult("dino_most_mutated", "fail", "Go dino_most_mutated output could not be parsed"))

    go_dino_babies = run(["go", "run", "./examples/dino_babies", str(save_path)], repo_root, env)
    private["go"]["dino_babies"] = {
        "exit_code": go_dino_babies.returncode,
        "stdout": go_dino_babies.stdout,
        "stderr": go_dino_babies.stderr,
    }
    if go_dino_babies.returncode != 0:
        cases.append(CaseResult("dino_babies", "fail", "Go example exited non-zero"))
    else:
        got = parse_key_value_lines(go_dino_babies.stdout)
        private["go"]["dino_babies"]["parsed"] = got
        cases.append(CaseResult("dino_babies", "pass" if {key: got.get(key) for key in py_dino_babies} == py_dino_babies else "fail", "wild and tamed baby dino counts compared"))

    go_dino_wild_tamables = run(["go", "run", "./examples/dino_wild_tamables", str(save_path)], repo_root, env)
    private["go"]["dino_wild_tamables"] = {
        "exit_code": go_dino_wild_tamables.returncode,
        "stdout": go_dino_wild_tamables.stdout,
        "stderr": go_dino_wild_tamables.stderr,
    }
    if go_dino_wild_tamables.returncode != 0:
        cases.append(CaseResult("dino_wild_tamables", "fail", "Go example exited non-zero"))
    else:
        got = parse_key_value_lines(go_dino_wild_tamables.stdout)
        private["go"]["dino_wild_tamables"]["parsed"] = got
        cases.append(CaseResult("dino_wild_tamables", "pass" if {key: got.get(key) for key in py_dino_wild_tamables} == py_dino_wild_tamables else "fail", "wild and tameable dino counts compared"))

    go_dino_wild_tamed = run(["go", "run", "./examples/dino_wild_tamed", str(save_path)], repo_root, env)
    private["go"]["dino_wild_tamed"] = {
        "exit_code": go_dino_wild_tamed.returncode,
        "stdout": go_dino_wild_tamed.stdout,
        "stderr": go_dino_wild_tamed.stderr,
    }
    if go_dino_wild_tamed.returncode != 0:
        cases.append(CaseResult("dino_wild_tamed", "fail", "Go example exited non-zero"))
    else:
        got = parse_key_value_lines(go_dino_wild_tamed.stdout)
        private["go"]["dino_wild_tamed"]["parsed"] = got
        cases.append(CaseResult("dino_wild_tamed", "pass" if {key: got.get(key) for key in py_dino_wild_tamed} == py_dino_wild_tamed else "fail", "wild-tamed dino count and max level compared"))

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

    if py_stackable_owned_by is None:
        cases.append(CaseResult("stackable_owned_by", "skip", "oracle save has no owned advanced rifle bullet stacks"))
    else:
        go_stackable_owned_by = run(
            [
                "go",
                "run",
                "./examples/stackable_owned_by",
                str(save_path),
                py_stackable_owned_by["blueprint"],
                str(py_stackable_owned_by["tribe_id"]),
            ],
            repo_root,
            env,
        )
        private["go"]["stackable_owned_by"] = {
            "exit_code": go_stackable_owned_by.returncode,
            "stdout": go_stackable_owned_by.stdout,
            "stderr": go_stackable_owned_by.stderr,
        }
        if go_stackable_owned_by.returncode != 0:
            cases.append(CaseResult("stackable_owned_by", "fail", "Go example exited non-zero"))
        else:
            got = parse_key_value_lines(go_stackable_owned_by.stdout)
            private["go"]["stackable_owned_by"]["parsed"] = got
            want = {key: py_stackable_owned_by[key] for key in ("tribe_id", "items", "total")}
            cases.append(CaseResult("stackable_owned_by", "pass" if {key: got.get(key) for key in want} == want else "fail", "owned advanced rifle bullet count and total compared"))

    domain_stackables_path = repo_root / ".oracle" / "output" / "export-domain-stackables.json"
    go_domain_stackables = run(["go", "run", "./cmd/arksave", "export-domain-json", str(save_path), "stackables", str(domain_stackables_path)], repo_root, env)
    private["go"]["domain_json_stackables"] = {
        "exit_code": go_domain_stackables.returncode,
        "stdout": go_domain_stackables.stdout,
        "stderr": go_domain_stackables.stderr,
        "output": str(domain_stackables_path),
    }
    if go_domain_stackables.returncode != 0:
        cases.append(CaseResult("domain_json_stackables", "fail", "Go CLI exited non-zero"))
    else:
        try:
            got_export = json.loads(domain_stackables_path.read_text(encoding="utf-8"))
            items = got_export.get("items", [])
            got = {
                "stackables": got_export.get("count"),
                "total_quantity": sum(int(item.get("quantity") or 0) for item in items),
                "classes": len({item.get("blueprint") for item in items}),
                "owned": sum(1 for item in items if item.get("owner_inventory_uuid")),
            }
            private["go"]["domain_json_stackables"]["parsed"] = got
            cases.append(CaseResult("domain_json_stackables", "pass" if got == py_domain_stackables else "fail", "stackable domain JSON aggregate counts compared"))
        except Exception as exc:  # noqa: BLE001 - private report captures details
            private["go"]["domain_json_stackables"]["parse_error"] = str(exc)
            cases.append(CaseResult("domain_json_stackables", "fail", "Go stackable domain JSON could not be parsed"))

    equipment_json_path = repo_root / ".oracle" / "output" / "export-domain-equipment.json"
    go_equipment_json = run(["go", "run", "./cmd/arksave", "export-domain-json", str(save_path), "equipment", str(equipment_json_path)], repo_root, env)
    private["go"]["equipment_longneck_blueprint_damage"] = {
        "exit_code": go_equipment_json.returncode,
        "stdout": go_equipment_json.stdout,
        "stderr": go_equipment_json.stderr,
        "output": str(equipment_json_path),
    }
    if py_equipment_longneck_blueprint["longneck_bp_count"] == 0:
        cases.append(CaseResult("equipment_longneck_blueprint_damage", "skip", "oracle save has no longneck blueprints"))
    elif go_equipment_json.returncode != 0:
        cases.append(CaseResult("equipment_longneck_blueprint_damage", "fail", "Go CLI exited non-zero"))
    else:
        try:
            got_export = json.loads(equipment_json_path.read_text(encoding="utf-8"))
            blueprint = py_equipment_longneck_blueprint["blueprint"]
            items = [
                item
                for item in got_export.get("items", [])
                if normalize_blueprint(item.get("blueprint", "")) == blueprint
                and item.get("kind") == "weapon"
                and item.get("is_blueprint")
            ]
            got = {
                "longneck_bp_count": len(items),
                "max_damage": max((float((item.get("stats") or {}).get("damage") or 0) for item in items), default=None),
            }
            private["go"]["equipment_longneck_blueprint_damage"]["parsed"] = got
            want = {key: py_equipment_longneck_blueprint[key] for key in ("longneck_bp_count", "max_damage")}
            cases.append(CaseResult("equipment_longneck_blueprint_damage", "pass" if got == want else "fail", "longneck blueprint count and max damage compared"))
        except Exception as exc:  # noqa: BLE001 - private report captures details
            private["go"]["equipment_longneck_blueprint_damage"]["parse_error"] = str(exc)
            cases.append(CaseResult("equipment_longneck_blueprint_damage", "fail", "Go equipment domain JSON could not be parsed"))

    go_equipment_best = run(["go", "run", "./examples/equipment_best", str(save_path)], repo_root, env)
    private["go"]["equipment_best"] = {
        "exit_code": go_equipment_best.returncode,
        "stdout": go_equipment_best.stdout,
        "stderr": go_equipment_best.stderr,
    }
    if go_equipment_best.returncode != 0:
        cases.append(CaseResult("equipment_best", "fail", "Go example exited non-zero"))
    else:
        got = parse_key_value_lines(go_equipment_best.stdout)
        private["go"]["equipment_best"]["parsed"] = got
        stable_keys = [key for key in py_equipment_best if key not in {"weapon", "armor"}]
        cases.append(CaseResult("equipment_best", "pass" if {key: got.get(key) for key in stable_keys} == {key: py_equipment_best[key] for key in stable_keys} else "fail", "highest weapon damage and armor durability values compared"))

    go_equipment_ascendant_weapon_bps = run(["go", "run", "./examples/equipment_ascendant_weapon_bps", str(save_path)], repo_root, env)
    private["go"]["equipment_ascendant_weapon_bps"] = {
        "exit_code": go_equipment_ascendant_weapon_bps.returncode,
        "stdout": go_equipment_ascendant_weapon_bps.stdout,
        "stderr": go_equipment_ascendant_weapon_bps.stderr,
    }
    if go_equipment_ascendant_weapon_bps.returncode != 0:
        cases.append(CaseResult("equipment_ascendant_weapon_bps", "fail", "Go example exited non-zero"))
    else:
        got = parse_key_value_lines(go_equipment_ascendant_weapon_bps.stdout)
        private["go"]["equipment_ascendant_weapon_bps"]["parsed"] = got
        cases.append(CaseResult("equipment_ascendant_weapon_bps", "pass" if {key: got.get(key) for key in py_equipment_ascendant_weapon_bps} == py_equipment_ascendant_weapon_bps else "fail", "ascendant weapon blueprint count and max damage compared"))

    go_equipment_saddles = run(["go", "run", "./examples/equipment_saddles", str(save_path)], repo_root, env)
    private["go"]["equipment_saddles"] = {
        "exit_code": go_equipment_saddles.returncode,
        "stdout": go_equipment_saddles.stdout,
        "stderr": go_equipment_saddles.stderr,
    }
    if go_equipment_saddles.returncode != 0:
        cases.append(CaseResult("equipment_saddles", "fail", "Go example exited non-zero"))
    else:
        got = parse_key_value_lines(go_equipment_saddles.stdout)
        private["go"]["equipment_saddles"]["parsed"] = got
        cases.append(CaseResult("equipment_saddles", "pass" if {key: got.get(key) for key in py_equipment_saddles} == py_equipment_saddles else "fail", "direct saddle count compared; upstream cryopod saddle extraction blocked by malformed private cryopods and armor-value parity needs default armor tables"))

    if py_equipment_owned_by["items"] == 0:
        cases.append(CaseResult("equipment_owned_by", "skip", "oracle save has no owned advanced weapon blueprints"))
    else:
        go_equipment_owned_by = run([
            "go",
            "run",
            "./examples/equipment_owned_by",
            str(save_path),
            py_equipment_owned_by["blueprint"],
            str(py_equipment_owned_by["tribe_id"]),
        ], repo_root, env)
        private["go"]["equipment_owned_by"] = {
            "exit_code": go_equipment_owned_by.returncode,
            "stdout": go_equipment_owned_by.stdout,
            "stderr": go_equipment_owned_by.stderr,
        }
        if go_equipment_owned_by.returncode != 0:
            cases.append(CaseResult("equipment_owned_by", "fail", "Go example exited non-zero"))
        else:
            got = parse_key_value_lines(go_equipment_owned_by.stdout)
            private["go"]["equipment_owned_by"]["parsed"] = got
            want = {key: py_equipment_owned_by[key] for key in ("tribe_id", "items", "max_damage")}
            cases.append(CaseResult("equipment_owned_by", "pass" if {key: got.get(key) for key in want} == want else "fail", "owned advanced weapon blueprint count and max damage compared"))

    if py_structure_owner_count is None:
        cases.append(CaseResult("structure_owner_count", "skip", "oracle save has no owned structures with nonzero tribe IDs"))
    else:
        go_structure_owner_count = run([
            "go",
            "run",
            "./examples/structure_owner_count",
            str(save_path),
            str(py_structure_owner_count["tribe_id"]),
        ], repo_root, env)
        private["go"]["structure_owner_count"] = {
            "exit_code": go_structure_owner_count.returncode,
            "stdout": go_structure_owner_count.stdout,
            "stderr": go_structure_owner_count.stderr,
        }
        if go_structure_owner_count.returncode != 0:
            cases.append(CaseResult("structure_owner_count", "fail", "Go example exited non-zero"))
        else:
            got = parse_key_value_lines(go_structure_owner_count.stdout)
            private["go"]["structure_owner_count"]["parsed"] = got
            cases.append(CaseResult("structure_owner_count", "pass" if {key: got.get(key) for key in py_structure_owner_count} == py_structure_owner_count else "fail", "owned structure count compared"))

    go_base_components = run(["go", "run", "./examples/base_components", str(save_path)], repo_root, env)
    private["go"]["base_components"] = {
        "exit_code": go_base_components.returncode,
        "stdout": go_base_components.stdout,
        "stderr": go_base_components.stderr,
    }
    if go_base_components.returncode != 0:
        cases.append(CaseResult("base_components", "fail", "Go example exited non-zero"))
    else:
        got = parse_key_value_lines(go_base_components.stdout)
        private["go"]["base_components"]["parsed"] = got
        want = {key: py_base_components[key] for key in ("bases", "total_structures", "largest", "min10")}
        cases.append(CaseResult("base_components", "pass" if {key: got.get(key) for key in want} == want else "fail", "connected base component aggregate counts compared"))

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
