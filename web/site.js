import { BoobaTerminal, installActionBar, installRendererHud, parseRendererFromURL } from './booba/booba.js';

const view = new URLSearchParams(location.search).get('view') === 'typeset' ? 'typeset' : 'terminal';
const status = document.querySelector('#status');
document.querySelector('#restart').href = location.href;
for (const link of document.querySelectorAll('[data-view]')) {
  if (link.dataset.view === view) link.setAttribute('aria-current', 'page');
}
function failed(error) {
  status.dataset.error = '';
  status.textContent = `Could not run the demo: ${error.message}. Try Restart.`;
  console.error(error);
}
async function main() {
  const go = new Go();
  go.argv = ['ntdiff-demo', ...(view === 'typeset' ? ['-typeset'] : [])];
  const response = await fetch('app.wasm');
  if (!response.ok) throw new Error(`Download failed (${response.status})`);
  const { instance } = await WebAssembly.instantiateStreaming(response, go.importObject);
  const running = go.run(instance);
  let ended = false;
  running.then(() => { ended = true; status.textContent = 'Session ended. Choose Restart to run again.'; }, failed);
  const terminal = new BoobaTerminal('terminal', { renderer: parseRendererFromURL(), allowOSC7501Actions: true });
  await terminal.init();
  terminal.setActionKeyEncoder({
    canEncode: gesture => terminal.encodeKey(gesture) !== null,
    encode: gesture => terminal.encodeKey(gesture),
  });
  installActionBar(terminal.actions, {
    parent: document.querySelector('#controls'),
    openKeyboard: () => terminal.focus(),
    onLayout: () => terminal.fitAddon?.fit(),
  });
  terminal.onActionError = failed;
  window.addEventListener('pagehide', () => terminal.dispose(), { once: true });
  terminal.connectWasm();
  terminal.focus();
  installRendererHud(terminal.term, { parent: document.querySelector('.terminal-frame'), position: 'absolute', bindToggleHotkey: false });
  if (!ended) status.textContent = 'Running locally in your browser. Press v to switch views without reparsing the diff.';
}
main().catch(failed);
