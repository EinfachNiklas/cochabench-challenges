# cochabench-challenges

This repository contains coding challenges for [cochabench](https://github.com/EinfachNiklas/cochabench).

The repository is published under the [MIT License](LICENSE).

It currently supports challenges in:
- Go
- Python
- JavaScript

The challenge directories in this repository are intentionally not structured like conventional standalone projects. They are prepared for packaging and later import into `cochabench`, where starter code, solutions, and tests are copied into evaluator-specific locations.

## Repository Structure

The repository is organized around four main parts:

- `challenges/` contains the actual challenge packages.
- `template/` contains starter templates for each supported language.
- `manifest.json` contains the published challenge metadata.
- `.github/workflows/release.yml` builds release assets from the repository contents.

## Challenge Layout

Each challenge is expected to follow this layout:

```text
challenges/<slug>/
├── challenge.config.json
├── src/
└── test/
```

Typical contents:

- `challenge.config.json` defines the challenge name, id, and language.
- `src/` contains the challenge statement and starter code.
- `test/` contains the evaluation tests used later by the benchmarking pipeline.

This structure is intentionally atypical. During evaluation, solution files and tests are copied into the directory layout expected by the target runtime or evaluator.

## Templates

Start new challenges from one of the language templates:

- `template/go`
- `template/python`
- `template/javascript`

Each template includes a `CHALLENGE.md`. That file should contain these sections:

- `Task`
- `Context`
- `Dependencies`
- `Constraints`
- `Edge Cases`

For contribution rules and licensing expectations, see [CONTRIBUTING.md](CONTRIBUTING.md).

## Adding a New Challenge

To add a new challenge:

1. Copy the correct language template from `template/<language>`.
2. Create or rename the challenge directory under `challenges/<slug>/`.
3. Fill in `challenge.config.json`.
4. Write the challenge description in `src/CHALLENGE.md`.
5. Add the starter code in `src/`.
6. Add the evaluation tests in `test/`.
7. Register the challenge in `manifest.json`.

Keep the challenge id, slug, language, and packaged filename conventions consistent across the directory name, `challenge.config.json`, and `manifest.json`.

If a challenge is inspired by or derived from third-party material, follow [THIRD_PARTY_CHALLENGES.md](THIRD_PARTY_CHALLENGES.md) before opening a pull request.

## Notes on Tests

Local tests in this repository are not required to be green in the current layout.

That is expected. The repository stores challenge material in a packaging-oriented structure, and the files are later moved into evaluator-specific directories before execution. Even so, contributors should keep starter code and tests logically compatible with the final evaluation setup.

## Releases

Releases package the challenge directories together with `manifest.json`.

The release artifacts are intended to be consumed by `cochabench`, not treated as polished standalone challenge repositories.

## Current Challenges

- `graph-pathfinding` - JavaScript
- `lru-cache-with-ttl` - Python
- `web-crawler` - Go
- `task-scheduler` - Go
