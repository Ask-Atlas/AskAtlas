# AskAtlas API — Bruno Collection

A [Bruno](https://www.usebruno.com) collection covering every operation in
`api/openapi.yaml`. Replaces the previous Postman workflow.

## Layout

```
bruno/
├── bruno.json               # collection manifest (auto-detected by Bruno)
├── collection.bru           # collection-root auth block (bearer {{authToken}})
├── README.md                # this file
├── environments/
│   ├── dev.bru              # baseUrl + authToken + path-param UUIDs for dev
│   └── staging.bru          # same shape, pointed at staging
├── Courses/                 # one folder per top-level path segment
├── Files/
├── Me/
├── Quizzes/
├── Schools/
├── Sessions/
└── Study Guides/
```

Each folder contains one `.bru` per operation. Folder grouping is the first
non-variable segment of the OpenAPI path, so `/api/files/{id}` lands under
`Files/` and `/api/me/recents` lands under `Me/`.

## Quickstart

1. Install Bruno (`brew install --cask bruno` or grab the app).
2. `File → Open Collection` → point at `api/bruno/`.
3. Pick the `dev` or `staging` environment from the dropdown.
4. Paste a Clerk JWT into the `authToken` var (it's flagged secret so
   it won't be printed in logs or committed accidentally).
5. Send any request — bearer auth inherits from the collection root.

### Grabbing a test JWT

Clerk's dashboard under *Configure → Sessions → Tokens → Customize session
token* exposes a long-lived testing JWT per environment. Use those values
in `authToken`; they're already whitelisted against the matching API
origin. Do not check real JWTs into git — the `vars:secret` block keeps
them out of the file on save, but the `~authToken:` line is disabled by
default so accidental commits land empty.

## Regenerating

`.bru` files (except `bruno.json`, `collection.bru`, and anything under
`environments/`) are generated from `api/openapi.yaml` by
`api/scripts/brunogen/main.go`. Re-run after any spec change:

```sh
make bruno
```

The generator:

- wipes every subfolder except `environments/`
- walks `paths` in sorted order, writes one `.bru` per operation
- picks the first JSON example (or synthesizes one from the schema) for
  request bodies
- substitutes path params `{file_id}` → `{{fileId}}` (snake → camel) so
  they bind to the env vars
- renders `summary` + `description` into the `docs { }` block so each
  request shows inline context

Output is deterministic — re-running without spec changes produces zero
diff.

## Path parameters

Every path-param UUID defaults to the NIL UUID
(`00000000-0000-0000-0000-000000000000`) in both env files. An
accidental send hits a deterministic 404 instead of mutating real data.
Override individual vars in the env before sending requests that
actually need to reach a row.

## Adding an environment

Copy `environments/dev.bru` to `environments/<name>.bru`, change
`baseUrl`, keep the `vars:secret [ authToken ]` block. The generator
leaves the `environments/` folder alone, so hand-edits stick.
