// Real WASM/booba bridge smoke test. Canvas rendering is checked in a browser.
import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import vm from 'node:vm';
import { setTimeout as delay } from 'node:timers/promises';
vm.runInThisContext(await readFile('web/wasm_exec.js', 'utf8'));
const typeset = process.argv.includes('typeset');
const go = new Go();
go.argv = ['demo', ...(typeset ? ['-typeset'] : [])];
let code;
go.exit = value => { code = value; };
const { instance } = await WebAssembly.instantiate(await readFile('web/app.wasm'), go.importObject);
const timeout = setTimeout(() => { console.error('Demo timed out'); process.exit(1); }, 30000);
const running = go.run(instance);
async function outputUntil(predicate) {
  let output = '';
  for (let n = 0; n < 200; n++) {
    await delay(25);
    output += bubbletea_read();
    if (predicate(output)) return output;
  }
  assert.fail(`Expected output not received: ${JSON.stringify(output).slice(-1200)}`);
}
function hasContext(output, context) {
  return [...output.matchAll(/:msg=([A-Za-z0-9+/=]+)\x1b\\/g)].some(match => {
    const payload = Buffer.from(match[1], 'base64').toString('utf8');
    const snapshot = JSON.parse(payload.split(' | booba-actions/v1 ')[1]);
    return snapshot.context === context;
  });
}
try {
  for (let n = 0; typeof globalThis.bubbletea_read !== 'function' && n < 100; n++) await delay(20);
  assert.equal(typeof globalThis.bubbletea_read, 'function');
  bubbletea_resize(100, 36);
  const initial = await outputUntil(s => s.includes(typeset ? 'TYPESET' : 'TERMINAL'));
  assert.ok(initial.includes('Changed Files'), 'file list is present in either starting view');
  bubbletea_write('\x1b]7501;?\x1b\\');
  const output = await outputUntil(s => /\x1b\]7501;state=idle:.*:msg=/.test(s));
  const match = output.match(/:msg=([A-Za-z0-9+/=]+)\x1b\\/);
  const payload = Buffer.from(match[1], 'base64').toString('utf8');
  const snapshot = JSON.parse(payload.split(' | booba-actions/v1 ')[1]);
  assert.ok(snapshot.actions.some(a => a.label === 'Switch view' && a.enabled));
  assert.ok(snapshot.actions.some(a => a.label === 'Switch pane' && a.enabled));
  bubbletea_write('j');
  await outputUntil(s => s.includes('config.json'));
  bubbletea_write('v');
  await outputUntil(s => hasContext(s, typeset ? 'terminal' : 'typeset'));
  bubbletea_write('v');
  await outputUntil(s => hasContext(s, typeset ? 'typeset' : 'terminal'));
  bubbletea_resize(70, 24);
  await outputUntil(s => s.length > 0);
  bubbletea_write('q');
  await running;
  assert.equal(code, 0);
  assert.equal(globalThis.bubbletea_exited, true);
  console.log(`${typeset ? 'typeset' : 'terminal'}: startup, actions, view switching, resize, and exit passed`);
} finally { clearTimeout(timeout); }
// Match go_js_wasm_exec: terminate the host after verified Go exit so pending
// runtime timers do not resume an already exited instance.
process.exit(0);
