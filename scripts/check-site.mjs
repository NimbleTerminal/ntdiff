import assert from 'node:assert/strict';
import { readFile, stat } from 'node:fs/promises';
for (const file of ['index.html', 'site.js', 'site.css', 'app.wasm', 'wasm_exec.js', 'booba/booba.js', 'ghostty-web/ghostty-vt.wasm']) {
  assert.ok((await stat(`web/${file}`)).size > 0, `missing or empty ${file}`);
}
const html = await readFile('web/index.html', 'utf8');
for (const id of ['terminal', 'controls', 'status', 'restart']) assert.ok(html.includes(`id="${id}"`));
assert.ok(!html.includes('src="/'), 'assets must work under the Pages /ntdiff/ prefix');
console.log('Site assets checked.');
