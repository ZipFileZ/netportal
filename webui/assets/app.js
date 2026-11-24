/* NET PORTAL — interface logic. Plain JS, no build step.
 * The page talks to Go through chrome.webview.postMessage (see bridge.go). */
'use strict';

/* ------------------------------------------------------------------ RPC -- */

const np = {
  seq: 0,
  pending: new Map(),
  listeners: {},
  call(method, args) {
    const id = ++this.seq;
    return new Promise((resolve, reject) => {
      this.pending.set(id, { resolve, reject });
      window.chrome.webview.postMessage(JSON.stringify({ id, m: method, a: args || {} }));
    });
  },
  _reply(id, ok, data) {
    const p = this.pending.get(id);
    if (!p) return;
    this.pending.delete(id);
    ok ? p.resolve(data) : p.reject(data);
  },
  _event(ev) {
    (this.listeners[ev.name] || []).forEach((f) => f(ev.data));
  },
  on(name, f) { (this.listeners[name] = this.listeners[name] || []).push(f); },
};
window.np = np;

const BOOT = window.NP_BOOT || { view: 'main' };
const IS_TRAY = BOOT.view === 'tray';

/* -------------------------------------------------------------- helpers -- */

const esc = (s) => String(s == null ? '' : s).replace(/[&<>"']/g, (c) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[c]));
const $ = (sel, root) => (root || document).querySelector(sel);

const ICON = {
  tunnels: 'M8 1.5l5.5 2v4.2c0 3.2-2.3 5.7-5.5 6.8-3.2-1.1-5.5-3.6-5.5-6.8V3.5z',
  log: 'M3 3.5h10M3 6.5h10M3 9.5h10M3 12.5h6',
  settings: 'M2.5 4.5h6M11.5 4.5h2M2.5 11.5h2M7.5 11.5h6M10 3v3M6 10v3',
  about: 'M8 14.5A6.5 6.5 0 1 0 8 1.5a6.5 6.5 0 0 0 0 13zM8 7.2v4M8 4.8v.1',
  download: 'M8 2.5v7.5M4.8 7l3.2 3.2L11.2 7M3 13.5h10',
  plus: 'M8 3v10M3 8h10',
  chevron: 'M4 6l4 4 4-4',
  back: 'M10 3L5 8l5 5',
  search: 'M7 11.5a4.5 4.5 0 1 0 0-9 4.5 4.5 0 0 0 0 9zM10.5 10.5L14 14',
  zip: 'M2.5 4.5a2 2 0 0 1 2-2h7a2 2 0 0 1 2 2v7a2 2 0 0 1-2 2h-7a2 2 0 0 1-2-2zM8 2.5v3M7 6.5h2M7 8.5h2v2H7z',
  multi: 'M3.2 2h2.6A1.2 1.2 0 0 1 7 3.2v2.6A1.2 1.2 0 0 1 5.8 7H3.2A1.2 1.2 0 0 1 2 5.8V3.2A1.2 1.2 0 0 1 3.2 2zM3.2 9h2.6A1.2 1.2 0 0 1 7 10.2v2.6A1.2 1.2 0 0 1 5.8 14H3.2A1.2 1.2 0 0 1 2 12.8v-2.6A1.2 1.2 0 0 1 3.2 9zM9.5 4.5h4.5M9.5 11.5h4.5',
  trash: 'M3 4.5h10M6.5 4.5V3h3v1.5M4.5 4.5l.6 8.5a1 1 0 0 0 1 .9h3.8a1 1 0 0 0 1-.9l.6-8.5',
  edit: 'M10.5 2.5l3 3L6 13H3v-3z',
  export: 'M8 10V2.5M5 5.5l3-3 3 3M3 10v2.5a1 1 0 0 0 1 1h8a1 1 0 0 0 1-1V10',
  power: 'M8 2v5.5M4.6 4.2a5 5 0 1 0 6.8 0',
  copy: 'M7 5.5h5A1.5 1.5 0 0 1 13.5 7v5a1.5 1.5 0 0 1-1.5 1.5H7A1.5 1.5 0 0 1 5.5 12V7A1.5 1.5 0 0 1 7 5.5zM10.5 3.5V3a1 1 0 0 0-1-1H3a1 1 0 0 0-1 1v6.5a1 1 0 0 0 1 1h.5',
  eye: 'M1.5 8s2.4-4.5 6.5-4.5S14.5 8 14.5 8 12.1 12.5 8 12.5 1.5 8 1.5 8zM8 10a2 2 0 1 0 0-4 2 2 0 0 0 0 4z',
  shieldObf: 'M8 1.5l5.5 2v4.2c0 3.2-2.3 5.7-5.5 6.8-3.2-1.1-5.5-3.6-5.5-6.8V3.5zM5.2 8.4l1.4-1.6 1.4 1.6 1.4-1.6 1.4 1.6',
  warn: 'M8 1.8l6.5 11.4H1.5zM8 6.2v3.2M8 11.3v.1',
  alertCircle: 'M8 14.2A6.2 6.2 0 1 0 8 1.8a6.2 6.2 0 0 0 0 12.4zM8 4.8v3.6M8 10.9v.1',
  folder: 'M2 4.5a1 1 0 0 1 1-1h3.5l1.5 1.5h5a1 1 0 0 1 1 1v6.5a1 1 0 0 1-1 1H3a1 1 0 0 1-1-1z',
  paste: 'M5.5 2.5h5v2h-5zM5.5 3.5H4a1 1 0 0 0-1 1v9a1 1 0 0 0 1 1h8a1 1 0 0 0 1-1v-9a1 1 0 0 0-1-1h-1.5',
  qr: 'M2.5 2.5h4v4h-4zM9.5 2.5h4v4h-4zM2.5 9.5h4v4h-4zM9.5 9.5h1.5v1.5M13.5 9.5v1.5M9.5 13.5h4',
  newFile: 'M9.5 1.5H4a1 1 0 0 0-1 1v11a1 1 0 0 0 1 1h8a1 1 0 0 0 1-1V5zM8 7.5v4M6 9.5h4',
  fileDown: 'M9.5 1.5H4a1 1 0 0 0-1 1v11a1 1 0 0 0 1 1h8a1 1 0 0 0 1-1V5zM9.5 1.5V5H13M8 7.5v4.5M5.8 9.8L8 12l2.2-2.2',
  check: 'M2.5 6.2l2.4 2.3 4.6-5',
  ok: 'M3.5 8.3l3 2.9 6-6.4',
  bang: 'M8 4v5M8 11.6v.1',
  close: 'M1 1l8 8M9 1L1 9',
  monitor: 'M2.5 3h11v7.5h-11zM6 13.5h4M8 10.5v3',
  image: 'M2.5 2.5h11v11h-11zM2.5 11l3.5-3.5 3 3 2-2 2.5 2.5M10.5 5.5v.1',
};

function svg(name, size, extra) {
  const sw = (extra && extra.sw) || 1.5;
  const cls = extra && extra.cls ? ` class="${extra.cls}"` : '';
  return `<svg${cls} aria-hidden="true" width="${size || 16}" height="${size || 16}" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="${sw}" stroke-linecap="round" stroke-linejoin="round"><path d="${ICON[name]}"/></svg>`;
}
const svgCheck = (size) => `<svg aria-hidden="true" width="${size}" height="${size}" viewBox="0 0 12 12" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="${ICON.check}"/></svg>`;
const logoSvg = (size) => `<svg aria-hidden="true" width="${size}" height="${size}" viewBox="0 0 64 64"><rect width="64" height="64" rx="16" fill="#0077D9"/><path d="M32 13l15 5.5v11.3c0 8.8-6.3 15.6-15 18.7-8.7-3.1-15-9.9-15-18.7V18.5z" fill="none" stroke="#fff" stroke-width="2.4" stroke-linejoin="round" opacity="0.55"/><path d="M21.5 33.5l5.2-7 5.3 7 5.3-7 5.2 7" fill="none" stroke="#fff" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"/></svg>`;

function locale() { return LANG === 'ru' ? 'ru-RU' : 'en-US'; }
function fmtNum(n, digits) { return Number(n).toLocaleString(locale(), { minimumFractionDigits: digits, maximumFractionDigits: digits }); }
function fmtBytes(b) {
  const units = ['B', 'KiB', 'MiB', 'GiB', 'TiB'];
  let i = 0;
  let v = b;
  while (v >= 1024 && i < units.length - 1) { v /= 1024; i++; }
  return (i === 0 ? fmtNum(v, 0) : fmtNum(v, v < 10 ? 2 : 1)) + ' ' + t('unit.' + units[i]);
}
function fmtMbps(bytesPerSec) {
  const v = (bytesPerSec * 8) / 1e6;
  return fmtNum(v, v < 10 ? 2 : 1);
}
function fmtDuration(ms) {
  const s = Math.max(0, Math.floor(ms / 1000));
  const p = (n) => String(n).padStart(2, '0');
  return `${p(Math.floor(s / 3600))}:${p(Math.floor((s % 3600) / 60))}:${p(s % 60)}`;
}
function fmtAgo(unixSec) {
  if (!unixSec) return t('peer.never');
  const s = Math.max(0, Math.floor(Date.now() / 1000 - unixSec));
  if (s < 2) return t('time.now');
  if (s < 60) return t('time.sAgo', s);
  if (s < 3600) return t('time.mAgo', Math.floor(s / 60), s % 60);
  if (s < 86400) return t('time.hAgo', Math.floor(s / 3600), Math.floor((s % 3600) / 60));
  return t('time.dAgo', Math.floor(s / 86400));
}
function fmtTime(ms) {
  const d = new Date(ms);
  const p = (n, l) => String(n).padStart(l || 2, '0');
  return `${p(d.getHours())}:${p(d.getMinutes())}:${p(d.getSeconds())}.${p(d.getMilliseconds(), 3)}`;
}
function errText(e) {
  if (!e) return '';
  if (typeof e === 'string') return e;
  const key = 'code.' + e.code;
  if (I18N.ru[key]) return t(key, e.detail || '');
  return e.detail || e.code || String(e);
}

/* ---------------------------------------------------------------- state -- */

const S = {
  ready: false,
  admin: false, demo: false,
  settings: { theme: 'system', language: 'ru' },
  autostart: false, autostartKnown: false,
  about: {},
  update: 'none', upd: null,
  tunnels: [],
  selected: null,
  query: '',
  multi: false, checked: new Set(),
  screen: 'tunnels', // tunnels | log | settings | about | editor
  details: null, detailsFor: null,
  obfOpen: false, showPsk: {},
  stats: {}, // name -> {samples:[{t,rx,tx}], rx, tx, hs, rates:[{rx,tx}], peers}
  since: {}, // name -> ms when we saw it become connected
  extIP: {}, // name -> {state, ip}
  overlay: null, // {kind:'add'|'qr'|'delete'|'picker'|'leave', ...}
  drop: 0,
  toasts: [],
  log: { lines: [], cursor: null, level: 'all', query: '', auto: true },
  editor: null,
};

function stateOf(tn) {
  if (!tn) return 'idle';
  switch (tn.state) {
    case 'started': return 'connected';
    case 'starting': return 'connecting';
    case 'stopping': return 'disconnecting';
    case 'stopped': return tn.error ? 'error' : 'idle';
  }
  return 'unknown';
}
const tunnelByName = (n) => S.tunnels.find((x) => x.name === n);
const activeTunnel = () => S.tunnels.find((x) => x.state === 'started') || S.tunnels.find((x) => x.state === 'starting' || x.state === 'stopping');

/* ---------------------------------------------------------------- theme -- */

const mq = window.matchMedia('(prefers-color-scheme: dark)');
function effectiveTheme() {
  const th = S.settings.theme;
  return th === 'dark' || th === 'light' ? th : (mq.matches ? 'dark' : 'light');
}
function applyTheme() {
  const th = effectiveTheme();
  document.documentElement.dataset.theme = th;
  document.documentElement.lang = LANG;
  if (!IS_TRAY) np.call('setChrome', { dark: th === 'dark' }).catch(() => {});
}
mq.addEventListener('change', () => { applyTheme(); render(); });

/* --------------------------------------------------------------- toasts -- */

function toast(kind, title, text) {
  const id = Math.random().toString(36).slice(2);
  S.toasts.push({ id, kind, title, text });
  if (S.toasts.length > 3) S.toasts.shift();
  renderToasts();
  setTimeout(() => { S.toasts = S.toasts.filter((x) => x.id !== id); renderToasts(); }, kind === 'err' ? 7000 : 3800);
}
function renderToasts() {
  const el = $('#toasts');
  if (!el) return;
  el.innerHTML = S.toasts.map((x) => `
    <div class="toast ${x.kind}" role="status">
      <span class="ti">${svg(x.kind === 'err' ? 'bang' : 'ok', 13, { sw: 2 })}</span>
      <div class="grow"><b>${esc(x.title)}</b>${x.text ? `<span>${esc(x.text)}</span>` : ''}</div>
      <button class="icon-btn plain" data-act="closeToast" data-id="${x.id}" aria-label="${esc(t('toast.close'))}"><svg aria-hidden="true" width="10" height="10" viewBox="0 0 10 10"><path d="${ICON.close}" stroke="currentColor" stroke-width="1.4" stroke-linecap="round"/></svg></button>
    </div>`).join('');
}
function errorToast(title, e) { toast('err', title, errText(e)); }

/* ------------------------------------------------------------ rendering -- */

let renderQueued = false;
function render() {
  if (renderQueued) return;
  renderQueued = true;
  requestAnimationFrame(() => { renderQueued = false; renderNow(); });
}

// Preserve focus, caret and scroll positions across full re-renders.
function renderNow() {
  if (!S.ready) return;
  const active = document.activeElement;
  const focusKey = active && active.dataset ? active.dataset.key : null;
  const sel = focusKey && 'selectionStart' in active ? [active.selectionStart, active.selectionEnd] : null;
  const scrolls = {};
  document.querySelectorAll('[data-scroll]').forEach((el) => { scrolls[el.dataset.scroll] = el.scrollTop; });

  if (IS_TRAY) {
    document.body.classList.add('tray');
    $('#app').innerHTML = renderTray();
  } else if (S.screen === 'editor' && $('#editor-root')) {
    // Keep the editor's DOM (text, caret, undo); refresh only the rest.
    $('#overlay-slot').innerHTML = renderOverlay() + renderDrop();
    renderEditorChrome();
    renderToasts();
    if (S.overlay && S.overlay.kind === 'leave') { const c = $('[data-key="dlg-cancel"]'); if (c) c.focus(); }
    return;
  } else {
    $('#app').innerHTML = renderMain();
    if (S.screen === 'editor') mountEditor();
  }

  document.querySelectorAll('[data-scroll]').forEach((el) => {
    if (scrolls[el.dataset.scroll] !== undefined) el.scrollTop = scrolls[el.dataset.scroll];
  });
  if (focusKey) {
    const el = document.querySelector(`[data-key="${focusKey}"]`);
    if (el) {
      el.focus();
      if (sel && 'setSelectionRange' in el) try { el.setSelectionRange(sel[0], sel[1]); } catch (e) { /* not a text input */ }
    }
  }
  if (S.screen === 'log') scrollLogIfNeeded(true);
  renderToasts();
}

function renderMain() {
  const navItems = [['tunnels', 'nav.tunnels'], ['log', 'nav.log'], ['settings', 'nav.settings'], ['about', 'nav.about']];
  const cur = S.screen === 'editor' ? 'tunnels' : S.screen;
  const showUpd = S.update === 'available' || (S.upd && !S.upd.complete);
  const nav = `
    <nav class="nav" aria-label="${esc(t('nav.tunnels'))}">
      ${navItems.map(([id, key]) => `
        <button class="nav-item" data-act="nav" data-screen="${id}" ${cur === id ? 'aria-current="page"' : ''} title="${esc(t(key))}">
          ${svg(id, 16)}<span class="nav-label">${esc(t(key))}</span>
        </button>`).join('')}
      <div class="nav-spacer"></div>
      ${S.demo ? `<div class="demo-pill">${esc(t('demo'))}</div>` : ''}
      ${showUpd ? `
        <div class="update-card">
          <div class="title">${svg('download', 16)}${esc(t('update.available'))}</div>
          <div class="sub">${esc(t('update.short'))}</div>
          ${S.admin ? `<button class="btn primary" data-act="nav" data-screen="about">${esc(t('update.button'))}</button>` : ''}
        </div>
        <button class="update-pill" data-act="nav" data-screen="about" aria-label="${esc(t('update.available'))}">${svg('download', 18)}</button>` : ''}
    </nav>`;

  let body = '';
  switch (S.screen) {
    case 'log': body = renderLog(); break;
    case 'settings': body = renderSettings(); break;
    case 'about': body = renderAbout(); break;
    case 'editor': body = '<section class="editor" id="editor-root"></section>'; break;
    default: body = S.tunnels.length ? renderTunnelsList() + renderDetails() : renderEmpty();
  }
  return `
    <div class="app">
      ${nav}
      <main class="panel">${body}</main>
      <div id="overlay-slot">${renderOverlay()}${renderDrop()}</div>
      <div class="toasts" id="toasts" aria-live="polite"></div>
    </div>`;
}

function renderDrop() {
  return S.drop > 0 ? `<div class="drop"><div><div class="big">${svg('download', 30, { sw: 1.3 })}</div><div class="t">${esc(t('drop.title'))}</div><div class="s">${esc(t('drop.sub'))}</div></div></div>` : '';
}

/* ----- tunnels list ----- */

function rowSub(tn) {
  const st = stateOf(tn);
  if (st === 'connected') return S.since[tn.name] ? t('row.connected', fmtDuration(Date.now() - S.since[tn.name])) : t('row.connectedNoTime');
  if (st === 'idle') return t(tn.full ? 'type.full' : 'type.partial');
  return t('st.' + st);
}

function renderTunnelsList() {
  const q = S.query.trim().toLowerCase();
  const items = S.tunnels.filter((x) => !q || x.name.toLowerCase().includes(q));
  const rows = items.map((tn) => {
    const st = stateOf(tn);
    const on = st === 'connected' || st === 'connecting';
    const sel = tn.name === S.selected && !S.multi;
    const checked = S.multi && S.checked.has(tn.name);
    return `
      <div class="row st-${st}${sel ? ' sel' : ''}${checked ? ' checked' : ''}">
        ${S.multi ? `<button class="check sm" role="checkbox" aria-checked="${checked}" aria-label="${esc(t('list.select', tn.name))}" data-act="check" data-name="${esc(tn.name)}">${svgCheck(10)}</button>` : ''}
        <button class="row-main" data-act="${S.multi ? 'check' : 'select'}" data-name="${esc(tn.name)}">
          <span class="dot st halo st-${st}"></span>
          <span class="row-text">
            <span class="row-name ellipsis">${esc(tn.name)}</span>
            <span class="row-sub${st !== 'idle' ? ' live' : ''}" data-live="rowsub" data-name="${esc(tn.name)}">${esc(rowSub(tn))}</span>
          </span>
        </button>
        <button class="switch sm${st === 'connecting' ? ' busy' : ''}" role="switch" aria-checked="${on}" aria-label="${esc(t('list.toggle', tn.name))}" data-act="toggle" data-name="${esc(tn.name)}"></button>
      </div>`;
  }).join('');

  const foot = S.multi ? `
      <div class="multi-bar">
        <b>${esc(t('list.selected', S.checked.size))}</b>
        ${S.admin ? `<button class="btn sm" data-act="exportChecked" ${S.checked.size ? '' : 'disabled'}>${esc(t('list.export'))}</button>
        <button class="btn sm danger" data-act="deleteChecked" ${S.checked.size ? '' : 'disabled'}>${esc(t('list.remove'))}</button>` : ''}
        <button class="btn sm" data-act="multiOff">${esc(t('list.cancel'))}</button>
      </div>` : (S.admin ? `
      <div class="list-foot">
        <button class="btn ghost sm grow" data-act="exportAll">${svg('zip', 14)}${esc(t('list.exportAll'))}</button>
        <button class="btn ghost icon sm" data-act="multiOn" aria-label="${esc(t('list.multi'))}" title="${esc(t('list.multi'))}">${svg('multi', 15)}</button>
        <button class="btn ghost icon sm" data-act="deleteSelected" aria-label="${esc(t('list.removeSel'))}" title="${esc(t('list.removeSel'))}">${svg('trash', 15)}</button>
      </div>` : '');

  return `
    <section class="list" aria-label="${esc(t('list.title'))}">
      <div class="list-head">
        <div class="list-title">
          <h2>${esc(t('list.title'))}</h2>
          <span class="count">${S.tunnels.length}</span>
          <span class="grow"></span>
          ${S.admin ? `<button class="btn primary add-btn" data-act="addMenu" aria-haspopup="menu" id="add-btn">${svg('plus', 14, { sw: 1.8 })}${esc(t('list.add'))}${svg('chevron', 12, { sw: 1.8 })}</button>` : ''}
        </div>
        <label class="field">${svg('search', 14, { sw: 1.6 })}<input data-key="query" data-input="query" placeholder="${esc(t('list.search'))}" aria-label="${esc(t('list.search'))}" value="${esc(S.query)}"><span class="kbd">Ctrl+F</span></label>
      </div>
      <div class="list-body" data-scroll="list">${rows || `<div class="list-empty">${esc(t('list.nothing'))}</div>`}</div>
      ${foot}
    </section>`;
}

/* ----- details ----- */

function heroSub(st, d, tn) {
  const ep = d && d.peers && d.peers[0] && d.peers[0].endpoint;
  switch (st) {
    case 'connecting': return ep ? t('hero.sub.connecting', ep.replace(/:\d+$/, '')) : t('hero.sub.connectingNoEp');
    case 'connected':
      if (tn && tn.full) return d && d.killSwitch ? t('hero.sub.connected.full.ks') : t('hero.sub.connected.full');
      return t('hero.sub.connected.partial');
    default: return t('hero.sub.' + (st === 'unknown' ? 'idle' : st));
  }
}

function extIPCell(name, st) {
  if (st !== 'connected') return `<span class="mono muted">—</span>`;
  if (!S.settings.externalIP) return `<span class="muted">${esc(t('hero.extip.off'))}</span>`;
  const e = S.extIP[name];
  if (!e || e.state === 'loading') return `<span class="muted">${esc(t('hero.extip.loading'))}</span>`;
  if (e.state === 'failed') return `<span class="muted">${esc(t('hero.extip.failed'))}</span>`;
  return `<span class="mono selectable">${esc(e.ip)}</span>`;
}

function renderStats(name, st) {
  const s = S.stats[name];
  const live = st === 'connected' && s;
  const rates = live ? s.rates : [];
  const last = rates.length ? rates[rates.length - 1] : { rx: 0, tx: 0 };
  let graph;
  if (live && rates.length > 1) {
    const W = 600, H = 72;
    const pts = rates.slice(-60);
    const max = Math.max(1, ...pts.map((r) => Math.max(r.rx, r.tx))) * 1.15;
    const off = 60 - pts.length;
    const xy = (v, i) => `${(((i + off) / 59) * W).toFixed(1)},${(H - 3 - (v / max) * (H - 8)).toFixed(1)}`;
    const down = pts.map((r, i) => xy(r.rx, i)).join(' ');
    const up = pts.map((r, i) => xy(r.tx, i)).join(' ');
    const x0 = ((off / 59) * W).toFixed(1);
    graph = `
      <svg class="graph" role="img" aria-label="${esc(t('stats.graph'))}" viewBox="0 0 600 72" preserveAspectRatio="none">
        <path d="M0 24H600M0 48H600" stroke="var(--border-soft)" stroke-width="1" vector-effect="non-scaling-stroke"/>
        <polygon points="${x0},72 ${down} 600,72" fill="var(--accent-soft)"/>
        <polyline points="${down}" fill="none" stroke="var(--accent)" stroke-width="1.75" stroke-linejoin="round" vector-effect="non-scaling-stroke"/>
        <polyline points="${up}" fill="none" stroke="var(--up)" stroke-width="1.5" stroke-linejoin="round" vector-effect="non-scaling-stroke"/>
      </svg>`;
  } else {
    graph = `<div class="graph-empty">${esc(t(st === "connecting" ? "stats.wait" : (st === "connected" ? "stats.collecting" : "stats.after")))}</div>`;
  }
  return `
    <div class="card" id="stats-card">
      <div class="stats-head">
        <div><div class="stat-label"><i style="background:var(--accent)"></i>${esc(t('stats.down'))}</div>
          <div class="stat-value tnum">${live ? fmtMbps(last.rx) : '0'} <small>${esc(t('stats.mbps'))}</small></div></div>
        <div><div class="stat-label"><i style="background:var(--up)"></i>${esc(t('stats.up'))}</div>
          <div class="stat-value tnum">${live ? fmtMbps(last.tx) : '0'} <small>${esc(t('stats.mbps'))}</small></div></div>
        <span class="grow"></span>
        <span class="muted" style="font-size:11px;line-height:16px">${esc(t('stats.window'))}</span>
      </div>
      ${graph}
      <div class="stats-foot">
        <div><div class="k">${esc(t('stats.received'))}</div><div class="v tnum">${live ? fmtBytes(s.rx) : '—'}</div></div>
        <div><div class="k">${esc(t('stats.sent'))}</div><div class="v tnum">${live ? fmtBytes(s.tx) : '—'}</div></div>
        <div><div class="k">${esc(t('stats.handshake'))}</div><div class="v">${live ? esc(fmtAgo(s.hs)) : '—'}</div></div>
      </div>
    </div>`;
}

function copyBtn(value, label) {
  return `<button class="icon-btn" data-act="copy" data-value="${esc(value)}" data-label="${esc(label)}" aria-label="${esc(t('copy'))}" title="${esc(t('copy'))}">${svg('copy', 13)}</button>`;
}

function renderInterface(d) {
  const i = d.interface;
  const row = (label, value, opts) => `
    <span>${esc(label)}</span>
    <span class="val"><span class="ellipsis grow ${opts && opts.mono ? 'mono' : ''} selectable" title="${esc(value)}">${esc(value || '—')}</span>${opts && opts.copy && value ? copyBtn(value, label) : ''}</span>`;
  const scripts = [['pre-up', i.preUp], ['post-up', i.postUp], ['pre-down', i.preDown], ['post-down', i.postDown]];
  return `
    <div class="card">
      <h3>${esc(t('iface.title'))}</h3>
      <div class="table">
        ${row(t('iface.publicKey'), i.publicKey, { mono: true, copy: true })}
        ${i.listenPort ? row(t('iface.port'), i.listenPort, { mono: true }) : ''}
        ${i.mtu ? row(t('iface.mtu'), i.mtu, { mono: true }) : ''}
        ${row(t('iface.addresses'), (i.addresses || []).join(', '), { mono: true })}
        ${i.dns && i.dns.length ? row(t('iface.dns'), i.dns.join(', '), { mono: true }) : ''}
        ${row(t('iface.table'), i.tableOff ? t('iface.off') : t('iface.auto'), { mono: true })}
        ${scripts.some((x) => x[1]) ? `<span>${esc(t('iface.scripts'))}</span>
          <span class="scripts mono">${scripts.map(([k, v]) => `<span class="muted">${k}</span><span class="ellipsis ${v ? '' : 'muted'}" title="${esc(v)}">${esc(v || '—')}</span>`).join('')}</span>` : ''}
      </div>
    </div>`;
}

function renderObfuscation(d) {
  const items = d.obfuscation || [];
  const label = (k) => (I18N.ru['obf.' + k] ? t('obf.' + k) : k);
  return `
    <div class="card flush">
      <button class="collapse-head" aria-expanded="${S.obfOpen}" data-act="obf">
        <span style="color:var(--accent);display:flex">${svg('shieldObf', 16)}</span>
        <span style="font-weight:600">${esc(t('obf.title'))}</span>
        <span class="badge accent">${esc(t('obf.badge'))}</span>
        <span class="grow"></span>
        <span class="muted" style="font-size:12px">${S.obfOpen ? '' : esc(items.length ? t('obf.count', items.length) : t('obf.none'))}</span>
        <span class="dim" style="display:flex">${svg('chevron', 14, { sw: 1.6, cls: 'chev' })}</span>
      </button>
      ${S.obfOpen ? `
        <div class="collapse-body">
          <div class="dim" style="font-size:12px;line-height:18px">${esc(t('obf.hint'))}</div>
          ${items.length ? `<div class="obf-grid">${items.map((o) => `
            <div class="obf-item"><span class="k">${esc(label(o.key))}</span><span class="v mono selectable">${esc(o.value === 'set' ? t('obf.set') : o.value)}</span></div>`).join('')}</div>` : ''}
        </div>` : ''}
    </div>`;
}

function renderPeers(d, st) {
  const live = st === 'connected';
  const stats = S.stats[d.name];
  return (d.peers || []).map((p, idx) => {
    const ps = live && stats && stats.peers ? stats.peers.find((x) => x.publicKey === p.publicKey) : null;
    const hs = ps ? ps.handshake : p.handshake;
    const open = !!S.showPsk[idx];
    let psk;
    if (!p.hasPsk) psk = `<span class="muted">${esc(t('peer.none'))}</span>`;
    else if (!p.presharedKey) psk = `<span class="muted">${esc(t('peer.hidden'))}</span>`;
    else psk = `<span class="ellipsis grow mono selectable" style="letter-spacing:${open ? 'normal' : '0.08em'}">${open ? esc(p.presharedKey) : '••••••••••••••••••••'}</span>
      <button class="btn xs" data-act="psk" data-idx="${idx}">${svg('eye', 13)}${esc(open ? t('peer.hide') : t('peer.show'))}</button>`;
    const ka = !p.keepalive || p.keepalive === '0' || p.keepalive === 'off' ? t('peer.off') : (/^\d+$/.test(p.keepalive) ? t('peer.every', p.keepalive) : p.keepalive);
    const traffic = live && ps ? `↓ ${fmtBytes(ps.rx)} · ↑ ${fmtBytes(ps.tx)}` : '—';
    return `
      <div class="card">
        <div class="peer-head">
          <h3>${esc(t('peer.title', idx + 1))}</h3>
          <span class="mono muted ellipsis" style="font-size:12px">${esc(p.endpoint || '')}</span>
          <span class="grow"></span>
          <span class="peer-hs"><span class="dot" style="background:${live && hs ? 'var(--green)' : 'var(--gray)'}"></span><span data-live="peerhs" data-idx="${idx}">${esc(live ? fmtAgo(hs) : t('peer.noData'))}</span></span>
        </div>
        <div class="table">
          <span>${esc(t('peer.publicKey'))}</span><span class="val"><span class="ellipsis grow mono selectable">${esc(p.publicKey)}</span>${copyBtn(p.publicKey, t('peer.publicKey'))}</span>
          <span>${esc(t('peer.psk'))}</span><span class="val">${psk}</span>
          <span>${esc(t('peer.endpoint'))}</span><span class="mono selectable">${esc(p.endpoint || '—')}</span>
          <span>${esc(t('peer.allowed'))}</span><span class="mono selectable" style="word-break:break-all">${esc((p.allowedIps || []).join(', ') || '—')}</span>
          <span>${esc(t('peer.keepalive'))}</span><span>${esc(ka)}</span>
          <span>${esc(t('peer.traffic'))}</span><span class="tnum" data-live="peertraffic" data-idx="${idx}">${esc(traffic)}</span>
        </div>
      </div>`;
  }).join('');
}

function renderDetails() {
  const tn = tunnelByName(S.selected) || S.tunnels[0];
  if (!tn) return '';
  const st = stateOf(tn);
  const d = S.detailsFor === tn.name ? S.details : null;
  const ep = d && d.peers && d.peers[0] ? d.peers[0].endpoint : '';
  const addr = d ? (d.interface.addresses || []).join(', ') : '';
  const busy = st === 'connecting' || st === 'disconnecting';
  const errMsg = tn.error || (d && d.error);
  return `
    <section class="details" aria-label="${esc(tn.name)}" data-scroll="details-${esc(tn.name)}">
      <div class="details-inner st-${st}">
        <div class="det-head">
          <button class="picker" data-act="picker" aria-haspopup="listbox" id="picker-btn">
            <span class="dot st st-${st}"></span><b class="ellipsis">${esc(tn.name)}</b>
            <span class="muted ellipsis" style="font-size:12px">${esc(t(tn.full ? 'type.full' : 'type.partial'))} · ${esc(t('list.others', S.tunnels.length - 1))}</span>
            <span class="grow"></span><span class="dim" style="display:flex">${svg('chevron', 14, { sw: 1.6 })}</span>
          </button>
          <div class="det-title"><h1 class="ellipsis selectable">${esc(tn.name)}</h1><span class="badge">${esc(t(tn.full ? 'type.full' : 'type.partial'))}</span></div>
          ${S.admin ? `
            <button class="btn" data-act="edit" data-name="${esc(tn.name)}">${svg('edit', 14)}${esc(t('det.edit'))}</button>
            <button class="btn icon" data-act="exportOne" data-name="${esc(tn.name)}" aria-label="${esc(t('det.export'))}" title="${esc(t('det.export'))}">${svg('export', 15)}</button>
            <button class="btn icon red" data-act="deleteOne" data-name="${esc(tn.name)}" aria-label="${esc(t('det.delete'))}" title="${esc(t('det.delete'))}">${svg('trash', 15)}</button>` : ''}
        </div>

        <div class="card hero">
          <button class="power" data-act="toggle" data-name="${esc(tn.name)}" aria-label="${esc(t('hero.action.' + (st === 'unknown' ? 'idle' : st)))}" title="${esc(t('hero.action.' + (st === 'unknown' ? 'idle' : st)))}">
            ${busy ? `<svg class="spin" aria-hidden="true" width="112" height="112" viewBox="0 0 112 112"><circle cx="56" cy="56" r="54" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-dasharray="70 270"/></svg>` : ''}
            ${svg('power', 34)}
          </button>
          <div class="hero-body">
            <div>
              <div class="hero-state"><span class="label">${esc(t('st.' + st))}</span>${st === 'connected' && S.since[tn.name] ? `<span class="tnum" data-live="uptime">${fmtDuration(Date.now() - S.since[tn.name])}</span>` : ''}</div>
              <div class="hero-sub">${esc(heroSub(st, d, tn))}</div>
            </div>
            <div class="kvgrid">
              <span>${esc(t('hero.endpoint'))}</span><span class="mono ellipsis selectable">${esc(ep || '—')}</span>
              <span>${esc(t('hero.addr'))}</span><span class="mono ellipsis selectable">${esc(addr || '—')}</span>
              <span>${esc(t('hero.extip'))}</span>${extIPCell(tn.name, st)}
            </div>
          </div>
        </div>

        ${st === 'error' && errMsg ? `
          <div class="alert" role="alert">
            <span style="color:var(--red);display:flex;margin-top:1px">${svg('warn', 18)}</span>
            <div class="grow">
              <div class="t">${esc(t('err.title'))}</div>
              <div class="d">${esc(errMsg)}</div>
              <div class="d">${esc(t('err.hint'))}</div>
              <div class="actions">
                <button class="btn sm primary" data-act="toggle" data-name="${esc(tn.name)}">${esc(t('err.retry'))}</button>
                <button class="btn sm" data-act="nav" data-screen="log">${esc(t('err.openLog'))}</button>
              </div>
            </div>
          </div>` : ''}

        <div id="stats-slot">${renderStats(tn.name, st)}</div>
        ${d ? renderInterface(d) + renderObfuscation(d) + renderPeers(d, st) : ''}
        ${!S.admin ? `<div class="notice">${esc(t('det.readonly'))}</div>` : ''}
      </div>
    </section>`;
}

/* ----- empty ----- */

function renderEmpty() {
  return `
    <section class="center" aria-label="${esc(t('empty.title'))}">
      <div class="empty">
        <div class="empty-icon">${svg('fileDown', 28, { sw: 1.2 })}</div>
        <h1>${esc(t('empty.title'))}</h1>
        ${S.admin ? `
          <p>${esc(t('empty.text'))}</p>
          <button class="btn primary" style="margin-top:16px;height:36px;padding:0 18px" data-act="importFile">${svg('folder', 15, { sw: 1.6 })}${esc(t('empty.import'))}</button>
          <div class="row-btns">
            <button class="btn sm" data-act="paste">${esc(t('empty.paste'))}</button>
            <button class="btn sm" data-act="qrScreen">${esc(t('empty.qr'))}</button>
            <button class="btn sm" data-act="newTunnel">${esc(t('empty.new'))}</button>
          </div>` : `<p>${esc(t('empty.readonly'))}</p>`}
        <div class="muted" style="margin-top:12px;font-size:11px;line-height:16px">${esc(t('empty.support'))}</div>
      </div>
    </section>`;
}

/* ----- log ----- */

function logLevel(m) {
  const s = m.toLowerCase();
  if (/\b(error|failed|fatal|panic|unable to|cannot|could not)\b/.test(s)) return 'error';
  if (/\b(warn|warning|retrying|did not complete|timeout|timed out)\b/.test(s)) return 'warn';
  return 'info';
}
function filteredLog() {
  const q = S.log.query.trim().toLowerCase();
  return S.log.lines.filter((l) => (S.log.level === 'all' || l.lv === S.log.level) && (!q || l.m.toLowerCase().includes(q)));
}
function logLineHtml(l) {
  return `<div class="log-cols log-line ${l.lv} mono"><span class="t">${fmtTime(l.t)}</span><span class="lv">${esc(t('log.lv.' + l.lv))}</span><span class="m">${esc(l.m)}</span></div>`;
}
function renderLog() {
  const lines = filteredLog();
  const levels = ['all', 'info', 'warn', 'error'];
  return `
    <section class="page" aria-label="${esc(t('log.title'))}">
      <div class="page-head">
        <div class="grow"><h1>${esc(t('log.title'))}</h1><div class="sub" id="log-count">${esc(t('log.sub', S.log.lines.length))}</div></div>
        <button class="btn" data-act="copyLog">${svg('copy', 14)}${esc(t('log.copy'))}</button>
        <button class="btn primary" data-act="saveLog">${svg('download', 14, { sw: 1.6 })}${esc(t('log.save'))}</button>
      </div>
      <div class="toolbar">
        <label class="field">${svg('search', 14, { sw: 1.6 })}<input data-key="logq" data-input="logq" placeholder="${esc(t('log.filter'))}" aria-label="${esc(t('log.filter'))}" value="${esc(S.log.query)}"></label>
        <div class="seg" role="radiogroup" aria-label="${esc(t('log.col.level'))}">
          ${levels.map((l) => `<button role="radio" aria-checked="${S.log.level === l}" data-act="logLevel" data-level="${l}">${esc(t('log.' + l))}</button>`).join('')}
        </div>
        <span class="grow"></span>
        <button class="inline-check" data-act="logAuto"><span class="check sm" role="checkbox" aria-checked="${S.log.auto}">${svgCheck(10)}</span>${esc(t('log.auto'))}</button>
      </div>
      <div class="log">
        <div class="log-cols log-head"><span>${esc(t('log.col.time'))}</span><span>${esc(t('log.col.level'))}</span><span>${esc(t('log.col.msg'))}</span></div>
        <div class="log-body" id="log-body">${lines.length ? lines.map(logLineHtml).join('') : `<div class="list-empty">${esc(t('log.empty'))}</div>`}</div>
      </div>
    </section>`;
}
let logAtBottom = true;
function scrollLogIfNeeded(force) {
  const el = $('#log-body');
  if (el && S.log.auto && (force || logAtBottom)) el.scrollTop = el.scrollHeight;
}

/* ----- settings ----- */

function setRow(label, hint, control) {
  return `<div class="set-row"><div class="l"><b>${esc(label)}</b><span>${esc(hint)}</span></div>${control}</div>`;
}
function sw(on, act, label, disabled) {
  return `<button class="switch" role="switch" aria-checked="${!!on}" aria-label="${esc(label)}" data-act="${act}" ${disabled ? 'disabled' : ''}></button>`;
}
function renderSettings() {
  const s = S.settings;
  const themes = ['dark', 'light', 'system'];
  const tunnelOpts = [`<option value="">${esc(t('set.noTunnel'))}</option>`].concat(S.tunnels.map((x) => `<option value="${esc(x.name)}" ${x.name === s.autoConnectTunnel ? 'selected' : ''}>${esc(x.name)}</option>`)).join('');
  return `
    <section class="page scroll" aria-label="${esc(t('set.title'))}" data-scroll="settings">
      <div class="settings">
        <h1 style="font-size:20px;line-height:28px;font-weight:600">${esc(t('set.title'))}</h1>
        <div class="group"><h2>${esc(t('set.look'))}</h2><div class="group-box">
          ${setRow(t('set.theme'), t('set.theme.hint'), `<div class="seg" role="radiogroup" aria-label="${esc(t('set.theme'))}">${themes.map((x) => `<button role="radio" aria-checked="${s.theme === x}" data-act="theme" data-value="${x}">${esc(t('set.theme.' + x))}</button>`).join('')}</div>`)}
          ${setRow(t('set.lang'), t('set.lang.hint'), `<select class="select" data-key="lang" data-change="lang" aria-label="${esc(t('set.lang'))}"><option value="ru" ${s.language === 'ru' ? 'selected' : ''}>Русский</option><option value="en" ${s.language === 'en' ? 'selected' : ''}>English</option></select>`)}
        </div></div>
        <div class="group"><h2>${esc(t('set.startup'))}</h2><div class="group-box">
          ${setRow(t('set.autostart'), S.admin ? t('set.autostart.hint') : t('set.autostart.admin'), sw(S.autostart, 'autostart', t('set.autostart'), !S.admin || !S.autostartKnown))}
          ${setRow(t('set.minimized'), t('set.minimized.hint'), sw(s.startMinimized, 'setting', t('set.minimized')).replace('data-act="setting"', 'data-act="setting" data-field="startMinimized"'))}
          ${setRow(t('set.autoconnect'), t('set.autoconnect.hint'), `<div style="display:flex;align-items:center;gap:10px"><select class="select sm" data-key="autotn" data-change="autoTunnel" aria-label="${esc(t('set.autoconnect'))}" ${s.autoConnect ? '' : 'disabled'}>${tunnelOpts}</select>${sw(s.autoConnect, 'setting', t('set.autoconnect')).replace('data-act="setting"', 'data-act="setting" data-field="autoConnect"')}</div>`)}
          ${setRow(t('set.notify'), t('set.notify.hint'), sw(s.notifications, 'setting', t('set.notify')).replace('data-act="setting"', 'data-act="setting" data-field="notifications"'))}
        </div></div>
        <div class="group"><h2>${esc(t('set.privacy'))}</h2><div class="group-box">
          ${setRow(t('set.extip'), t('set.extip.hint'), sw(s.externalIP, 'setting', t('set.extip')).replace('data-act="setting"', 'data-act="setting" data-field="externalIP"'))}
        </div></div>
      </div>
    </section>`;
}

/* ----- about ----- */

function renderAbout() {
  const a = S.about || {};
  const versions = [['about.app', a.app], ['about.wintun', a.wintun], ['about.go', a.go], ['about.os', a.os], ['about.arch', a.arch]];
  let upd = '';
  if (S.upd && !S.upd.complete && !S.upd.error) {
    const pct = S.upd.total ? Math.round((S.upd.done / S.upd.total) * 100) : 0;
    upd = `
      <div class="upd-banner">
        <div class="upd-row"><span class="upd-icon">${svg('download', 18, { sw: 1.6 })}</span>
          <div class="grow"><div style="font-weight:600">${esc(t('about.upd.installing'))}</div><div class="dim" style="font-size:12px;line-height:18px">${esc(t('about.upd.restart'))}</div></div></div>
        <div style="display:flex;flex-direction:column;gap:6px">
          <div class="progress" role="progressbar" aria-valuenow="${pct}" aria-valuemin="0" aria-valuemax="100"><div style="width:${pct}%"></div></div>
          <div class="dim tnum" style="display:flex;font-size:11px;line-height:16px"><span class="grow">${esc(t('about.upd.progress', S.upd.activity || '', fmtBytes(S.upd.done || 0), fmtBytes(S.upd.total || 0)))}</span><span>${pct}%</span></div>
        </div>
      </div>`;
  } else if (S.upd && S.upd.complete) {
    upd = `<div class="upd-banner"><div class="upd-row"><span class="upd-icon">${svg('ok', 18, { sw: 2 })}</span><div class="grow" style="font-weight:600">${esc(t('about.upd.restarting'))}</div></div></div>`;
  } else if (S.upd && S.upd.error) {
    upd = `<div class="alert"><span style="color:var(--red);display:flex">${svg('warn', 18)}</span><div class="d">${esc(t('about.upd.error', S.upd.error))}</div></div>`;
  } else if (S.update === 'available') {
    upd = `
      <div class="upd-banner"><div class="upd-row"><span class="upd-icon">${svg('download', 18, { sw: 1.6 })}</span>
        <div class="grow"><div style="font-weight:600">${esc(t('about.upd.title'))}</div><div class="dim" style="font-size:12px;line-height:18px">${esc(t('about.upd.sub'))}</div></div>
        ${S.admin ? `<button class="btn primary" data-act="update">${esc(t('about.upd.now'))}</button>` : ''}</div></div>`;
  } else if (S.update === 'unofficial') {
    upd = `<div class="notice">${esc(t('about.unofficial'))}</div>`;
  }
  return `
    <section class="page scroll" aria-label="${esc(t('nav.about'))}" data-scroll="about">
      <div class="about">
        ${upd}
        <div class="brand">${logoSvg(64)}<div><h1>NET PORTAL</h1><div class="dim">${esc(t('about.tagline'))}</div></div></div>
        <div class="versions">${versions.map(([k, v]) => `<span>${esc(t(k))}</span><span class="mono selectable">${esc(v || '—')}</span>`).join('')}</div>
      </div>
    </section>`;
}

/* ----- overlays ----- */

function renderOverlay() {
  const o = S.overlay;
  if (!o) return '';
  if (o.kind === 'add') {
    const items = [
      ['importFile', 'folder', 'add.file', 'add.file.hint', 'Ctrl+O'],
      ['paste', 'paste', 'add.paste', 'add.paste.hint', 'Ctrl+V'],
      ['qrMenu', 'qr', 'add.qr', 'add.qr.hint', ''],
      ['newTunnel', 'newFile', 'add.new', 'add.new.hint', 'Ctrl+N'],
    ];
    return `
      <div class="menu" role="menu" style="top:${o.y}px;left:${o.x}px">
        ${items.map(([act, ic, l, h, k]) => `
          <button class="menu-item" role="menuitem" data-act="${act}">
            <span class="ico">${svg(ic, 15)}</span><span class="tx"><b>${esc(t(l))}</b><span>${esc(t(h))}</span></span><span class="kbd">${k}</span>
          </button>
          ${act === 'qrMenu' && o.qr ? `<div class="menu-sub"><button class="btn sm" data-act="qrScreen">${svg('monitor', 14)}${esc(t('add.qr.screen'))}</button><button class="btn sm" data-act="qrFile">${svg('image', 14)}${esc(t('add.qr.file'))}</button></div>` : ''}`).join('')}
      </div>`;
  }
  if (o.kind === 'picker') {
    return `
      <div class="menu" role="listbox" style="top:${o.y}px;left:${o.x}px;width:${o.w}px;max-height:60vh;overflow:auto">
        ${S.tunnels.map((x) => `<button class="menu-item compact" role="option" aria-selected="${x.name === S.selected}" data-act="select" data-name="${esc(x.name)}"><span class="dot st st-${stateOf(x)}"></span><span class="tx"><b>${esc(x.name)}</b></span><span class="kbd">${esc(t('st.' + stateOf(x)))}</span></button>`).join('')}
      </div>`;
  }
  if (o.kind === 'delete' || o.kind === 'leave') {
    const del = o.kind === 'delete';
    const title = del ? (o.names.length === 1 ? t('del.title.one', o.names[0]) : t('del.title.many', plural(o.names.length, 'tunnels'))) : t('ed.leave');
    return `
      <div class="scrim" data-act="closeOverlay" data-self="1">
        <div class="dialog" role="dialog" aria-modal="true" aria-labelledby="dlg-title">
          <div class="dialog-body">
            <span class="dialog-icon">${svg(del ? 'trash' : 'warn', 18)}</span>
            <div><h2 id="dlg-title">${esc(title)}</h2><p>${esc(del ? t('del.text') : t('ed.leave.text'))}</p></div>
          </div>
          <div class="dialog-foot">
            <button class="btn" data-act="closeOverlay" data-key="dlg-cancel">${esc(t('cancel'))}</button>
            <button class="btn danger" data-act="${del ? 'confirmDelete' : 'confirmLeave'}">${esc(del ? t('del.confirm') : t('ed.leave.ok'))}</button>
          </div>
        </div>
      </div>`;
  }
  return '';
}

/* ------------------------------------------------------------ the editor -- */

const KNOWN_KEYS = ['PrivateKey', 'ListenPort', 'Address', 'DNS', 'MTU', 'Table', 'PreUp', 'PostUp', 'PreDown', 'PostDown',
  'Jc', 'Jmin', 'Jmax', 'S1', 'S2', 'S3', 'S4', 'H1', 'H2', 'H3', 'H4', 'I1', 'I2', 'I3', 'I4', 'I5',
  'HeaderProtectionKey', 'ContentPaddingAddition', 'RekeyAfterTime', 'RekeyTimeout', 'RejectAfterTime', 'KeepaliveTimeout',
  'MaxHandshakeAttempts', 'RandomTrailers', 'DisableCookies',
  'PublicKey', 'PresharedKey', 'AllowedIPs', 'Endpoint', 'PersistentKeepalive'];

function levenshtein(a, b) {
  a = a.toLowerCase(); b = b.toLowerCase();
  const d = Array.from({ length: a.length + 1 }, (_, i) => [i]);
  for (let j = 1; j <= b.length; j++) d[0][j] = j;
  for (let i = 1; i <= a.length; i++) for (let j = 1; j <= b.length; j++) {
    d[i][j] = Math.min(d[i - 1][j] + 1, d[i][j - 1] + 1, d[i - 1][j - 1] + (a[i - 1] === b[j - 1] ? 0 : 1));
  }
  return d[a.length][b.length];
}
function suggestKey(k) {
  let best = null, bestD = 3;
  for (const c of KNOWN_KEYS) { const dd = levenshtein(k, c); if (dd < bestD) { best = c; bestD = dd; } }
  return best && best !== k ? best : null;
}

function openEditor(opts) {
  // opts: {oldName, name, text, dirty}
  S.editor = { oldName: opts.oldName || '', name: opts.name || '', text: opts.text || '', saved: opts.dirty ? null : (opts.text || ''),
    savedName: opts.dirty ? null : (opts.name || ''), analysis: null, nameError: null, saving: false, line: 1, col: 1 };
  S.overlay = null;
  S.screen = 'editor';
  render();
}
function editorDirty() {
  const e = S.editor;
  return !!e && (e.saved === null || e.text !== e.saved || e.name !== e.savedName);
}

function mountEditor() {
  const e = S.editor;
  const root = $('#editor-root');
  if (!e || !root) return;
  root.innerHTML = `
    <div class="ed-head">
      <button class="btn icon" data-act="editorCancel" aria-label="${esc(t('cancel'))}">${svg('back', 14, { sw: 1.6 })}</button>
      <div class="grow"><div class="crumbs">${esc(e.oldName ? t('ed.crumb', e.oldName) : t('ed.crumbNew'))}</div><h1>${esc(t(e.oldName ? 'ed.title.edit' : 'ed.title.new'))}</h1></div>
      <span class="unsaved" id="ed-unsaved"></span>
    </div>
    <div class="ed-body">
      <div class="ed-fields">
        <div style="display:flex;flex-direction:column;gap:4px">
          <label class="lbl" for="ed-name">${esc(t('ed.name'))}</label>
          <input id="ed-name" class="input" data-key="edname" maxlength="32" spellcheck="false" autocomplete="off" value="${esc(e.name)}">
          <span class="hint" id="ed-name-msg"></span>
        </div>
        <div style="display:flex;flex-direction:column;gap:4px;min-width:0">
          <span class="lbl">${esc(t('ed.pub'))}</span>
          <div class="readonly"><span class="mono ellipsis selectable" id="ed-pub"></span><button class="icon-btn plain" id="ed-pub-copy" aria-label="${esc(t('copy'))}">${svg('copy', 13)}</button></div>
          <span class="hint">${esc(t('ed.pub.hint'))}</span>
        </div>
      </div>
      <div class="code">
        <div class="code-bar"><b>${esc(t('ed.config'))}</b><span>${esc(t('ed.meta'))}</span><span class="grow"></span><span class="mono" id="ed-pos"></span></div>
        <div class="code-area">
          <div class="gutter" id="ed-gutter"></div>
          <div class="code-scroll">
            <div class="code-lines" id="ed-lines"></div>
            <pre id="ed-pre" aria-hidden="true"></pre>
            <textarea id="ed-text" spellcheck="false" autocomplete="off" autocapitalize="off" wrap="off" aria-label="${esc(t('ed.config'))}"></textarea>
          </div>
        </div>
        <div id="ed-problem"></div>
      </div>
      <div class="ks" id="ed-ks"></div>
    </div>
    <div class="ed-foot">
      <span class="hint mono">${esc(t('ed.keys'))}</span>
      <button class="btn" data-act="editorCancel">${esc(t('cancel'))}</button>
      <button class="btn primary" data-act="editorSave" id="ed-save">${esc(t('ed.save'))}</button>
    </div>`;

  const ta = $('#ed-text');
  ta.value = e.text;
  const name = $('#ed-name');
  name.addEventListener('input', () => { e.name = name.value; e.nameError = null; renderEditorChrome(); });
  ta.addEventListener('input', () => { e.text = ta.value; paintCode(null); analyzeSoon(); renderEditorChrome(); });
  ta.addEventListener('scroll', syncScroll);
  ['keyup', 'click', 'select'].forEach((ev) => ta.addEventListener(ev, updateCaret));
  ta.addEventListener('keydown', (ev) => {
    if (ev.key === 'Tab' && !ev.ctrlKey && !ev.altKey) {
      ev.preventDefault();
      document.execCommand('insertText', false, '\t');
    }
  });
  $('#ed-pub-copy').addEventListener('click', () => {
    const pub = e.analysis && e.analysis.publicKey;
    if (pub) copyText(pub, t('ed.pub'));
  });
  paintCode(null);
  analyzeNow();
  renderEditorChrome();
  (e.oldName ? ta : name).focus();
}

let analyzeTimer = null;
function analyzeSoon() { clearTimeout(analyzeTimer); analyzeTimer = setTimeout(analyzeNow, 90); }
function analyzeNow() {
  const e = S.editor;
  if (!e) return;
  const text = e.text;
  np.call('analyze', { text }).then((res) => {
    if (S.editor !== e || e.text !== text) return;
    e.analysis = res;
    paintCode(res);
    renderEditorChrome();
  }).catch(() => {});
}

function syncScroll() {
  const ta = $('#ed-text');
  if (!ta) return;
  $('#ed-pre').scrollTop = ta.scrollTop; $('#ed-pre').scrollLeft = ta.scrollLeft;
  $('#ed-gutter').scrollTop = ta.scrollTop; $('#ed-lines').scrollTop = ta.scrollTop;
}
function updateCaret() {
  const ta = $('#ed-text');
  const e = S.editor;
  if (!ta || !e) return;
  const before = ta.value.slice(0, ta.selectionStart);
  const line = before.split('\n').length;
  const col = before.length - before.lastIndexOf('\n');
  if (line !== e.line || col !== e.col) {
    e.line = line; e.col = col;
    const pos = $('#ed-pos');
    if (pos) pos.textContent = t('ed.pos', line, col);
    paintLines();
  }
}

function paintLines() {
  const e = S.editor;
  const n = e.text.split('\n').length;
  const probs = new Map(((e.analysis && e.analysis.problems) || []).map((p) => [p.line, p]));
  let g = '', bg = '';
  for (let i = 1; i <= n; i++) {
    const p = probs.get(i);
    const cls = p ? (p.warning ? 'warn' : 'err') : (i === e.line ? 'cur' : '');
    g += `<div class="${cls}">${i}</div>`;
    bg += `<div class="${p && !p.warning ? 'err' : (i === e.line ? 'cur' : '')}"></div>`;
  }
  $('#ed-gutter').innerHTML = g + '<div style="height:40px"></div>';
  $('#ed-lines').innerHTML = bg + '<div style="height:40px"></div>';
  syncScroll();
}

function paintCode(analysis) {
  const e = S.editor;
  const pre = $('#ed-pre');
  if (!pre) return;
  if (analysis && analysis.tokens) {
    pre.innerHTML = analysis.tokens.map((tk) => `<span class="k-${tk.k}">${esc(tk.t)}</span>`).join('') + '\n\n';
  } else {
    pre.textContent = e.text + '\n\n';
  }
  paintLines();
}

function renderEditorChrome() {
  const e = S.editor;
  if (!e || !$('#editor-root')) return;
  const a = e.analysis;
  const dirty = editorDirty();
  $('#ed-unsaved').innerHTML = dirty ? `<span class="dot"></span>${esc(t('ed.unsaved'))}` : '';
  $('#ed-pub').textContent = a ? (a.publicKey || t('ed.pub.unknown')) : '';
  const pos = $('#ed-pos');
  if (pos) pos.textContent = t('ed.pos', e.line, e.col);

  const nameEl = $('#ed-name');
  const msg = $('#ed-name-msg');
  const exists = e.name && e.name.toLowerCase() !== e.oldName.toLowerCase() && S.tunnels.some((x) => x.name.toLowerCase() === e.name.toLowerCase());
  const nameErr = e.nameError || (exists ? t('code.nameExists') : null);
  nameEl.classList.toggle('invalid', !!nameErr);
  nameEl.setAttribute('aria-invalid', nameErr ? 'true' : 'false');
  msg.className = 'hint' + (nameErr ? ' err' : '');
  msg.textContent = nameErr || t('ed.name.hint');

  const prob = $('#ed-problem');
  const p = a && a.problems && a.problems[0];
  if (p) {
    let text;
    let fix = null;
    if (p.kind === 'unknownKey') {
      const sug = suggestKey(p.text);
      text = t('ed.unknownKey', p.text) + (sug ? ' ' + t('ed.didYouMean', sug) : '');
      if (sug) fix = { line: p.line, from: p.text, to: sug };
    } else if (p.kind === 'badSection') text = t('ed.badSection', p.text);
    else if (p.kind === 'warning') text = t('ed.warning', p.text);
    else text = t('ed.badValue', p.text.length > 48 ? p.text.slice(0, 48) + '…' : p.text);
    prob.innerHTML = `
      <div class="problem${p.warning ? ' warn' : ''}" role="alert">
        <span style="color:var(${p.warning ? '--yellow' : '--red'});display:flex">${svg('alertCircle', 14, { sw: 1.6 })}</span>
        <span class="grow"><b>${esc(t('ed.line', p.line))}</b> ${esc(text)}</span>
        ${fix ? `<button class="btn xs" id="ed-fix">${esc(t('ed.fix'))}</button>` : ''}
      </div>`;
    if (fix) $('#ed-fix').addEventListener('click', () => applyFix(fix));
  } else prob.innerHTML = '';

  const ks = $('#ed-ks');
  if (a && a.killSwitch !== 'na') {
    const on = a.killSwitch === 'on';
    ks.innerHTML = `<button class="check" role="checkbox" aria-checked="${on}" aria-labelledby="ks-l" id="ks-btn">${svgCheck(11)}</button><div><b id="ks-l">${esc(t('ed.ks'))}</b><span>${esc(t('ed.ks.hint'))}</span></div>`;
    $('#ks-btn').addEventListener('click', () => toggleKillSwitch(!on));
  } else ks.innerHTML = '';

  const save = $('#ed-save');
  const blocked = !!(a && a.hasErrors) || !!nameErr || !e.name || e.saving;
  save.setAttribute('aria-disabled', blocked ? 'true' : 'false');
}

function setEditorText(text) {
  const e = S.editor;
  const ta = $('#ed-text');
  e.text = text;
  ta.focus();
  ta.select();
  document.execCommand('insertText', false, text); // keeps undo history
  paintCode(null);
  analyzeNow();
  renderEditorChrome();
}
function applyFix(fix) {
  const lines = S.editor.text.split('\n');
  const i = fix.line - 1;
  if (lines[i] === undefined) return;
  lines[i] = lines[i].replace(fix.from, fix.to);
  setEditorText(lines.join('\n'));
}
function toggleKillSwitch(on) {
  np.call('killSwitch', { text: S.editor.text, on }).then((r) => setEditorText(r.text)).catch((err) => errorToast(t('ed.ks'), err));
}

function editorSave() {
  const e = S.editor;
  if (!e || e.saving) return;
  if ($('#ed-save').getAttribute('aria-disabled') === 'true') return;
  e.saving = true;
  np.call('save', { oldName: e.oldName, name: e.name, text: e.text }).then((r) => {
    const name = (r && r.name) || e.name;
    e.saved = e.text; e.savedName = e.name;
    toast('ok', t('toast.saved', name));
    S.selected = name;
    closeEditor(true);
    refreshList();
  }).catch((err) => {
    e.saving = false;
    if (err && /^name/.test(err.code)) { e.nameError = errText(err); renderEditorChrome(); $('#ed-name').focus(); }
    else errorToast(t('ed.save'), err);
  });
}
function closeEditor(force) {
  if (!force && editorDirty()) { S.overlay = { kind: 'leave' }; render(); return; }
  S.editor = null;
  S.overlay = null;
  S.screen = 'tunnels';
  render();
}

/* ------------------------------------------------------------ tray view -- */

function renderTray() {
  const act = activeTunnel();
  const last = S.settings.autoConnectTunnel || (S.tunnels[0] && S.tunnels[0].name);
  const cur = act || tunnelByName(S.selected) || tunnelByName(last) || S.tunnels[0];
  const st = act ? stateOf(act) : (cur && cur.error ? 'error' : 'idle');
  const on = st === 'connected' || st === 'connecting';
  const busy = st === 'connecting' || st === 'disconnecting';
  const stats = act && S.stats[act.name];
  const rate = stats && stats.rates.length ? stats.rates[stats.rates.length - 1] : null;
  const label = { connected: 'tray.connected', connecting: 'tray.connecting', disconnecting: 'tray.connecting', error: 'tray.error' }[st] || 'tray.disconnected';
  let sub = '';
  if (!cur) sub = t('tray.none');
  else if (st === 'idle') sub = t('tray.last', cur.name);
  else if (st === 'connected') {
    const d = S.detailsFor === act.name && S.details;
    const ep = d && d.peers && d.peers[0] && d.peers[0].endpoint;
    sub = cur.name + (ep ? ' · ' + ep.replace(/:\d+$/, '') : '');
  } else sub = cur.name;

  const items = S.tunnels.map((x) => {
    const xs = stateOf(x);
    const s = xs === 'connected' ? t('tray.active') : (xs === 'connecting' ? t('tray.connectingShort') : '');
    return `<button class="tray-item st-${xs}" role="radio" aria-checked="${cur && x.name === cur.name}" data-act="trayPick" data-name="${esc(x.name)}">
      <span class="dot st st-${xs}"></span><span class="grow ellipsis">${esc(x.name)}</span><span class="s">${esc(s)}</span></button>`;
  }).join('');

  return `
    <div class="tray-wrap" role="dialog" aria-label="NET PORTAL">
      <div class="tray-head">
        ${logoSvg(18)}<b>NET PORTAL</b>
        <button class="btn ghost icon" data-act="trayOpen" data-screen="settings" aria-label="${esc(t('tray.settings'))}" title="${esc(t('tray.settings'))}">${svg('settings', 16)}</button>
      </div>
      <div class="tray-card st-${st}">
        <div class="tray-status">
          <span class="tray-ring">${svg('tunnels', 20)}${busy ? `<svg class="spin" aria-hidden="true" width="48" height="48" viewBox="0 0 48 48"><circle cx="24" cy="24" r="22.5" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-dasharray="30 112"/></svg>` : ''}</span>
          <div class="grow"><div class="lbl">${esc(t(label))}</div><div class="sub ellipsis">${esc(sub)}</div></div>
          ${cur ? `<button class="switch lg${st === 'connecting' ? ' busy' : ''}" role="switch" aria-checked="${on}" aria-label="${esc(t('tray.toggle', cur.name))}" data-act="toggle" data-name="${esc(cur.name)}"></button>` : ''}
        </div>
        <div class="tray-stats tnum">
          <div><div class="k">${esc(t('tray.down'))}</div><div class="v">${rate ? fmtMbps(rate.rx) + ' ' + esc(t('stats.mbps')) : '—'}</div></div>
          <div><div class="k">${esc(t('tray.up'))}</div><div class="v">${rate ? fmtMbps(rate.tx) + ' ' + esc(t('stats.mbps')) : '—'}</div></div>
          <div><div class="k">${esc(t('tray.time'))}</div><div class="v">${st === 'connected' && S.since[act.name] ? fmtDuration(Date.now() - S.since[act.name]) : '—'}</div></div>
        </div>
      </div>
      <div class="tray-section">${esc(t('tray.tunnels'))}</div>
      <div class="tray-list" role="radiogroup" aria-label="${esc(t('tray.tunnels'))}" data-scroll="tray">${items || `<div class="list-empty">${esc(t('tray.none'))}</div>`}</div>
      <div class="tray-foot">
        <button class="btn primary" data-act="trayOpen">${esc(t('tray.open'))}</button>
        ${S.admin ? `<button class="btn" data-act="quit">${esc(t('tray.exit'))}</button>` : ''}
      </div>
    </div>`;
}

/* -------------------------------------------------------------- actions -- */

function copyText(text, label) {
  np.call('copyText', { text }).then(() => toast('ok', t('copied'), label ? t('copiedText', label) : '')).catch((e) => errorToast(t('copy'), e));
}

function toggleTunnel(name) {
  const tn = tunnelByName(name);
  if (!tn) return;
  const st = stateOf(tn);
  if (st === 'disconnecting') return;
  const stop = st === 'connected' || st === 'connecting';
  // Optimistic state so the switch reacts immediately.
  tn.state = stop ? 'stopping' : 'starting';
  tn.error = '';
  render();
  np.call(stop ? 'stop' : 'start', { name }).catch((e) => { errorToast(name, e); refreshList(); });
}

function importFile() {
  S.overlay = null; render();
  np.call('importDialog', { title: t('dlg.import') }).catch((e) => errorToast(t('toast.importFailed'), e));
}
function pasteImport() {
  S.overlay = null; render();
  np.call('clipboardText').then((r) => np.call('importText', { text: r.text })).then((r) => {
    openEditor({ name: r.name, text: r.text, dirty: true });
  }).catch((e) => errorToast(t('toast.importFailed'), e));
}
function qrImport(source) {
  S.overlay = null; render();
  np.call('importQR', { source, title: t('dlg.qr') }).then((r) => {
    if (!r || r.cancelled) return;
    toast('ok', t('toast.qrFound'), t('toast.qrFound.text'));
    openEditor({ name: r.name, text: r.text, dirty: true });
  }).catch((e) => errorToast(t('toast.importFailed'), e));
}
function newTunnel() {
  S.overlay = null;
  np.call('newConfig').then((r) => openEditor({ name: '', text: r.text, dirty: true })).catch((e) => errorToast(t('add.new'), e));
}
function editTunnel(name) {
  np.call('getConfig', { name }).then((r) => openEditor({ oldName: name, name, text: r.text })).catch((e) => errorToast(t('det.edit'), e));
}
function exportNames(names) {
  np.call('exportAll', { title: t('dlg.exportAll'), names }).then((r) => { if (r && !r.cancelled) toast('ok', t('toast.exported'), r.path); })
    .catch((e) => errorToast(t('list.exportAll'), e));
}
function confirmDelete(names) {
  if (!names.length) return;
  S.overlay = { kind: 'delete', names };
  render();
  setTimeout(() => { const c = $('[data-key="dlg-cancel"]'); if (c) c.focus(); }, 0);
}
function doDelete(names) {
  S.overlay = null;
  render();
  np.call('remove', { names }).then(() => {
    toast('ok', t('toast.deleted', names.join(', ')));
    S.checked.clear();
    S.multi = false;
    refreshList();
  }).catch((e) => { errorToast(t('del.confirm'), e); refreshList(); });
}

function openMenu(kind, anchorId, extra) {
  const a = document.getElementById(anchorId);
  if (!a) return;
  const r = a.getBoundingClientRect();
  const panel = $('.app').getBoundingClientRect();
  let x = r.left - panel.left;
  if (kind === 'add') x = Math.min(x, window.innerWidth - 320);
  S.overlay = Object.assign({ kind, x: Math.max(8, kind === 'add' ? r.left - 4 : x), y: r.bottom + 6, w: r.width }, extra || {});
  render();
  setTimeout(() => { const f = $('.menu .menu-item'); if (f) f.focus(); }, 0);
}

function setSettings(patch) {
  const next = Object.assign({}, S.settings, patch);
  S.settings = next;
  if (patch.language) LANG = patch.language;
  applyTheme();
  render();
  return np.call('setSettings', next).then((s) => { S.settings = s; }).catch((e) => errorToast(t('set.title'), e));
}

const ACTIONS = {
  nav: (el) => {
    if (S.screen === 'editor' && editorDirty()) { S.overlay = { kind: 'leave', then: el.dataset.screen }; render(); return; }
    S.editor = null;
    S.screen = el.dataset.screen; S.overlay = null;
    if (S.screen === 'log') loadLog();
    render();
  },
  select: (el) => { S.selected = el.dataset.name; S.overlay = null; S.showPsk = {}; loadDetails(); render(); },
  toggle: (el) => toggleTunnel(el.dataset.name),
  check: (el) => { const n = el.dataset.name; S.checked.has(n) ? S.checked.delete(n) : S.checked.add(n); render(); },
  multiOn: () => { S.multi = true; S.checked = new Set(S.selected ? [S.selected] : []); render(); },
  multiOff: () => { S.multi = false; S.checked.clear(); render(); },
  exportChecked: () => exportNames([...S.checked]),
  deleteChecked: () => confirmDelete([...S.checked]),
  deleteSelected: () => S.selected && confirmDelete([S.selected]),
  deleteOne: (el) => confirmDelete([el.dataset.name]),
  confirmDelete: () => doDelete(S.overlay.names),
  confirmLeave: () => {
    const then = S.overlay && S.overlay.then;
    closeEditor(true);
    if (then) { S.screen = then; if (then === 'log') loadLog(); render(); }
  },
  closeOverlay: (el, ev) => { if (el.dataset.self && ev.target !== el) return; S.overlay = null; render(); },
  exportAll: () => exportNames([]),
  exportOne: (el) => np.call('exportTunnel', { name: el.dataset.name, title: t('dlg.exportOne') })
    .then((r) => { if (r && !r.cancelled) toast('ok', t('toast.exported'), r.path); }).catch((e) => errorToast(t('det.export'), e)),
  addMenu: () => { if (S.overlay && S.overlay.kind === 'add') { S.overlay = null; render(); } else openMenu('add', 'add-btn'); },
  picker: () => { if (S.overlay && S.overlay.kind === 'picker') { S.overlay = null; render(); } else openMenu('picker', 'picker-btn'); },
  importFile, paste: pasteImport, newTunnel,
  qrMenu: () => { S.overlay = Object.assign({}, S.overlay, { qr: !S.overlay.qr }); render(); },
  qrScreen: () => qrImport('screen'),
  qrFile: () => qrImport('file'),
  edit: (el) => editTunnel(el.dataset.name),
  editorSave, editorCancel: () => closeEditor(false),
  obf: () => { S.obfOpen = !S.obfOpen; render(); },
  psk: (el) => { const i = el.dataset.idx; S.showPsk[i] = !S.showPsk[i]; render(); },
  copy: (el) => copyText(el.dataset.value, el.dataset.label),
  closeToast: (el) => { S.toasts = S.toasts.filter((x) => x.id !== el.dataset.id); renderToasts(); },
  logLevel: (el) => { S.log.level = el.dataset.level; render(); },
  logAuto: () => { S.log.auto = !S.log.auto; render(); },
  copyLog: () => copyText(filteredLog().map((l) => `${fmtTime(l.t)}  ${l.m}`).join('\r\n'), t('log.title')),
  saveLog: () => np.call('saveLog', { title: t('dlg.saveLog') }).then((r) => { if (r && !r.cancelled) toast('ok', t('toast.logSaved'), r.path); }).catch((e) => errorToast(t('log.save'), e)),
  theme: (el) => setSettings({ theme: el.dataset.value }),
  setting: (el) => setSettings({ [el.dataset.field]: !S.settings[el.dataset.field] }),
  autostart: () => {
    const on = !S.autostart;
    S.autostart = on; render();
    np.call('setAutostart', { on }).catch((e) => { S.autostart = !on; render(); errorToast(t('set.autostart'), { code: 'autostartFailed', detail: e && e.detail }); });
  },
  update: () => {
    S.upd = { done: 0, total: 0, activity: '' }; render();
    np.call('update').catch((e) => { S.upd = { error: errText(e) }; render(); });
  },
  url: (el) => np.call('openURL', { url: el.dataset.url }),
  trayOpen: (el) => np.call('showMain', { screen: el.dataset.screen || '' }),
  trayPick: (el) => {
    const n = el.dataset.name;
    const tn = tunnelByName(n);
    if (tn && stateOf(tn) !== 'connected' && stateOf(tn) !== 'connecting') toggleTunnel(n);
    S.selected = n; render();
  },
  quit: () => np.call('quit'),
};

document.addEventListener('click', (ev) => {
  const el = ev.target.closest('[data-act]');
  if (!el) {
    if (S.overlay && (S.overlay.kind === 'add' || S.overlay.kind === 'picker') && !ev.target.closest('.menu')) { S.overlay = null; render(); }
    return;
  }
  if (el.disabled || el.getAttribute('aria-disabled') === 'true') return;
  const f = ACTIONS[el.dataset.act];
  if (f) {
    if (S.overlay && (S.overlay.kind === 'add' || S.overlay.kind === 'picker') && !el.closest('.menu') && !['addMenu', 'picker'].includes(el.dataset.act)) S.overlay = null;
    f(el, ev);
  }
});

document.addEventListener('input', (ev) => {
  const k = ev.target.dataset && ev.target.dataset.input;
  if (k === 'query') { S.query = ev.target.value; render(); }
  if (k === 'logq') { S.log.query = ev.target.value; render(); }
});
document.addEventListener('change', (ev) => {
  const k = ev.target.dataset && ev.target.dataset.change;
  if (k === 'lang') setSettings({ language: ev.target.value });
  if (k === 'autoTunnel') setSettings({ autoConnectTunnel: ev.target.value });
});
document.addEventListener('scroll', (ev) => {
  if (ev.target.id === 'log-body') {
    const el = ev.target;
    logAtBottom = el.scrollHeight - el.scrollTop - el.clientHeight < 24;
  }
}, true);

// This is an app, not a browser page: no browser context menu (except the
// cut/copy/paste one in text fields) and no reload/print/devtools keys.
document.addEventListener('contextmenu', (ev) => {
  const editable = ev.target.closest && ev.target.closest('input, textarea, [contenteditable="true"]');
  if (!editable) ev.preventDefault();
});
document.addEventListener('keydown', (ev) => {
  const k = ev.key;
  const ctrl = ev.ctrlKey || ev.metaKey;
  const blocked = k === 'F5' || k === 'F3' || k === 'F7' || (k === 'F12' && !BOOT.demo) ||
    (ctrl && ['r', 'p', 'u', 'g', 'j', '+', '-', '=', '0'].includes(k.toLowerCase())) ||
    (ctrl && ev.shiftKey && ['i', 'c', 'j'].includes(k.toLowerCase()) && !BOOT.demo) ||
    (ev.altKey && (k === 'ArrowLeft' || k === 'ArrowRight'));
  if (blocked) { ev.preventDefault(); ev.stopImmediatePropagation(); }
}, true);
window.addEventListener('wheel', (ev) => { if (ev.ctrlKey) ev.preventDefault(); }, { passive: false });

document.addEventListener('keydown', (ev) => {
  const typing = ev.target.matches && ev.target.matches('input, textarea, select');
  if (ev.key === 'Escape') {
    if (S.overlay) { S.overlay = null; render(); ev.preventDefault(); return; }
    if (S.screen === 'editor') { closeEditor(false); ev.preventDefault(); return; }
    if (IS_TRAY) np.call('hidePopup');
    return;
  }
  if (IS_TRAY) return;
  const ctrl = ev.ctrlKey && !ev.altKey && !ev.shiftKey;
  if (S.screen === 'editor') {
    if (ctrl && ev.key.toLowerCase() === 's') { ev.preventDefault(); editorSave(); }
    return;
  }
  if (!ctrl || !S.admin) {
    if (ctrl && ev.key.toLowerCase() === 'f' && S.screen === 'tunnels') { ev.preventDefault(); const q = $('[data-key="query"]'); if (q) q.focus(); }
    return;
  }
  switch (ev.key.toLowerCase()) {
    case 'f': if (S.screen === 'tunnels') { ev.preventDefault(); const q = $('[data-key="query"]'); if (q) q.focus(); } break;
    case 'o': ev.preventDefault(); importFile(); break;
    case 'n': ev.preventDefault(); newTunnel(); break;
    case 'v': if (!typing && S.screen === 'tunnels') { ev.preventDefault(); pasteImport(); } break;
  }
});

/* ---- drag and drop of .conf / .zip ---- */
function hasFiles(ev) { return ev.dataTransfer && Array.from(ev.dataTransfer.types || []).includes('Files'); }
if (!IS_TRAY) {
  window.addEventListener('dragenter', (ev) => { if (!hasFiles(ev) || !S.admin) return; ev.preventDefault(); S.drop++; render(); });
  window.addEventListener('dragleave', (ev) => { if (!hasFiles(ev) || !S.admin) return; S.drop = Math.max(0, S.drop - 1); render(); });
  window.addEventListener('dragover', (ev) => { if (hasFiles(ev)) { ev.preventDefault(); ev.dataTransfer.dropEffect = S.admin ? 'copy' : 'none'; } });
  window.addEventListener('drop', (ev) => {
    if (!hasFiles(ev)) return;
    ev.preventDefault();
    S.drop = 0; render();
    if (!S.admin) return;
    const files = Array.from(ev.dataTransfer.files).filter((f) => /\.(conf|zip)$/i.test(f.name) && f.size < 8 * 1024 * 1024);
    if (!files.length) { toast('err', t('toast.importFailed'), t('code.noConfigs')); return; }
    Promise.all(files.map((f) => new Promise((resolve, reject) => {
      const r = new FileReader();
      r.onload = () => resolve({ name: f.name, data: String(r.result).split(',')[1] || '' });
      r.onerror = reject;
      r.readAsDataURL(f);
    }))).then((list) => np.call('importData', { files: list })).then(onImported).catch((e) => errorToast(t('toast.importFailed'), e));
  });
}

function onImported(res) {
  if (!res) return;
  if (res.imported > 0) {
    const title = res.imported === res.total ? t('toast.imported', plural(res.imported, 'tunnels')) : t('toast.importedSome', res.imported, res.total);
    toast(res.imported === res.total ? 'ok' : 'info', title, (res.names || []).join(', '));
    if (res.names && res.names.length) S.selected = res.names[0];
    S.screen = 'tunnels';
    refreshList();
  }
  if (res.errors && res.errors.length) toast('err', t('toast.importFailed'), res.errors.slice(0, 3).join('\n'));
}

/* --------------------------------------------------------- data loading -- */

function refreshList() {
  return np.call('list').then((list) => {
    const prev = new Map(S.tunnels.map((x) => [x.name, x]));
    S.tunnels = list || [];
    S.tunnels.forEach((x) => {
      if (x.state === 'started' && !S.since[x.name]) {
        const p = prev.get(x.name);
        if (p && p.state !== 'started') S.since[x.name] = Date.now();
      }
      if (x.state !== 'started') delete S.since[x.name];
    });
    if (!tunnelByName(S.selected)) S.selected = (activeTunnel() || S.tunnels[0] || {}).name || null;
    S.checked.forEach((n) => { if (!tunnelByName(n)) S.checked.delete(n); });
    loadDetails();
    render();
  }).catch(() => {});
}

function loadDetails() {
  const name = IS_TRAY ? (activeTunnel() || {}).name : S.selected;
  if (!name) return;
  np.call('details', { name }).then((d) => {
    if ((IS_TRAY ? (activeTunnel() || {}).name : S.selected) !== name) return;
    if (S.detailsFor !== name) S.obfOpen = false;
    S.details = d; S.detailsFor = name;
    render();
  }).catch(() => {});
}

function loadLog() {
  np.call('log', { cursor: S.log.cursor }).then((r) => {
    const fresh = (r.lines || []).map((l) => ({ t: l.t, m: l.m, lv: logLevel(l.m) }));
    S.log.cursor = r.cursor;
    if (!fresh.length) return;
    S.log.lines = S.log.lines.concat(fresh).slice(-4000);
    if (S.screen !== 'log') return;
    // Append without re-rendering the whole page to keep the selection.
    const body = $('#log-body');
    const vis = fresh.filter((l) => (S.log.level === 'all' || l.lv === S.log.level) && (!S.log.query || l.m.toLowerCase().includes(S.log.query.toLowerCase())));
    if (body && S.log.lines.length > fresh.length && !body.querySelector('.list-empty')) {
      body.insertAdjacentHTML('beforeend', vis.map(logLineHtml).join(''));
      const c = $('#log-count'); if (c) c.textContent = t('log.sub', S.log.lines.length);
      scrollLogIfNeeded(false);
    } else render();
  }).catch(() => {});
}

function fetchExternalIP(name) {
  if (!S.settings.externalIP) return;
  S.extIP[name] = { state: 'loading' };
  render();
  // Give routes a moment to settle so the request goes through the tunnel.
  setTimeout(() => {
    np.call('externalIP').then((r) => {
      S.extIP[name] = { state: 'ok', ip: r.ip };
      const tn = tunnelByName(name);
      if (tn && tn.state === 'started' && S.lastConnectToast === name) {
        // Update the connect toast text if it is still showing.
        const x = S.toasts.find((q) => q.kind === 'ok' && q.title === t('toast.connected', name));
        if (x) { x.text = t('toast.connected.ip', r.ip); renderToasts(); }
      }
      render();
    }).catch(() => { S.extIP[name] = { state: 'failed' }; render(); });
  }, 1500);
}

// Statistics: poll the runtime config of the tunnel on screen once a second.
function statsTarget() {
  if (IS_TRAY) { const a = activeTunnel(); return a && a.state === 'started' ? a.name : null; }
  if (S.screen !== 'tunnels') { const a = activeTunnel(); return a && a.state === 'started' ? a.name : null; }
  const tn = tunnelByName(S.selected);
  return tn && tn.state === 'started' ? tn.name : null;
}
function pollStats() {
  const name = statsTarget();
  if (!name || document.visibilityState === 'hidden') return;
  np.call('stats', { name }).then((r) => {
    let s = S.stats[name];
    if (!s) s = S.stats[name] = { rates: [], last: null };
    if (s.last && r.now > s.last.now) {
      const dt = (r.now - s.last.now) / 1000;
      s.rates.push({ rx: Math.max(0, (r.rx - s.last.rx) / dt), tx: Math.max(0, (r.tx - s.last.tx) / dt) });
      if (s.rates.length > 60) s.rates.shift();
    }
    s.last = r;
    s.rx = r.rx; s.tx = r.tx; s.hs = r.handshake; s.peers = r.peers;
    paintLive();
  }).catch(() => {});
}

// Cheap per-second refresh of the bits that change with time.
function paintLive() {
  if (IS_TRAY) { render(); return; }
  if (S.screen !== 'tunnels' || S.overlay) return;
  const tn = tunnelByName(S.selected);
  if (!tn) return;
  const slot = $('#stats-slot');
  if (slot) slot.innerHTML = renderStats(tn.name, stateOf(tn));
  const up = $('[data-live="uptime"]');
  if (up && S.since[tn.name]) up.textContent = fmtDuration(Date.now() - S.since[tn.name]);
  document.querySelectorAll('[data-live="rowsub"]').forEach((el) => {
    const x = tunnelByName(el.dataset.name);
    if (x) el.textContent = rowSub(x);
  });
  const s = S.stats[tn.name];
  if (s && s.peers && S.details && S.detailsFor === tn.name) {
    S.details.peers.forEach((p, i) => {
      const ps = s.peers.find((q) => q.publicKey === p.publicKey);
      if (!ps) return;
      const hs = $(`[data-live="peerhs"][data-idx="${i}"]`); if (hs) hs.textContent = fmtAgo(ps.handshake);
      const tr = $(`[data-live="peertraffic"][data-idx="${i}"]`); if (tr) tr.textContent = `↓ ${fmtBytes(ps.rx)} · ↑ ${fmtBytes(ps.tx)}`;
    });
  }
}

/* --------------------------------------------------------------- events -- */

np.on('tunnel', (ev) => {
  const tn = tunnelByName(ev.name);
  const prev = tn ? tn.state : null;
  if (!tn) { refreshList(); return; }
  tn.state = ev.state;
  tn.error = ev.error || (ev.state === 'starting' ? '' : tn.error);
  if (ev.state === 'started' && prev !== 'started') {
    S.since[ev.name] = Date.now();
    delete S.stats[ev.name];
    if (!IS_TRAY) {
      if (S.settings.notifications) { toast('ok', t('toast.connected', ev.name), ''); S.lastConnectToast = ev.name; }
      fetchExternalIP(ev.name);
    }
  }
  if (ev.state !== 'started') { delete S.since[ev.name]; delete S.extIP[ev.name]; }
  if (!IS_TRAY && S.settings.notifications) {
    if (ev.error) toast('err', t('toast.error', ev.name), ev.error);
    else if (ev.state === 'stopped' && prev === 'started') toast('info', t('toast.disconnected', ev.name), '');
  }
  if (ev.name === S.selected || IS_TRAY) loadDetails();
  render();
});
np.on('tunnels', () => refreshList());
np.on('settings', (s) => { S.settings = s; LANG = s.language || 'ru'; applyTheme(); render(); });
np.on('update', (u) => { S.update = u.state; render(); });
np.on('updateProgress', (p) => {
  S.upd = { activity: p.activity, done: p.done, total: p.total, complete: p.complete, error: p.error };
  if (S.screen === 'about' || IS_TRAY) render();
});
np.on('imported', (p) => {
  if (p.error) toast('err', t('toast.importFailed'), errText(p.error));
  else onImported(p.result);
});
np.on('navigate', (n) => {
  if (n.screen) { S.screen = n.screen; if (n.screen === 'log') loadLog(); }
  if (n.tunnel) { S.selected = n.tunnel; S.screen = 'tunnels'; loadDetails(); }
  render();
});
np.on('shown', () => { refreshList(); pollStats(); });

/* ----------------------------------------------------------------- boot -- */

function start() {
  np.call('init').then((r) => {
    S.admin = r.admin; S.demo = r.demo;
    S.settings = r.settings; LANG = r.settings.language || 'ru';
    S.autostart = r.autostart; S.autostartKnown = r.autostartKnown;
    S.update = r.update; S.about = r.about;
    S.tunnels = r.tunnels || [];
    // We do not know when tunnels that were already up started; the timer
    // starts from now for them.
    S.tunnels.forEach((x) => { if (x.state === 'started') S.since[x.name] = Date.now(); });
    S.selected = (activeTunnel() || S.tunnels[0] || {}).name || null;
    S.ready = true;
    const h = (BOOT.demo && BOOT.hints) || {};
    if (h.theme) S.settings.theme = h.theme;
    if (h.lang) { S.settings.language = h.lang; LANG = h.lang; }
    if (h.select) S.selected = h.select;
    if (h.screen === 'empty') S.tunnels = [];
    else if (h.screen === 'editor') {
      setTimeout(() => editTunnel(S.selected), 300);
      if (h.typo) setTimeout(() => setEditorText(S.editor.text.replace('Address', 'Adress')), 900);
    }
    else if (h.screen) S.screen = h.screen;
    if (h.obf) S.obfOpen = true;
    if (h.overlay === 'delete') S.overlay = { kind: 'delete', names: [S.selected] };
    if (h.overlay === 'drop') S.drop = 1;
    if (h.overlay === 'add') setTimeout(() => openMenu('add', 'add-btn', { qr: true }), 300);
    if (h.overlay === 'toasts') setTimeout(() => { toast('ok', t('toast.connected', 'ams-01'), t('toast.connected.ip', '203.0.113.24')); toast('info', t('toast.imported', plural(3, 'tunnels')), 'fra-home, tokyo-02, nyc-stream'); }, 300);
    if (h.toggle) setTimeout(() => toggleTunnel(h.toggle), 500);
    if (h.scroll) setTimeout(() => { const d = $('.details'); if (d) d.scrollTop = +h.scroll; }, 400);
    applyTheme();
    loadDetails();
    render();
    S.tunnels.filter((x) => x.state === 'started').forEach((x) => fetchExternalIP(x.name));
    setInterval(pollStats, 1000);
    setInterval(() => { if (S.screen === 'log' && !IS_TRAY) loadLog(); }, 1000);
    setInterval(() => { if (!IS_TRAY) paintLive(); }, 1000);
    if (!IS_TRAY) loadLog();
  }).catch((e) => {
    document.body.textContent = 'NET PORTAL: ' + errText(e);
  });
}
start();
