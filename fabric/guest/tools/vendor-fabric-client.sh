#!/bin/sh
# Bundles Common Fabric's iframe guest client out of a labs checkout into
# vendor/commonfabric-guest.js, which vite.fabric.config.ts aliases
# `@commonfabric/iframe-sandbox/guest` to.
#
#   tools/vendor-fabric-client.sh /path/to/labs
#
# The client is labs' own source (base/iframe-sandbox/src/guest.ts) and has to
# match the runtime the pattern is deployed to, so it is built from the
# checkout that will serve the pattern and is not kept in this repository.
# Needs Deno (labs pins the version in mise.toml).
set -e
labs=${1:-$CF_LABS_ROOT}
if [ -z "$labs" ] || [ ! -f "$labs/base/iframe-sandbox/src/guest.ts" ]; then
  echo "usage: $0 <path to a labs checkout>   (or set CF_LABS_ROOT)" >&2
  exit 2
fi
here=$(cd "$(dirname "$0")/.." && pwd)
mkdir -p "$here/vendor"
# Run from the labs checkout so its import map applies.
(cd "$labs" && deno bundle --platform browser base/iframe-sandbox/src/guest.ts \
  -o "$here/vendor/commonfabric-guest.js")
printf '%s\n' "$(cd "$labs" && git rev-parse HEAD)" > "$here/vendor/LABS_COMMIT"
echo "wrote $here/vendor/commonfabric-guest.js (labs $(cat "$here/vendor/LABS_COMMIT"))"
