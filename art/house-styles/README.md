# BOARDGAME house styles

This is the shared catalog of visual languages that have survived a game's
four-option exploration, iterative narrowing, and explicit style lock. Each
style directory contains:

- `style.md`: reusable prose art direction;
- `style.png`: the locked visual reference, stored with Git LFS;
- `style.png.imagegen.json`: original generation provenance when available;
- `style.png.style-lock.json`: the verified image hash and selection record;
- `house-style.json`: catalog metadata.

Do not add raw explorations here. Promote only an explicitly selected style
with `boardgame-util imagegen house-style-add`; the command rebuilds
`catalog.json` deterministically. House styles must be original or properly
licensed, and their metadata must accurately disclose source assets.
