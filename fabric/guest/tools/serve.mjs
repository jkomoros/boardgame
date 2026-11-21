// A static file server for the harness, rooted at the repository so the
// harness can read replays from examples/ in place. Nothing is proxied.
//
//   node tools/serve.mjs            # http://127.0.0.1:8719/fabric/guest/harness.html
//   GUEST_PORT=9000 node tools/serve.mjs
//
// `?csp=network` or `?csp=strict` on the guest document adds a
// Content-Security-Policy header, to show what the bundle needs:
//   network: no network of any kind; inline script and inline style allowed.
//   strict:  the inline script is allowed by hash only, and no inline style.
import { createServer } from 'node:http';
import { createHash } from 'node:crypto';
import { readFile, stat } from 'node:fs/promises';
import { dirname, extname, normalize, resolve, sep } from 'node:path';
import { fileURLToPath } from 'node:url';

const repo = resolve(dirname(fileURLToPath(import.meta.url)), '../../..');
const port = Number(process.env.GUEST_PORT || 8719);
const types = {
  '.html': 'text/html; charset=utf-8',
  '.js': 'text/javascript; charset=utf-8',
  '.mjs': 'text/javascript; charset=utf-8',
  '.json': 'application/json; charset=utf-8',
  '.css': 'text/css; charset=utf-8',
  '.md': 'text/plain; charset=utf-8',
};

function scriptHashes(html) {
  const hashes = [];
  for (const match of html.matchAll(/<script>([\s\S]*?)<\/script>/g)) {
    hashes.push(`'sha256-${createHash('sha256').update(match[1], 'utf8').digest('base64')}'`);
  }
  return hashes.join(' ');
}

const server = createServer(async (request, response) => {
  try {
    const url = new URL(request.url ?? '/', 'http://localhost');
    const path = normalize(decodeURIComponent(url.pathname));
    const file = resolve(repo, '.' + sep + path);
    if (file !== repo && !file.startsWith(repo + sep)) {
      response.writeHead(403).end('forbidden');
      return;
    }
    const info = await stat(file).catch(() => null);
    if (!info || !info.isFile()) {
      response.writeHead(404).end('not found');
      return;
    }
    const body = await readFile(file);
    const headers = {
      'Content-Type': types[extname(file)] ?? 'application/octet-stream',
      'Cache-Control': 'no-store',
    };
    // Only the guest document is put under a policy, never the harness.
    const csp = path.endsWith('/dist/guest.html') ? url.searchParams.get('csp') : null;
    if (csp === 'network') {
      headers['Content-Security-Policy'] =
        "default-src 'none'; script-src 'unsafe-inline'; style-src 'unsafe-inline'; img-src data:; font-src data:";
    } else if (csp === 'strict') {
      headers['Content-Security-Policy'] =
        `default-src 'none'; script-src ${scriptHashes(body.toString('utf8'))}; img-src data:; font-src data:`;
    }
    response.writeHead(200, headers).end(body);
  } catch (error) {
    response.writeHead(500).end(String(error));
  }
});

server.listen(port, '127.0.0.1', () => {
  console.log(`guest harness: http://127.0.0.1:${port}/fabric/guest/harness.html`);
});
