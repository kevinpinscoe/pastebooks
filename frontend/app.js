// Ensure BASE always ends with exactly one trailing slash
const BASE = (document.querySelector('base')?.getAttribute('href') || '/pastebooks/')
  .replace(/\/+$/, '') + '/';

// Join BASE + path, stripping any leading slashes from the path
const withBase = (p) => /^https?:\/\//i.test(p) ? p : BASE + p.replace(/^\/+/, '');


async function fetchJSON(u) {
  const r = await fetch(withBase(u), { credentials: 'same-origin' });
  if (!r.ok) throw new Error(await r.text());
  return r.json();
}
async function postJSON(u, body) {
  const r = await fetch(withBase(u), {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
    credentials: 'same-origin'
  });
  if (!r.ok) throw new Error(await r.text());
  return r.json();
}
async function putJSON(u, body) {
  const r = await fetch(withBase(u), {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
    credentials: 'same-origin'
  });
  if (!r.ok) throw new Error(await r.text());
  return r.json();
}
async function delJSON(u) {
  const r = await fetch(withBase(u), { method: 'DELETE', credentials: 'same-origin' });
  if (!r.ok) throw new Error(await r.text());
  return r.json();
}

// ===== tiny DOM helpers & state =====
const $  = (s, p = document) => p.querySelector(s);
const $$ = (s, p = document) => Array.from(p.querySelectorAll(s));
let state = { userId: "", books: [], currentBook: null, charms: [] };

// ===== API (all paths RELATIVE; no leading slash) =====
const api = {
  me:           () => fetchJSON('api/me'),
  login:        (email, passcode) => postJSON('api/login', { email, passcode }),
  register:     (email, passcode) => postJSON('api/register', { email, passcode }),
  logout:       () => postJSON('api/logout', {}),
  myBooks:      () => fetchJSON('api/books'),
  getBooks:      (id) => fetchJSON(`api/books/${id}`),
  saveBook:     (id, body) => putJSON(`api/books/${id}`, body),
  createBook:   (body) => postJSON('api/books', body),
  delBook:      (id) => delJSON(`api/books/${id}`),
  listCharms:   (pid) => fetchJSON(`api/books/${pid}/charms`),
  createCharm:  (pid, body) => postJSON(`api/books/${pid}/charms`, body),
  updateCharm:  (id, body) => putJSON(`api/charms/${id}`, body),
  delCharm:     (id) => delJSON(`api/charms/${id}`),
};

// ===== constants =====
const shapes = ["square","star","circle","triangle","rectangle","diamond","heart","clover","spade","hexagon","squiggle"];
const colors = ["red","green","blue","yellow","purple","pink","gold","black","orange","darkgray"];

// ===== auth handlers =====
const loginForm = $('#loginForm');
const emailEl   = $('#email');
const passEl    = $('#pass');

loginForm?.addEventListener('submit', async (e) => {
  e.preventDefault();
  const email = emailEl?.value?.trim();
  const pass  = passEl?.value || '';
  if (!email || !pass) {
    console.warn('[login] missing email or passcode');
    alert('Enter email and passcode');
    return;
  }
  try {
    console.log('[login] POST api/login');
    const res = await api.login(email, pass);
    console.log('[login] ok', res);
    // clear password field
    if (passEl) passEl.value = '';

    // verify session
    const me = await api.me();
    console.log('[me]', me);

    await boot();
  } catch (err) {
    console.error('[login] failed:', err);
    alert('Login failed. Check email/passcode and try again.');
  }
});

$('#btnRegister')?.addEventListener('click', async () => {
  const email = emailEl?.value?.trim();
  const pass  = passEl?.value || '';
  if (!email || !pass) { alert('Enter email and passcode'); return; }
  try {
    console.log('[register] POST api/register');
    const res = await api.register(email, pass);
    console.log('[register] ok', res);
    await boot();
  } catch (err) {
    console.error('[register] failed:', err);
    alert('Register failed: ' + err);
  }
});

$('#btnLogout')?.addEventListener('click', async () => {
  try {
    console.log('[logout] POST api/logout');
    await api.logout();
  } finally {
    await boot();
  }
});

// ===== book actions =====
$('#btnNewBook')?.addEventListener('click', async () => {
  const title = prompt('New book title');
  if (!title) return;
  const res = await api.createBook({ title, note:'', is_public:false });
  await loadBooks();
  await selectBook(res.id);
});

$('#btnSaveBook')?.addEventListener('click', async () => {
  if (!state.currentBook) return;
  await api.saveBook(state.currentBook.id, {
    title: $('#bookTitle').value,
    note: $('#bookNote').value,
    is_public: $('#bookPublic').checked
  });
  await selectBook(state.currentBook.id);
  await loadBooks();
});

$('#btnAddCharm')?.addEventListener('click', async () => {
  if (!state.currentBook) {
    const res = await api.createBook({ title: 'New Book', note:'', is_public:false });
    await loadBooks();
    await selectBook(res.id);
  }
  openCharmDialog();
});

// ===== rendering =====
function renderTabs() {
  const tabs = $('#tabs');
  tabs.innerHTML = '';
  state.books.forEach(b => {
    const el = document.createElement('div');
    el.className = 'tab' + (state.currentBook && b.id === state.currentBook.id ? ' active' : '');
    el.textContent = b.title;
    el.addEventListener('click', () => selectBook(b.id));
    tabs.appendChild(el);
  });
}

function renderCharms() {
  const host = $('#charms');
  if (!host) return;
  host.innerHTML = '';

  const list = Array.isArray(state.charms) ? state.charms : [];
  for (const ch of list) {
    host.appendChild(charmEl(ch));
  }

  const empty = $('#emptyCharms');
  if (empty) empty.classList.toggle('hidden', list.length > 0);
}

function charmEl(ch) {
  const tpl = $('#charmTpl').content.cloneNode(true);
  const root  = tpl.querySelector('.charm');
  const svg   = tpl.querySelector('svg');
  const title = tpl.querySelector('.ctitle');
  const btnDel= tpl.querySelector('.trash');

  svg.innerHTML = shapePath(ch.shape);
  svg.classList.add('color-' + ch.color);
  title.textContent = ch.title;
  root.title = (ch.text_value || '') + "\n(click to copy)";

  root.addEventListener('click', async (ev) => {
    if (ev.target === btnDel) return;
    await navigator.clipboard.writeText(ch.text_value || '');
  });

  btnDel.addEventListener('click', async (ev) => {
    ev.stopPropagation();
    if (!confirm('Delete this charm?')) return;
    await api.delCharm(ch.id);
    state.charms = state.charms.filter(x => x.id !== ch.id);
    renderCharms();
  });

  return tpl;
}

// ===== dialog =====
function openCharmDialog(existing) {
  const dlg  = $('#dlgCharm');
  const sSel = $('#chShape');
  const cSel = $('#chColor');

  sSel.innerHTML = shapes.map(s => `<option value="${s}">${s}</option>`).join('');
  cSel.innerHTML = colors.map(c => `<option value="${c}">${c}</option>`).join('');

  $('#chTitle').value = existing?.title || '';
  $('#chValue').value = existing?.text_value || '';
  sSel.value = existing?.shape || 'square';
  cSel.value = existing?.color || 'blue';

  dlg.showModal();
  $('#chSave').onclick = async () => {
    const body = {
      title: $('#chTitle').value,
      text_value: $('#chValue').value,
      shape: $('#chShape').value,
      color: $('#chColor').value
    };
    if (existing) await api.updateCharm(existing.id, body);
    else          await api.createCharm(state.currentBook.id, body);
    dlg.close();
    state.charms = await api.listCharms(state.currentBook.id);
    renderCharms();
  };
}

// ===== shapes =====
function shapePath(name) {
  switch (name) {
    case 'square':    return '<rect x="10" y="10" width="80" height="80" rx="8"/>';
    case 'circle':    return '<circle cx="50" cy="50" r="40"/>';
    case 'triangle':  return '<polygon points="50,10 90,90 10,90"/>';
    case 'rectangle': return '<rect x="10" y="25" width="80" height="50" rx="8"/>';
    case 'diamond':   return '<polygon points="50,5 95,50 50,95 5,50"/>';
    case 'heart':     return '<path d="M50 86 L15 50 A20 20 0 1 1 50 30 A20 20 0 1 1 85 50 Z"/>';
    case 'clover':    return '<path d="M50 60 a15 15 0 1 1 0-30 a15 15 0 1 1 30 0 a15 15 0 1 1 -30 0 a15 15 0 1 1 0 30 Z"/>';
    case 'spade':     return '<path d="M50 10 C35 30,15 45,15 60 a15 15 0 0 0 30 0 a15 15 0 0 0 30 0 C85 45,65 30,50 10 Z M42 90 h16 v-10 h-16 Z"/>';
    case 'hexagon':   return '<polygon points="30,10 70,10 90,50 70,90 30,90 10,50"/>';
    case 'star':      return '<polygon points="50,8 60,38 92,38 66,56 76,86 50,68 24,86 34,56 8,38 40,38"/>';
    case 'squiggle':  return '<path d="M10 60 C 20 20, 60 20, 50 60 S 80 100, 90 60" stroke-width="0"/>';
    default:          return '<rect x="15" y="15" width="70" height="70" rx="10"/>';
  }
}

// ===== data flows =====
async function loadBooks() {
  state.books = await api.myBooks();
  renderTabs();
}

async function selectBook(id) {
  state.currentBook = await api.getBooks(id);
  $('#bookTitle').value    = state.currentBook.title || '';
  $('#bookNote').value     = state.currentBook.note || '';
  $('#bookPublic').checked = !!state.currentBook.is_public;
  $('#createdAt').textContent = 'Created: ' + new Date(state.currentBook.created_at).toLocaleString();
  $('#updatedAt').textContent = 'Last edit: ' + new Date(state.currentBook.updated_at).toLocaleString();

  // Respect base path for public link
  const link = withBase(`api/public/books/${state.currentBook.id}`);
  $('#shareLink').innerHTML = `<a href="${link}" target="_blank">Public link</a>`;

  state.charms = await api.listCharms(id);
  renderCharms();
  renderTabs();
}

// ===== boot =====
async function boot() {
  try {
    const me = await api.me();            // { user_id, dev? }
    state.userId = me.user_id || '';

    if (state.userId) {
      $('#loginForm')?.classList.add('hidden');
      $('#meId').textContent = state.userId.slice(0,8) + (me.dev ? '… (dev)' : '…');
      $('#me')?.classList.remove('hidden');

      await loadBooks();
      if (state.books.length) await selectBook(state.books[0].id);
    } else {
      $('#loginForm')?.classList.remove('hidden');
      $('#me')?.classList.add('hidden');
      $('#tabs').innerHTML = '';
      $('#charms').innerHTML = '';
    }
  } catch (e) {
    console.error(e);
  }
}

boot();
