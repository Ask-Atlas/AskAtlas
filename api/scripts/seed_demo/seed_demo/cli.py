"""CLI entry-point for `python -m seed_demo`.

Exit codes:
    0 — success
    1 — schema / validation failure
    2 — liveness failure (separable so schema-only CI can ignore it)
    3 — internal error
"""

from __future__ import annotations

import argparse
import sys
from pathlib import Path

from .corpus.attributions import write_attributions_json
from .corpus.liveness import DEFAULT_PARALLEL, check_urls
from .corpus.loaders import (
    SchemaError,
    load_files_from_yaml,
    load_resources_from_yaml,
)
from .corpus.validator import validate_corpus

DEFAULT_FIXTURES_DIR = Path("fixtures")
DEFAULT_ATTRIBUTIONS_OUT = Path("data/attributions.json")


def _log(msg: str) -> None:
    print(f"[seed_demo.validate] {msg}")


def _cmd_validate(args: argparse.Namespace) -> int:
    fixtures_dir = Path(args.fixtures_dir)
    files_path = fixtures_dir / "files.yaml"
    resources_path = fixtures_dir / "resources.yaml"

    _log(f"loading {files_path}")
    try:
        files = load_files_from_yaml(files_path)
    except (FileNotFoundError, SchemaError) as exc:
        _log(f"failed to load files: {exc}")
        return 1
    _log(f"loaded {len(files)} file entries")

    _log(f"loading {resources_path}")
    try:
        resources = load_resources_from_yaml(resources_path)
    except (FileNotFoundError, SchemaError) as exc:
        _log(f"failed to load resources: {exc}")
        return 1
    _log(f"loaded {len(resources)} resource entries")

    report = validate_corpus(
        files,
        resources,
        enforce_coverage_gate=not args.no_coverage_gate,
    )

    for err in report.errors:
        _log(f"ERROR: {err}")
    for warn in report.warnings:
        _log(f"WARN:  {warn}")

    if report.errors:
        _log(f"FAIL — {len(report.errors)} error(s), {len(report.warnings)} warning(s)")
        return 1
    _log(f"schema + cross-reference: OK ({len(report.warnings)} warnings)")

    if args.check_urls:
        _log(f"URL liveness check ({len(files)} URLs, parallel={args.parallel})...")
        results = check_urls(files, parallel=args.parallel)
        failures = [r for r in results.values() if not r.ok]
        for r in failures:
            _log(f"  URL FAIL: {r.slug} — {r.error}")
        if failures:
            _log(f"liveness: {len(failures)}/{len(results)} FAILED")
            return 2
        _log(f"liveness: OK ({len(results)}/{len(results)})")

    attributions_out = Path(args.attributions_out)
    write_attributions_json(files, attributions_out)
    _log(f"wrote {attributions_out}")

    _log("PASS")
    return 0


def _build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(
        prog="python -m seed_demo",
        description="AskAtlas demo seed — corpus validator + attribution generator.",
    )
    sub = parser.add_subparsers(dest="command", required=True)

    v = sub.add_parser("validate", help="Validate fixture YAML files.")
    v.add_argument(
        "--fixtures-dir",
        default=str(DEFAULT_FIXTURES_DIR),
        help=f"Directory holding files.yaml + resources.yaml (default: {DEFAULT_FIXTURES_DIR})",
    )
    v.add_argument(
        "--attributions-out",
        default=str(DEFAULT_ATTRIBUTIONS_OUT),
        help=f"Where to write the attributions JSON (default: {DEFAULT_ATTRIBUTIONS_OUT})",
    )
    v.add_argument(
        "--check-urls",
        action="store_true",
        help="Opt-in HTTP liveness check of every source_url (~30s for full corpus).",
    )
    v.add_argument(
        "--parallel",
        type=int,
        default=DEFAULT_PARALLEL,
        help=f"Concurrent URL checks when --check-urls is set (default: {DEFAULT_PARALLEL}).",
    )
    v.add_argument(
        "--no-coverage-gate",
        action="store_true",
        help="Disable MIME-type coverage warnings (useful while iterating).",
    )

    return parser


def main(argv: list[str] | None = None) -> int:
    parser = _build_parser()
    args = parser.parse_args(argv)

    if args.command == "validate":
        return _cmd_validate(args)

    parser.print_help()
    return 3


if __name__ == "__main__":
    sys.exit(main())
