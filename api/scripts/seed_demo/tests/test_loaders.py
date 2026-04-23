"""Schema-shape tests for seed_demo.corpus.loaders.

These exercise the LOADER boundary — every schema violation a real
fixture author might produce. Semantic checks (catalog membership,
cross-references, slug uniqueness) live in `test_validator.py`.
"""

from __future__ import annotations

from copy import deepcopy

import pytest
import yaml

from seed_demo.corpus.loaders import (
    SchemaError,
    load_file_entry,
    load_files_from_yaml,
    load_resource_entry,
    load_resources_from_yaml,
)

# ---------------------------------------------------------------------------
# load_file_entry — required keys
# ---------------------------------------------------------------------------


@pytest.mark.parametrize(
    "missing_key",
    ["slug", "source_url", "mime_type", "filename", "license", "owner_role"],
)
def test_load_file_missing_required_key_raises(valid_file_dict, missing_key):
    bad = deepcopy(valid_file_dict)
    del bad[missing_key]
    with pytest.raises(SchemaError, match=missing_key):
        load_file_entry(bad)


def test_load_file_missing_license_id_raises(valid_file_dict):
    bad = deepcopy(valid_file_dict)
    del bad["license"]["id"]
    with pytest.raises(SchemaError, match="id"):
        load_file_entry(bad)


def test_load_file_missing_license_attribution_raises(valid_file_dict):
    bad = deepcopy(valid_file_dict)
    del bad["license"]["attribution"]
    with pytest.raises(SchemaError, match="attribution"):
        load_file_entry(bad)


# ---------------------------------------------------------------------------
# load_file_entry — wrong types
# ---------------------------------------------------------------------------


def test_load_file_non_dict_entry_raises():
    with pytest.raises(SchemaError, match="mapping"):
        load_file_entry(["not", "a", "dict"])  # type: ignore[arg-type]


def test_load_file_non_dict_license_raises(valid_file_dict):
    bad = deepcopy(valid_file_dict)
    bad["license"] = "CC-BY-4.0"  # common author mistake
    with pytest.raises(SchemaError, match="license"):
        load_file_entry(bad)


def test_load_file_non_string_slug_raises(valid_file_dict):
    bad = deepcopy(valid_file_dict)
    bad["slug"] = 12345
    with pytest.raises(SchemaError, match="string"):
        load_file_entry(bad)


def test_load_file_non_list_courses_raises(valid_file_dict):
    bad = deepcopy(valid_file_dict)
    bad["attached_to"]["courses"] = "wsu/cpts121"  # forgot the list wrapper
    with pytest.raises(SchemaError, match="list"):
        load_file_entry(bad)


def test_load_file_owner_seed_index_must_be_int(valid_file_dict):
    bad = deepcopy(valid_file_dict)
    bad["owner_seed_index"] = "42"
    with pytest.raises(SchemaError, match="int"):
        load_file_entry(bad)


def test_load_file_owner_seed_index_bool_rejected(valid_file_dict):
    """bool is a subclass of int in Python — guard explicitly."""
    bad = deepcopy(valid_file_dict)
    bad["owner_seed_index"] = True
    with pytest.raises(SchemaError, match="int"):
        load_file_entry(bad)


# ---------------------------------------------------------------------------
# load_file_entry — empty / whitespace strings (HIGH-4 fix from review)
# ---------------------------------------------------------------------------


@pytest.mark.parametrize("key", ["slug", "source_url", "mime_type", "filename"])
def test_load_file_empty_string_rejected(valid_file_dict, key):
    bad = deepcopy(valid_file_dict)
    bad[key] = ""
    with pytest.raises(SchemaError, match="non-empty"):
        load_file_entry(bad)


def test_load_file_whitespace_only_slug_rejected(valid_file_dict):
    bad = deepcopy(valid_file_dict)
    bad["slug"] = "   "
    with pytest.raises(SchemaError, match="non-empty"):
        load_file_entry(bad)


# ---------------------------------------------------------------------------
# load_resource_entry parity
# ---------------------------------------------------------------------------


@pytest.mark.parametrize("missing_key", ["slug", "title", "url", "type", "owner_role"])
def test_load_resource_missing_required_key_raises(valid_resource_dict, missing_key):
    bad = deepcopy(valid_resource_dict)
    del bad[missing_key]
    with pytest.raises(SchemaError, match=missing_key):
        load_resource_entry(bad)


def test_load_resource_optional_description_can_be_null(valid_resource_dict):
    bad = deepcopy(valid_resource_dict)
    bad["description"] = None
    entry = load_resource_entry(bad)
    assert entry.description is None


def test_load_resource_description_wrong_type_rejected(valid_resource_dict):
    bad = deepcopy(valid_resource_dict)
    bad["description"] = ["not", "a", "string"]
    with pytest.raises(SchemaError, match="string"):
        load_resource_entry(bad)


# ---------------------------------------------------------------------------
# YAML file loaders
# ---------------------------------------------------------------------------


def test_load_files_from_yaml_top_level_must_be_list(tmp_path):
    p = tmp_path / "files.yaml"
    p.write_text(yaml.safe_dump({"slug": "wrong-shape"}))
    with pytest.raises(SchemaError, match="list"):
        load_files_from_yaml(p)


def test_load_resources_from_yaml_top_level_must_be_list(tmp_path):
    p = tmp_path / "resources.yaml"
    p.write_text(yaml.safe_dump({"slug": "wrong-shape"}))
    with pytest.raises(SchemaError, match="list"):
        load_resources_from_yaml(p)


def test_load_files_from_yaml_empty_file_returns_empty(tmp_path):
    p = tmp_path / "files.yaml"
    p.write_text("")
    assert load_files_from_yaml(p) == []


def test_load_files_from_yaml_round_trips_smoke_fixture(tmp_path, valid_file_dict):
    p = tmp_path / "files.yaml"
    p.write_text(yaml.safe_dump([valid_file_dict]))
    files = load_files_from_yaml(p)
    assert len(files) == 1
    assert files[0].slug == valid_file_dict["slug"]
