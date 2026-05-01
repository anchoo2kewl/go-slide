// go-slide editor — orchestrates TipTap (WYSIWYG), CodeMirror (Code mode),
// and the live preview iframe. The defining design constraint is that the
// WYSIWYG canvas, the preview iframe, and the production presentation page
// all load the same theme CSS so they look identical.
//
// State model:
//   - editorState.slides[i] = { content, attrs, meta }
//     attrs = full <section> attributes (class, data-background-color, etc.)
//     content = the inner HTML (children of <section>)
//   - The TipTap document tree is the AAL/Default schema; HTML round-trips
//     losslessly because each design primitive has a Node extension.
//
// Mode switching:
//   - WYSIWYG ↔ Code: serialize current state, re-load into the other
//     editor. No content is ever passed through StarterKit's lossy parser.
//   - Preview: writes the full presentation HTML into an iframe srcdoc,
//     using the same theme CSS that ships in production.

import { Editor } from '@tiptap/core';
import StarterKit from '@tiptap/starter-kit';
import TextAlign from '@tiptap/extension-text-align';
import Underline from '@tiptap/extension-underline';
import { AalSchema } from './schema-aal.mjs';
import { DefaultSchema } from './schema-default.mjs';

// --- registry --------------------------------------------------------------

const SCHEMAS = {
  aal: AalSchema,
  default: DefaultSchema,
};

function schemaFor(themeID) {
  return SCHEMAS[themeID] || DefaultSchema;
}

// --- main editor class -----------------------------------------------------

export class GoSlideEditor {
  constructor(opts) {
    this.opts = Object.assign({
      mountSelector: '#go-slide-editor',
      previewEndpoint: '/slides/preview',
      assetBase: '/slides/assets',
      autosaveEndpoint: null,
      themeID: 'default',
      revealCDN: 'https://cdn.jsdelivr.net/npm/reveal.js@4.3.1',
      onChange: null, // (state) => void
      onSave: null,   // (combinedHTML) => Promise
    }, opts);

    this.state = {
      slides: [],
      currentIndex: 0,
      mode: 'wysiwyg', // wysiwyg | code | preview
      dirty: false,
    };

    this.tiptap = null;
    this.codeView = null;
    this.previewIframe = null;
    this.dom = {};

    this._buildShell();
    this._loadInitial(opts.initialContent || '');
    this._bindToolbar();
  }

  // ---- public API --------------------------------------------------------

  // getCombinedHTML serializes all slides back to a single HTML string
  // that matches the on-disk format. Use this when saving.
  getCombinedHTML() {
    this._captureCurrent();
    return this.state.slides.map((s) => {
      const attrStr = Object.entries(s.attrs)
        .filter(([, v]) => v != null && v !== '')
        .map(([k, v]) => `${k}="${escapeAttr(v)}"`).join(' ');
      return `<section ${attrStr}>\n${s.content}\n</section>`;
    }).join('\n\n');
  }

  // setMode switches between wysiwyg, code, preview without losing content.
  setMode(mode) {
    if (mode === this.state.mode) return;
    this._captureCurrent();
    this.state.mode = mode;
    this.dom.tiptap.style.display = mode === 'wysiwyg' ? '' : 'none';
    this.dom.code.style.display = mode === 'code' ? '' : 'none';
    this.dom.preview.style.display = mode === 'preview' ? '' : 'none';
    for (const id of ['mode-wysiwyg','mode-code','mode-preview']) {
      const el = this.dom[id];
      if (el) el.classList.toggle('active', el.dataset.mode === mode);
    }
    if (mode === 'wysiwyg') this._loadIntoTipTap();
    if (mode === 'code') this._loadIntoCodeMirror();
    if (mode === 'preview') this._loadIntoPreview();
  }

  // Replace the current slide list. Called when switching themes.
  setTheme(themeID) {
    if (this.opts.themeID === themeID) return;
    this._captureCurrent();
    this.opts.themeID = themeID;
    if (this.state.mode === 'wysiwyg') {
      // Rebuild TipTap with new schema
      this._destroyTipTap();
      this._loadIntoTipTap();
    }
    if (this.state.mode === 'preview') this._loadIntoPreview();
  }

  // ---- internals ---------------------------------------------------------

  _buildShell() {
    const root = document.querySelector(this.opts.mountSelector);
    if (!root) throw new Error(`go-slide: mount node not found: ${this.opts.mountSelector}`);
    root.innerHTML = `
      <div class="gs-editor">
        <aside class="gs-sidebar">
          <div class="gs-sidebar-head">SLIDES</div>
          <ol class="gs-thumbs" id="gs-thumbs"></ol>
          <button class="gs-add-slide" id="gs-add-slide">+ Add Slide</button>
        </aside>
        <main class="gs-main">
          <div class="gs-toolbar" id="gs-toolbar">
            <div class="gs-mode-toggle">
              <button class="gs-mode-btn active" data-mode="wysiwyg" id="mode-wysiwyg">WYSIWYG</button>
              <button class="gs-mode-btn" data-mode="code" id="mode-code">Code</button>
              <button class="gs-mode-btn" data-mode="preview" id="mode-preview">Preview</button>
            </div>
            <div class="gs-toolbar-spacer"></div>
            <div class="gs-toolbar-group" id="gs-block-toolbar">
              <button class="gs-tb-btn" data-cmd="aalH1" title="Hero heading">H1</button>
              <button class="gs-tb-btn" data-cmd="aalH2" title="Section heading">H2</button>
              <button class="gs-tb-btn" data-cmd="aalEyebrow" title="Eyebrow label">Eyebrow</button>
              <button class="gs-tb-btn" data-cmd="aalRule" title="Accent rule">Rule</button>
              <button class="gs-tb-btn" data-cmd="aalLede" title="Lede paragraph">Lede</button>
              <button class="gs-tb-btn" data-cmd="aalQuote" title="Pull quote">Quote</button>
              <button class="gs-tb-btn" data-cmd="aalGrid3" title="3-column grid of cards">Grid 3</button>
              <button class="gs-tb-btn" data-cmd="aalGrid2" title="2-column grid">Grid 2</button>
              <button class="gs-tb-btn" data-cmd="aalCard" title="Single card">Card</button>
              <button class="gs-tb-btn" data-cmd="aalStat" title="Big number stat">Stat</button>
              <button class="gs-tb-btn" data-cmd="aalPill" title="Pill chip">Pill</button>
              <button class="gs-tb-btn" data-cmd="aalFoot" title="Footer row">Foot</button>
            </div>
            <div class="gs-toolbar-spacer"></div>
            <button class="gs-tb-btn gs-save" id="gs-save">Save</button>
          </div>
          <div class="gs-canvas-wrap" id="gs-canvas-wrap">
            <div class="gs-canvas-fit" id="gs-canvas-fit">
              <div class="gs-canvas" id="gs-canvas">
                <div class="gs-tiptap" id="gs-tiptap"></div>
                <textarea class="gs-code" id="gs-code" spellcheck="false"></textarea>
                <iframe class="gs-preview" id="gs-preview"></iframe>
              </div>
            </div>
          </div>
        </main>
      </div>
    `;
    this.dom = {
      thumbs: root.querySelector('#gs-thumbs'),
      addSlide: root.querySelector('#gs-add-slide'),
      tiptap: root.querySelector('#gs-tiptap'),
      code: root.querySelector('#gs-code'),
      preview: root.querySelector('#gs-preview'),
      canvas: root.querySelector('#gs-canvas'),
      canvasFit: root.querySelector('#gs-canvas-fit'),
      canvasWrap: root.querySelector('#gs-canvas-wrap'),
      'mode-wysiwyg': root.querySelector('#mode-wysiwyg'),
      'mode-code': root.querySelector('#mode-code'),
      'mode-preview': root.querySelector('#mode-preview'),
      save: root.querySelector('#gs-save'),
      toolbar: root.querySelector('#gs-block-toolbar'),
    };
    this.dom.code.style.display = 'none';
    this.dom.preview.style.display = 'none';

    // Scale the 1280×800 canvas to fit whatever width the editor area
    // exposes. Re-runs on every layout change so resizing the window or
    // toggling sidebars stays pixel-correct.
    this._scale = () => {
      const fit = this.dom.canvasFit;
      if (!fit) return;
      const w = fit.clientWidth;
      const scale = w > 0 ? w / 1280 : 1;
      fit.style.setProperty('--gs-scale', String(scale));
    };
    this._scaleObserver = new ResizeObserver(this._scale);
    this._scaleObserver.observe(this.dom.canvasWrap);
    window.addEventListener('resize', this._scale);
    queueMicrotask(this._scale);

    // Mode buttons
    for (const id of ['mode-wysiwyg','mode-code','mode-preview']) {
      this.dom[id].addEventListener('click', () => this.setMode(this.dom[id].dataset.mode));
    }
    this.dom.addSlide.addEventListener('click', () => this._addSlide());
    this.dom.save.addEventListener('click', () => this._save());
  }

  _loadInitial(html) {
    if (!html.trim()) {
      this.state.slides = [{
        attrs: { class: this.opts.themeID === 'aal' ? 'aal pptx-slide' : 'slide pptx-slide' },
        content: '',
      }];
    } else {
      this.state.slides = parseSlides(html);
    }
    this._renderThumbs();
    this._loadIntoTipTap();
  }

  _renderThumbs() {
    this.dom.thumbs.innerHTML = this.state.slides.map((s, i) => `
      <li class="gs-thumb ${i === this.state.currentIndex ? 'active' : ''}" data-i="${i}">
        <span class="gs-thumb-num">${i + 1}</span>
        <div class="gs-thumb-preview">${escapeHTML(stripTags(s.content)).slice(0, 80)}</div>
        <button class="gs-thumb-del" title="Delete slide">×</button>
      </li>
    `).join('');
    this.dom.thumbs.querySelectorAll('.gs-thumb').forEach((el) => {
      el.addEventListener('click', (e) => {
        if (e.target.classList.contains('gs-thumb-del')) {
          this._deleteSlide(parseInt(el.dataset.i, 10));
          return;
        }
        this._switchSlide(parseInt(el.dataset.i, 10));
      });
    });
  }

  _switchSlide(i) {
    this._captureCurrent();
    this.state.currentIndex = i;
    this._renderThumbs();
    if (this.state.mode === 'wysiwyg') this._loadIntoTipTap();
    else if (this.state.mode === 'code') this._loadIntoCodeMirror();
    else this._loadIntoPreview();
  }

  _addSlide() {
    this._captureCurrent();
    const themeCls = this.opts.themeID === 'aal' ? 'aal pptx-slide' : 'slide pptx-slide';
    const wrapCls = this.opts.themeID === 'aal' ? 'aal-wrap' : 'slide-wrap';
    this.state.slides.push({
      attrs: { class: themeCls, 'data-background-color': '#0A1628' },
      content: `<div class="${wrapCls}"><h2 class="${this.opts.themeID === 'aal' ? 'aal-h2' : 'slide-h2'}">New Slide</h2></div>`,
    });
    this.state.currentIndex = this.state.slides.length - 1;
    this._renderThumbs();
    this._loadIntoTipTap();
    this.state.dirty = true;
  }

  _deleteSlide(i) {
    if (this.state.slides.length <= 1) return;
    this.state.slides.splice(i, 1);
    if (this.state.currentIndex >= this.state.slides.length) {
      this.state.currentIndex = this.state.slides.length - 1;
    }
    this._renderThumbs();
    if (this.state.mode === 'wysiwyg') this._loadIntoTipTap();
    this.state.dirty = true;
  }

  _captureCurrent() {
    const cur = this.state.slides[this.state.currentIndex];
    if (!cur) return;
    if (this.state.mode === 'wysiwyg' && this.tiptap) {
      cur.content = this.tiptap.getHTML();
    } else if (this.state.mode === 'code' && this.dom.code) {
      cur.content = this.dom.code.value;
    }
  }

  _destroyTipTap() {
    if (this.tiptap) {
      try { this.tiptap.destroy(); } catch (e) { /* noop */ }
      this.tiptap = null;
      this.dom.tiptap.innerHTML = '';
    }
  }

  _loadIntoTipTap() {
    this._destroyTipTap();
    const cur = this.state.slides[this.state.currentIndex];
    if (!cur) return;
    const extensions = [
      StarterKit.configure({
        heading: false, // we use AalH1/AalH2 instead
        horizontalRule: false, // we use AalRule
      }),
      TextAlign.configure({ types: ['paragraph'] }),
      Underline,
      ...schemaFor(this.opts.themeID),
    ];
    this.tiptap = new Editor({
      element: this.dom.tiptap,
      extensions,
      content: cur.content || '<p></p>',
      onUpdate: () => {
        this.state.dirty = true;
        if (this.opts.onChange) this.opts.onChange(this.state);
      },
    });
  }

  _loadIntoCodeMirror() {
    const cur = this.state.slides[this.state.currentIndex];
    if (!cur) return;
    this.dom.code.value = cur.content || '';
    this.dom.code.oninput = () => {
      cur.content = this.dom.code.value;
      this.state.dirty = true;
      if (this.opts.onChange) this.opts.onChange(this.state);
    };
  }

  _loadIntoPreview() {
    const html = this.getCombinedHTML();
    const themeID = this.opts.themeID;
    fetch(this.opts.previewEndpoint, {
      method: 'POST',
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      body: new URLSearchParams({ content: html, theme: themeID }).toString(),
    })
      .then((r) => r.json())
      .then((data) => { this.dom.preview.srcdoc = data.html || ''; })
      .catch((err) => { this.dom.preview.srcdoc = `<pre>preview error: ${err}</pre>`; });
  }

  _bindToolbar() {
    this.dom.toolbar.querySelectorAll('[data-cmd]').forEach((btn) => {
      btn.addEventListener('click', () => {
        if (this.state.mode !== 'wysiwyg' || !this.tiptap) return;
        const cmd = btn.dataset.cmd;
        this._runCommand(cmd);
      });
    });
  }

  _runCommand(cmd) {
    const e = this.tiptap;
    if (!e) return;
    const isAal = this.opts.themeID === 'aal';
    const ns = isAal ? 'aal' : 'slide';
    switch (cmd) {
      case 'aalH1': return e.chain().focus().insertContent({ type: isAal ? 'aalH1' : 'slideH1', content: [{ type: 'text', text: 'Heading' }] }).run();
      case 'aalH2': return e.chain().focus().insertContent({ type: isAal ? 'aalH2' : 'slideH2', content: [{ type: 'text', text: 'Section' }] }).run();
      case 'aalEyebrow': return e.chain().focus().insertContent({ type: isAal ? 'aalEyebrow' : 'slideEyebrow', content: [{ type: 'text', text: 'EYEBROW' }] }).run();
      case 'aalRule': return e.chain().focus().insertContent({ type: isAal ? 'aalRule' : 'slideRule' }).run();
      case 'aalLede': return e.chain().focus().insertContent({ type: isAal ? 'aalLede' : 'slideLede', content: [{ type: 'text', text: 'Large lede paragraph.' }] }).run();
      case 'aalQuote': return e.chain().focus().insertContent({ type: isAal ? 'aalQuote' : 'slideQuote', content: [{ type: 'text', text: 'Pull quote.' }] }).run();
      case 'aalGrid3': return e.chain().focus().insertContent(gridContent(ns, 3)).run();
      case 'aalGrid2': return e.chain().focus().insertContent(gridContent(ns, 2)).run();
      case 'aalCard':  return e.chain().focus().insertContent(cardContent(ns)).run();
      case 'aalStat':  return e.chain().focus().insertContent(statContent(ns)).run();
      case 'aalPill':  return e.chain().focus().insertContent({ type: isAal ? 'aalPill' : 'slidePill', content: [{ type: 'text', text: 'Pill' }] }).run();
      case 'aalFoot':  return e.chain().focus().insertContent({ type: isAal ? 'aalFoot' : 'slideFoot', content: [{ type: 'text', text: 'aiagentlens.com' }] }).run();
    }
  }

  _save() {
    this._captureCurrent();
    const html = this.getCombinedHTML();
    if (this.opts.onSave) {
      Promise.resolve(this.opts.onSave(html)).then(() => {
        this.state.dirty = false;
      });
    } else if (this.opts.autosaveEndpoint) {
      fetch(this.opts.autosaveEndpoint, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ content: html }),
      });
      this.state.dirty = false;
    }
  }
}

// --- helpers --------------------------------------------------------------

function gridContent(ns, cols) {
  const card = ns === 'aal' ? 'aalCard' : 'slideCard';
  const grid = ns === 'aal' ? 'aalGrid' : 'slideGrid';
  const h = ns === 'aal' ? 'aalH2' : 'slideH2';
  const cards = Array.from({ length: cols }, (_, i) => ({
    type: card,
    content: [
      { type: h, content: [{ type: 'text', text: `Card ${i + 1}` }] },
      { type: 'paragraph', content: [{ type: 'text', text: 'Body text.' }] },
    ],
  }));
  return { type: grid, attrs: { cols }, content: cards };
}

function cardContent(ns) {
  const card = ns === 'aal' ? 'aalCard' : 'slideCard';
  const h = ns === 'aal' ? 'aalH2' : 'slideH2';
  return {
    type: card,
    content: [
      { type: h, content: [{ type: 'text', text: 'Card' }] },
      { type: 'paragraph', content: [{ type: 'text', text: 'Body text.' }] },
    ],
  };
}

function statContent(ns) {
  const stat = ns === 'aal' ? 'aalStat' : 'slideStat';
  const label = ns === 'aal' ? 'aalStatLabel' : 'slideStatLabel';
  return [
    { type: stat, content: [{ type: 'text', text: '42' }] },
    { type: label, content: [{ type: 'text', text: 'Label describing the stat' }] },
  ];
}

function parseSlides(html) {
  const doc = new DOMParser().parseFromString(`<div>${html}</div>`, 'text/html');
  const sections = doc.querySelectorAll('section');
  if (sections.length === 0) {
    return [{ attrs: { class: 'slide pptx-slide' }, content: html }];
  }
  return Array.from(sections).map((sec) => {
    const attrs = {};
    for (const a of sec.attributes) attrs[a.name] = a.value;
    return { attrs, content: sec.innerHTML.trim() };
  });
}

function stripTags(s) {
  return (s || '').replace(/<[^>]+>/g, ' ').replace(/\s+/g, ' ').trim();
}
function escapeHTML(s) {
  return (s || '').replace(/[&<>"]/g, (c) => ({ '&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;' }[c]));
}
function escapeAttr(s) {
  return String(s).replace(/"/g, '&quot;');
}

// Globals for non-module callers (most blogs don't use ESM in admin pages).
if (typeof window !== 'undefined') {
  window.GoSlideEditor = GoSlideEditor;
}
