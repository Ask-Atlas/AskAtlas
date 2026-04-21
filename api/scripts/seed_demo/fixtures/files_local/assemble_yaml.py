"""Assemble fixtures/files.yaml from Phase-1b sources.

Walks fixtures/files_local/sources/**/*.md, parses frontmatter, and emits a
file entry per source pointing at the pseudo-host that resolves to the
generated file in fixtures/files_local/generated/.

Then appends a curated list of verified public-source and Unsplash entries.

Run manually:
    cd api/scripts/seed_demo
    uv run python fixtures/files_local/assemble_yaml.py
"""

from __future__ import annotations

import re
from pathlib import Path

import yaml

HERE = Path(__file__).resolve().parent
SOURCES_DIR = HERE / "sources"
ROOT = HERE.parent.parent  # api/scripts/seed_demo
FILES_YAML = ROOT / "fixtures" / "files.yaml"

PSEUDO_HOST = "https://files-local.askatlas-demo.example"


def parse_frontmatter(md: str) -> dict:
    m = re.match(r"^---\s*\n(.*?)\n---\s*\n", md, flags=re.DOTALL)
    if not m:
        raise ValueError("no frontmatter")
    return yaml.safe_load(m.group(1))


def self_generated_entries() -> list[dict]:
    out = []
    for md_path in sorted(SOURCES_DIR.rglob("*.md")):
        fm = parse_frontmatter(md_path.read_text())
        entry = {
            "slug": fm["slug"],
            "source_url": f"{PSEUDO_HOST}/{fm['filename']}",
            "mime_type": fm["mime"],
            "filename": fm["filename"],
            "license": {
                "id": "MIT",
                "attribution": f"AskAtlas Demo Seed — {fm['title']} (2026)",
            },
            "attached_to": {
                "courses": [fm["course"]],
                "study_guides": [],
            },
            "owner_role": "bot",
        }
        out.append(entry)
    return out


# ---------------------------------------------------------------------------
# Verified public-source entries — URLs curl'd 200 with correct Content-Type
# during the Phase-1b curation session. Licenses looked up per-file.
# ---------------------------------------------------------------------------

PUBLIC_SOURCE_ENTRIES: list[dict] = [
    # Phase 1a smoke fixture entries (preserved)
    {
        "slug": "smoke-wikimedia-png-transparency-demo",
        "source_url": "https://upload.wikimedia.org/wikipedia/commons/4/47/PNG_transparency_demonstration_1.png",
        "mime_type": "image/png",
        "filename": "png-transparency-demo.png",
        "license": {
            "id": "PUBLIC-DOMAIN",
            "attribution": "Wikimedia Commons — PNG transparency demonstration (public domain)",
        },
        "attached_to": {"courses": ["stanford/cs106a"], "study_guides": []},
        "owner_role": "bot",
    },
    {
        "slug": "smoke-gutenberg-pride-and-prejudice-epub",
        "source_url": "https://www.gutenberg.org/ebooks/1342.epub.noimages",
        "mime_type": "application/epub+zip",
        "filename": "pride-and-prejudice.epub",
        "license": {
            "id": "PUBLIC-DOMAIN",
            "attribution": "Jane Austen, Pride and Prejudice — Project Gutenberg #1342",
        },
        "attached_to": {"courses": ["stanford/hist1b"], "study_guides": []},
        "owner_role": "bot",
    },
    {
        "slug": "smoke-unsplash-tech-desk-jpeg",
        "source_url": "https://images.unsplash.com/photo-1517694712202-14dd9538aa97?w=1200&q=80",
        "mime_type": "image/jpeg",
        "filename": "unsplash-tech-desk.jpg",
        "license": {
            "id": "UNSPLASH",
            "attribution": "Photo by Nicolas Hoizey on Unsplash",
        },
        "attached_to": {"courses": ["wsu/cpts121"], "study_guides": []},
        "owner_role": "bot",
    },
    {
        "slug": "smoke-wikibooks-calculus-pdf",
        "source_url": "https://upload.wikimedia.org/wikipedia/commons/a/a3/Calculus.pdf",
        "mime_type": "application/pdf",
        "filename": "wikibooks-calculus.pdf",
        "license": {
            "id": "CC-BY-SA-3.0",
            "attribution": "English Wikibook on Calculus community — Calculus (PDF by Dirk Hünniger, 2011)",
        },
        "attached_to": {
            "courses": ["wsu/math171", "stanford/math51"],
            "study_guides": [],
        },
        "owner_role": "bot",
    },
    # --- Wikimedia Commons direct-upload PDFs (humanities) ---
    {
        "slug": "wikimedia-burckhardt-renaissance-italy-pdf",
        "source_url": "https://upload.wikimedia.org/wikipedia/commons/f/f6/The_civilisation_of_the_period_of_the_renaissance_in_Italy_(IA_civilisationofpe01burc).pdf",
        "mime_type": "application/pdf",
        "filename": "burckhardt-civilisation-renaissance-italy.pdf",
        "license": {
            "id": "PUBLIC-DOMAIN",
            "attribution": "Jacob Burckhardt, The Civilisation of the Renaissance in Italy (1860)",
        },
        "attached_to": {
            "courses": ["stanford/hist1b", "wsu/hist105"],
            "study_guides": [],
        },
        "owner_role": "bot",
    },
    {
        "slug": "wikimedia-renaissance-court-crystal-palace-pdf",
        "source_url": "https://upload.wikimedia.org/wikipedia/commons/9/97/The_Renaissance_court_in_the_Crystal_Palace_(IA_renaissancecourt00wyat).pdf",
        "mime_type": "application/pdf",
        "filename": "renaissance-court-crystal-palace.pdf",
        "license": {
            "id": "PUBLIC-DOMAIN",
            "attribution": "Matthew Digby Wyatt, The Renaissance Court in the Crystal Palace (1854)",
        },
        "attached_to": {"courses": ["stanford/hist1b"], "study_guides": []},
        "owner_role": "bot",
    },
    {
        "slug": "wikimedia-reformation-italy-pdf",
        "source_url": "https://upload.wikimedia.org/wikipedia/commons/a/a3/History_of_the_progress_and_suppression_of_the_reformation_in_Italy_in_the_16th_century_(microform)_-_including_a_sketch_of_the_history_of_the_Reformation_in_the_Grisons_(IA_historyofprogres01mcri).pdf",
        "mime_type": "application/pdf",
        "filename": "reformation-italy-16th-century.pdf",
        "license": {
            "id": "PUBLIC-DOMAIN",
            "attribution": "Thomas McCrie, History of the Reformation in Italy in the 16th Century (1827)",
        },
        "attached_to": {
            "courses": ["wsu/hist105", "stanford/hist1b"],
            "study_guides": [],
        },
        "owner_role": "bot",
    },
    {
        "slug": "wikimedia-renaissance-age-of-despots-pdf",
        "source_url": "https://upload.wikimedia.org/wikipedia/commons/3/3c/Renaissance_in_Italy-_the_age_of_the_despots_(IA_renaissanceinita00symoiala).pdf",
        "mime_type": "application/pdf",
        "filename": "symonds-renaissance-age-despots.pdf",
        "license": {
            "id": "PUBLIC-DOMAIN",
            "attribution": "John Addington Symonds, Renaissance in Italy: The Age of the Despots (1875)",
        },
        "attached_to": {"courses": ["stanford/hist1b"], "study_guides": []},
        "owner_role": "bot",
    },
    # --- Wikimedia Commons direct PDFs (CS / study skills) ---
    {
        "slug": "wikimedia-modern-c-gustedt-pdf",
        "source_url": "https://upload.wikimedia.org/wikipedia/commons/0/0a/Modern_C.pdf",
        "mime_type": "application/pdf",
        "filename": "modern-c-gustedt.pdf",
        "license": {
            "id": "CC-BY-SA-4.0",
            "attribution": "Jens Gustedt, Modern C (INRIA, 2019)",
        },
        "attached_to": {
            "courses": ["wsu/cpts121", "stanford/cs106a"],
            "study_guides": [],
        },
        "owner_role": "bot",
    },
    {
        "slug": "wikimedia-basic-book-design-pdf",
        "source_url": "https://upload.wikimedia.org/wikipedia/commons/0/0c/Basic_Book_Design.pdf",
        "mime_type": "application/pdf",
        "filename": "basic-book-design.pdf",
        "license": {
            "id": "CC-BY-SA-4.0",
            "attribution": "Wikibooks contributors — Basic Book Design",
        },
        "attached_to": {
            "courses": [
                "wsu/cpts121",
                "wsu/cpts260",
                "wsu/math171",
                "wsu/psych105",
                "wsu/hist105",
                "stanford/cs106a",
                "stanford/cs161",
                "stanford/math51",
                "stanford/psych1",
                "stanford/hist1b",
            ],
            "study_guides": [],
        },
        "owner_role": "bot",
    },
    # --- Project Gutenberg EPUB ---
    {
        "slug": "gutenberg-burckhardt-renaissance-epub",
        "source_url": "https://www.gutenberg.org/ebooks/2074.epub.noimages",
        "mime_type": "application/epub+zip",
        "filename": "burckhardt-renaissance-italy.epub",
        "license": {
            "id": "PUBLIC-DOMAIN",
            "attribution": "Jacob Burckhardt, The Civilisation of the Renaissance in Italy — Project Gutenberg #2074",
        },
        "attached_to": {
            "courses": ["stanford/hist1b", "wsu/hist105"],
            "study_guides": [],
        },
        "owner_role": "bot",
    },
    {
        "slug": "gutenberg-pater-renaissance-epub",
        "source_url": "https://www.gutenberg.org/ebooks/4060.epub.noimages",
        "mime_type": "application/epub+zip",
        "filename": "pater-the-renaissance.epub",
        "license": {
            "id": "PUBLIC-DOMAIN",
            "attribution": "Walter Pater, The Renaissance: Studies in Art and Poetry — Project Gutenberg #4060",
        },
        "attached_to": {"courses": ["stanford/hist1b"], "study_guides": []},
        "owner_role": "bot",
    },
    {
        "slug": "gutenberg-beacon-lights-renaissance-reformation-epub",
        "source_url": "https://www.gutenberg.org/ebooks/10532.epub.noimages",
        "mime_type": "application/epub+zip",
        "filename": "beacon-lights-renaissance-reformation.epub",
        "license": {
            "id": "PUBLIC-DOMAIN",
            "attribution": "John Lord, Beacon Lights of History Vol 6: Renaissance and Reformation — Project Gutenberg #10532",
        },
        "attached_to": {
            "courses": ["wsu/hist105", "stanford/hist1b"],
            "study_guides": [],
        },
        "owner_role": "bot",
    },
    # --- Project Gutenberg TXT (text/plain, public domain) ---
    # (RFC mirrors 403'd httpx even with polite UA — Gutenberg is more reliable.)
    {
        "slug": "gutenberg-huckleberry-finn-txt",
        "source_url": "https://www.gutenberg.org/cache/epub/76/pg76.txt",
        "mime_type": "text/plain",
        "filename": "huckleberry-finn.txt",
        "license": {
            "id": "PUBLIC-DOMAIN",
            "attribution": "Mark Twain, Adventures of Huckleberry Finn — Project Gutenberg #76",
        },
        "attached_to": {
            "courses": ["stanford/hist1b", "wsu/hist105"],
            "study_guides": [],
        },
        "owner_role": "bot",
    },
    {
        "slug": "gutenberg-pride-and-prejudice-txt",
        "source_url": "https://www.gutenberg.org/cache/epub/1342/pg1342.txt",
        "mime_type": "text/plain",
        "filename": "pride-and-prejudice.txt",
        "license": {
            "id": "PUBLIC-DOMAIN",
            "attribution": "Jane Austen, Pride and Prejudice — Project Gutenberg #1342 (text)",
        },
        "attached_to": {"courses": ["stanford/hist1b"], "study_guides": []},
        "owner_role": "bot",
    },
    # --- Unsplash JPEGs (all UNSPLASH license, verified 200 image/jpeg) ---
    {
        "slug": "unsplash-abstract-computer-code-1",
        "source_url": "https://images.unsplash.com/photo-1481627834876-b7833e8f5570?w=1200&q=80",
        "mime_type": "image/jpeg",
        "filename": "unsplash-code-1.jpg",
        "license": {"id": "UNSPLASH", "attribution": "Photo on Unsplash"},
        "attached_to": {"courses": ["wsu/cpts121"], "study_guides": []},
        "owner_role": "bot",
    },
    {
        "slug": "unsplash-study-desk-2",
        "source_url": "https://images.unsplash.com/photo-1456513080510-7bf3a84b82f8?w=1200&q=80",
        "mime_type": "image/jpeg",
        "filename": "unsplash-study-desk-2.jpg",
        "license": {"id": "UNSPLASH", "attribution": "Photo on Unsplash"},
        "attached_to": {"courses": ["wsu/math171"], "study_guides": []},
        "owner_role": "bot",
    },
    {
        "slug": "unsplash-library-3",
        "source_url": "https://images.unsplash.com/photo-1532153975070-2e9ab71f1b14?w=1200&q=80",
        "mime_type": "image/jpeg",
        "filename": "unsplash-library-3.jpg",
        "license": {"id": "UNSPLASH", "attribution": "Photo on Unsplash"},
        "attached_to": {"courses": ["stanford/hist1b"], "study_guides": []},
        "owner_role": "bot",
    },
    {
        "slug": "unsplash-study-materials-4",
        "source_url": "https://images.unsplash.com/photo-1497633762265-9d179a990aa6?w=1200&q=80",
        "mime_type": "image/jpeg",
        "filename": "unsplash-study-materials-4.jpg",
        "license": {"id": "UNSPLASH", "attribution": "Photo on Unsplash"},
        "attached_to": {"courses": ["wsu/hist105"], "study_guides": []},
        "owner_role": "bot",
    },
    {
        "slug": "unsplash-neural-abstract-5",
        "source_url": "https://images.unsplash.com/photo-1488751045188-3c55bbf9a3fa?w=1200&q=80",
        "mime_type": "image/jpeg",
        "filename": "unsplash-neural-abstract-5.jpg",
        "license": {"id": "UNSPLASH", "attribution": "Photo on Unsplash"},
        "attached_to": {"courses": ["stanford/cs161"], "study_guides": []},
        "owner_role": "bot",
    },
    {
        "slug": "unsplash-brain-scan-6",
        "source_url": "https://images.unsplash.com/photo-1532012197267-da84d127e765?w=1200&q=80",
        "mime_type": "image/jpeg",
        "filename": "unsplash-brain-6.jpg",
        "license": {"id": "UNSPLASH", "attribution": "Photo on Unsplash"},
        "attached_to": {"courses": ["stanford/psych1"], "study_guides": []},
        "owner_role": "bot",
    },
    {
        "slug": "unsplash-matrix-pattern-7",
        "source_url": "https://images.unsplash.com/photo-1513475382585-d06e58bcb0e0?w=1200&q=80",
        "mime_type": "image/jpeg",
        "filename": "unsplash-matrix-7.jpg",
        "license": {"id": "UNSPLASH", "attribution": "Photo on Unsplash"},
        "attached_to": {"courses": ["stanford/math51"], "study_guides": []},
        "owner_role": "bot",
    },
    {
        "slug": "unsplash-psychology-mind-8",
        "source_url": "https://images.unsplash.com/photo-1576091160550-2173dba999ef?w=1200&q=80",
        "mime_type": "image/jpeg",
        "filename": "unsplash-psych-8.jpg",
        "license": {"id": "UNSPLASH", "attribution": "Photo on Unsplash"},
        "attached_to": {"courses": ["wsu/psych105"], "study_guides": []},
        "owner_role": "bot",
    },
    {
        "slug": "unsplash-circuit-board-9",
        "source_url": "https://images.unsplash.com/photo-1560785496-3c9d27877182?w=1200&q=80",
        "mime_type": "image/jpeg",
        "filename": "unsplash-circuit-9.jpg",
        "license": {"id": "UNSPLASH", "attribution": "Photo on Unsplash"},
        "attached_to": {"courses": ["wsu/cpts260"], "study_guides": []},
        "owner_role": "bot",
    },
    # --- Unsplash PNG (format-converted via ?fm=png) ---
    # Each entry reuses an existing Unsplash photo ID but asks the CDN for
    # a native-PNG rendering. Verified 200 image/png via curl this session.
    {
        "slug": "unsplash-png-code-abstract-1",
        "source_url": "https://images.unsplash.com/photo-1481627834876-b7833e8f5570?fm=png&w=1200",
        "mime_type": "image/png",
        "filename": "unsplash-png-code-1.png",
        "license": {"id": "UNSPLASH", "attribution": "Photo on Unsplash"},
        "attached_to": {"courses": ["wsu/cpts121"], "study_guides": []},
        "owner_role": "bot",
    },
    {
        "slug": "unsplash-png-study-desk-2",
        "source_url": "https://images.unsplash.com/photo-1456513080510-7bf3a84b82f8?fm=png&w=1200",
        "mime_type": "image/png",
        "filename": "unsplash-png-study-desk-2.png",
        "license": {"id": "UNSPLASH", "attribution": "Photo on Unsplash"},
        "attached_to": {"courses": ["wsu/math171"], "study_guides": []},
        "owner_role": "bot",
    },
    {
        "slug": "unsplash-png-library-3",
        "source_url": "https://images.unsplash.com/photo-1532153975070-2e9ab71f1b14?fm=png&w=1200",
        "mime_type": "image/png",
        "filename": "unsplash-png-library-3.png",
        "license": {"id": "UNSPLASH", "attribution": "Photo on Unsplash"},
        "attached_to": {"courses": ["stanford/hist1b"], "study_guides": []},
        "owner_role": "bot",
    },
    {
        "slug": "unsplash-png-materials-4",
        "source_url": "https://images.unsplash.com/photo-1497633762265-9d179a990aa6?fm=png&w=1200",
        "mime_type": "image/png",
        "filename": "unsplash-png-materials-4.png",
        "license": {"id": "UNSPLASH", "attribution": "Photo on Unsplash"},
        "attached_to": {"courses": ["wsu/hist105"], "study_guides": []},
        "owner_role": "bot",
    },
    {
        "slug": "unsplash-png-neural-5",
        "source_url": "https://images.unsplash.com/photo-1488751045188-3c55bbf9a3fa?fm=png&w=1200",
        "mime_type": "image/png",
        "filename": "unsplash-png-neural-5.png",
        "license": {"id": "UNSPLASH", "attribution": "Photo on Unsplash"},
        "attached_to": {"courses": ["stanford/cs161"], "study_guides": []},
        "owner_role": "bot",
    },
    {
        "slug": "unsplash-png-brain-6",
        "source_url": "https://images.unsplash.com/photo-1532012197267-da84d127e765?fm=png&w=1200",
        "mime_type": "image/png",
        "filename": "unsplash-png-brain-6.png",
        "license": {"id": "UNSPLASH", "attribution": "Photo on Unsplash"},
        "attached_to": {"courses": ["stanford/psych1"], "study_guides": []},
        "owner_role": "bot",
    },
    {
        "slug": "unsplash-png-matrix-7",
        "source_url": "https://images.unsplash.com/photo-1513475382585-d06e58bcb0e0?fm=png&w=1200",
        "mime_type": "image/png",
        "filename": "unsplash-png-matrix-7.png",
        "license": {"id": "UNSPLASH", "attribution": "Photo on Unsplash"},
        "attached_to": {"courses": ["stanford/math51"], "study_guides": []},
        "owner_role": "bot",
    },
    {
        "slug": "unsplash-png-psych-8",
        "source_url": "https://images.unsplash.com/photo-1576091160550-2173dba999ef?fm=png&w=1200",
        "mime_type": "image/png",
        "filename": "unsplash-png-psych-8.png",
        "license": {"id": "UNSPLASH", "attribution": "Photo on Unsplash"},
        "attached_to": {"courses": ["wsu/psych105"], "study_guides": []},
        "owner_role": "bot",
    },
    {
        "slug": "unsplash-png-circuit-9",
        "source_url": "https://images.unsplash.com/photo-1560785496-3c9d27877182?fm=png&w=1200",
        "mime_type": "image/png",
        "filename": "unsplash-png-circuit-9.png",
        "license": {"id": "UNSPLASH", "attribution": "Photo on Unsplash"},
        "attached_to": {"courses": ["wsu/cpts260"], "study_guides": []},
        "owner_role": "bot",
    },
    {
        "slug": "unsplash-png-tech-desk-10",
        "source_url": "https://images.unsplash.com/photo-1517694712202-14dd9538aa97?fm=png&w=1200",
        "mime_type": "image/png",
        "filename": "unsplash-png-tech-desk-10.png",
        "license": {"id": "UNSPLASH", "attribution": "Photo on Unsplash"},
        "attached_to": {"courses": ["stanford/cs106a"], "study_guides": []},
        "owner_role": "bot",
    },
    {
        "slug": "unsplash-png-abstract-tech-11",
        "source_url": "https://images.unsplash.com/photo-1515879218367-8466d910aaa4?fm=png&w=1200",
        "mime_type": "image/png",
        "filename": "unsplash-png-abstract-tech-11.png",
        "license": {"id": "UNSPLASH", "attribution": "Photo on Unsplash"},
        "attached_to": {"courses": ["wsu/cpts121"], "study_guides": []},
        "owner_role": "bot",
    },
    {
        "slug": "unsplash-png-notebook-12",
        "source_url": "https://images.unsplash.com/photo-1455390582262-044cdead277a?fm=png&w=1200",
        "mime_type": "image/png",
        "filename": "unsplash-png-notebook-12.png",
        "license": {"id": "UNSPLASH", "attribution": "Photo on Unsplash"},
        "attached_to": {"courses": ["wsu/math171"], "study_guides": []},
        "owner_role": "bot",
    },
    # --- Unsplash WebP (format-converted via ?fm=webp) ---
    {
        "slug": "unsplash-webp-code-1",
        "source_url": "https://images.unsplash.com/photo-1481627834876-b7833e8f5570?fm=webp&w=1200",
        "mime_type": "image/webp",
        "filename": "unsplash-webp-code-1.webp",
        "license": {"id": "UNSPLASH", "attribution": "Photo on Unsplash"},
        "attached_to": {"courses": ["stanford/cs106a"], "study_guides": []},
        "owner_role": "bot",
    },
    {
        "slug": "unsplash-webp-library-2",
        "source_url": "https://images.unsplash.com/photo-1532153975070-2e9ab71f1b14?fm=webp&w=1200",
        "mime_type": "image/webp",
        "filename": "unsplash-webp-library-2.webp",
        "license": {"id": "UNSPLASH", "attribution": "Photo on Unsplash"},
        "attached_to": {"courses": ["wsu/hist105"], "study_guides": []},
        "owner_role": "bot",
    },
    {
        "slug": "unsplash-webp-neural-3",
        "source_url": "https://images.unsplash.com/photo-1488751045188-3c55bbf9a3fa?fm=webp&w=1200",
        "mime_type": "image/webp",
        "filename": "unsplash-webp-neural-3.webp",
        "license": {"id": "UNSPLASH", "attribution": "Photo on Unsplash"},
        "attached_to": {"courses": ["stanford/cs161"], "study_guides": []},
        "owner_role": "bot",
    },
    {
        "slug": "unsplash-webp-circuit-4",
        "source_url": "https://images.unsplash.com/photo-1560785496-3c9d27877182?fm=webp&w=1200",
        "mime_type": "image/webp",
        "filename": "unsplash-webp-circuit-4.webp",
        "license": {"id": "UNSPLASH", "attribution": "Photo on Unsplash"},
        "attached_to": {"courses": ["wsu/cpts260"], "study_guides": []},
        "owner_role": "bot",
    },
]


def main() -> None:
    entries = self_generated_entries() + PUBLIC_SOURCE_ENTRIES
    header = (
        "# Phase 1b corpus — 54 self-generated (pandoc from markdown in files_local/sources/)\n"
        "# + verified Wikimedia / Gutenberg / RFC / Unsplash entries.\n"
        "#\n"
        "# Regenerate via: uv run python fixtures/files_local/assemble_yaml.py\n"
        "# after editing sources/ or this script's PUBLIC_SOURCE_ENTRIES.\n"
        "# Then re-run `./fixtures/files_local/build.sh` if sources changed.\n\n"
    )
    body = yaml.safe_dump(entries, sort_keys=False, default_flow_style=False, allow_unicode=True)
    FILES_YAML.write_text(header + body)
    print(f"wrote {FILES_YAML} — {len(entries)} entries")


if __name__ == "__main__":
    main()
