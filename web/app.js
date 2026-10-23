'use strict';

const state = {
  symbols: [],
  current: null,
  history: [],
};

const els = {
  updated: document.getElementById('updated'),
  tbody: document.querySelector('#quote-table tbody'),
  select: document.getElementById('symbol-select'),
  chart: document.getElementById('chart'),
};

function fmtNum(v, digits = 2) {
  if (v === null || v === undefined || Number.isNaN(Number(v))) return '—';
  return Number(v).toLocaleString('zh-CN', {
    minimumFractionDigits: digits,
    maximumFractionDigits: digits,
  });
}

function cls(v) {
  return v > 0 ? 'up' : v < 0 ? 'down' : 'flat';
}

async function loadLatest() {
  const res = await fetch('data/latest.json', { cache: 'no-store' });
  if (!res.ok) throw new Error('latest.json HTTP ' + res.status);
  const data = await res.json();
  state.symbols = data.symbols || [];
  els.updated.textContent = '数据更新时间：' + (data.generated_at || '未知');
  renderTable();
  fillSelect();
  if (state.symbols.length) selectSymbol(state.symbols[0].code);
}

function renderTable() {
  els.tbody.innerHTML = '';
  state.symbols.forEach((q) => {
    const tr = document.createElement('tr');
    tr.dataset.code = q.code;
    tr.innerHTML = `
      <td>${q.name || ''}</td>
      <td class="mono">${q.code || ''}</td>
      <td>${q.date || '—'}</td>
      <td class="mono">${fmtNum(q.close)}</td>
      <td class="${cls(q.change)}">${q.change > 0 ? '+' : ''}${fmtNum(q.change)}</td>
      <td class="${cls(q.change_pct)}">${q.change_pct > 0 ? '+' : ''}${fmtNum(q.change_pct)}%</td>
      <td class="mono">${fmtNum(q.high)}</td>
      <td class="mono">${fmtNum(q.low)}</td>
      <td class="mono">${fmtNum(q.volume, 0)}</td>
    `;
    tr.addEventListener('click', () => selectSymbol(q.code));
    els.tbody.appendChild(tr);
  });
}

function fillSelect() {
  els.select.innerHTML = '';
  state.symbols.forEach((q) => {
    const o = document.createElement('option');
    o.value = q.code;
    o.textContent = `${q.name} (${q.code})`;
    els.select.appendChild(o);
  });
}

async function selectSymbol(code) {
  state.current = code;
  els.select.value = code;
  [...els.tbody.children].forEach((tr) => {
    tr.classList.toggle('active', tr.dataset.code === code);
  });
  const res = await fetch(`data/history/${code}.json`, { cache: 'no-store' });
  if (!res.ok) {
    drawChart([], '加载失败');
    return;
  }
  state.history = await res.json();
  const name = (state.symbols.find((s) => s.code === code) || {}).name || code;
  drawChart(state.history, name);
}

function drawChart(bars, title) {
  const canvas = els.chart;
  const dpr = window.devicePixelRatio || 1;
  const w = canvas.clientWidth || canvas.parentElement.clientWidth || 800;
  const h = parseInt(canvas.getAttribute('height'), 10) || 360;
  canvas.width = Math.round(w * dpr);
  canvas.height = Math.round(h * dpr);
  const ctx = canvas.getContext('2d');
  ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
  ctx.clearRect(0, 0, w, h);

  const pad = { l: 66, r: 16, t: 20, b: 32 };
  const iw = w - pad.l - pad.r;
  const ih = h - pad.t - pad.b;

  ctx.fillStyle = '#0f172a';
  ctx.fillRect(0, 0, w, h);

  if (!bars || bars.length < 2) {
    ctx.fillStyle = '#94a3b8';
    ctx.font = '14px system-ui, sans-serif';
    ctx.fillText(title || '图表', pad.l, pad.t + 20);
    ctx.fillText('暂无足够历史数据', pad.l, h / 2);
    return;
  }

  const closes = bars.map((b) => b.close);
  const min = Math.min(...closes);
  const max = Math.max(...closes);
  const span = max - min || 1;
  const x = (i) => pad.l + (i / (bars.length - 1)) * iw;
  const y = (v) => pad.t + ih - ((v - min) / span) * ih;

  ctx.strokeStyle = '#1e293b';
  ctx.fillStyle = '#64748b';
  ctx.font = '11px system-ui, sans-serif';
  const rows = 5;
  for (let i = 0; i <= rows; i++) {
    const v = min + (span * i) / rows;
    const yy = y(v);
    ctx.beginPath();
    ctx.moveTo(pad.l, yy);
    ctx.lineTo(pad.l + iw, yy);
    ctx.stroke();
    ctx.fillText(v.toFixed(2), 6, yy + 4);
  }

  ctx.beginPath();
  bars.forEach((b, i) => {
    if (i === 0) ctx.moveTo(x(i), y(b.close));
    else ctx.lineTo(x(i), y(b.close));
  });
  ctx.strokeStyle = '#38bdf8';
  ctx.lineWidth = 2;
  ctx.stroke();

  ctx.lineTo(x(bars.length - 1), pad.t + ih);
  ctx.lineTo(x(0), pad.t + ih);
  ctx.closePath();
  ctx.fillStyle = 'rgba(56,189,248,0.12)';
  ctx.fill();

  ctx.fillStyle = '#64748b';
  ctx.fillText(bars[0].date, pad.l, h - 8);
  const last = bars[bars.length - 1].date;
  const lastW = ctx.measureText(last).width;
  ctx.fillText(last, pad.l + iw - lastW, h - 8);
}

els.select.addEventListener('change', () => selectSymbol(els.select.value));

loadLatest().catch((e) => {
  els.updated.textContent = '加载失败：' + e.message;
});
