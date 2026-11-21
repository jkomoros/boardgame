// Writes one recorded Go game from conformance/ as plain JSON, for a pattern
// test to read with dataFile() (which reads text, not gzip).
//
//   node fabric/patterns/tools/recording-fixture.mjs memory/scenario-mismatch fabric/patterns/memory/testdata/recording.json
//   node fabric/patterns/tools/recording-fixture.mjs blackjack/golden-714104226cb9a317 fabric/patterns/blackjack/testdata/recording.json 16
//
// Only the parts a replay reads are kept. The output is not checked in: it
// is a copy of what conformance/ already holds.
import { mkdirSync, readFileSync, writeFileSync } from 'node:fs';
import { dirname, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';
import { gunzipSync } from 'node:zlib';

const [name, out, limit] = process.argv.slice(2);
if (!name || !out) {
  console.error('usage: recording-fixture.mjs <game>/<document> <output.json> [versions]');
  process.exit(2);
}
const corpus = resolve(dirname(fileURLToPath(import.meta.url)), '../../../conformance');
const doc = JSON.parse(gunzipSync(readFileSync(resolve(corpus, `${name}.json.gz`))).toString('utf8'));
// An optional third argument keeps only the first so many versions.
const versions = limit ? doc.versions.slice(0, Number(limit)) : doc.versions;
const kept = { name: doc.name, game: doc.game, engine: doc.engine, versions };
mkdirSync(dirname(resolve(out)), { recursive: true });
writeFileSync(resolve(out), JSON.stringify(kept));
console.log(`${out}: ${versions.length} versions, ${JSON.stringify(kept).length} bytes`);
