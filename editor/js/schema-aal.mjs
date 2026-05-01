// AAL theme schema — TipTap Node extensions for every primitive in the
// AI Agent Lens design system. Each extension defines parseHTML and
// renderHTML so HTML loaded into the editor round-trips losslessly.
//
// The pattern is:
//   - Block nodes inherit from Node, declare their group/content rules,
//     and parse/render their wrapping <div>.
//   - Inline nodes are TipTap "atoms" or marks, parsed from a span/class.
//   - Variant attributes (e.g. aal-card vs aal-card-light) flow through
//     parseHTML → state → renderHTML so the editor can switch variants
//     and the source HTML reflects it.
//
// Every node MUST round-trip: load any HTML matching parseHTML, then
// renderHTML must produce HTML that parseHTML accepts back. Tests in
// _examples/basic verify this by re-loading rendered output.

import { Node, mergeAttributes } from '@tiptap/core';

// Helper: a class attribute that pulls from the element's actual className,
// keeping any extra classes the author added. Used by AalSection and AalWrap.
const classAttr = (defaultClass) => ({
  default: defaultClass,
  parseHTML: (el) => el.getAttribute('class') || defaultClass,
  renderHTML: (attrs) => (attrs.class ? { class: attrs.class } : {}),
});

// Helper: a plain HTML attribute that survives load/save unchanged.
const passthroughAttr = (name, def = null) => ({
  default: def,
  parseHTML: (el) => el.getAttribute(name),
  renderHTML: (attrs) => (attrs[name] != null ? { [name]: attrs[name] } : {}),
});

// AalSection — top-level <section class="aal pptx-slide" ...>.
// In go-slide each editor instance edits the inside of one section, so
// this node is rarely nested in the document; it's used by the slide
// loader to capture the section's bg color, theme variant, and class
// list, then re-emit them on save.
export const AalSection = Node.create({
  name: 'aalSection',
  group: 'block',
  content: 'block+',
  defining: true,
  parseHTML() { return [{ tag: 'section' }]; },
  renderHTML({ HTMLAttributes }) {
    return ['section', mergeAttributes(HTMLAttributes), 0];
  },
  addAttributes() {
    return {
      class: classAttr('aal pptx-slide'),
      'data-background-color': passthroughAttr('data-background-color'),
      'data-background-image': passthroughAttr('data-background-image'),
      'data-transition': passthroughAttr('data-transition'),
    };
  },
});

// AalWrap — the flex column that contains everything visible.
// Renders as <div class="aal-wrap"> (or .aal-wrap-top variant).
export const AalWrap = Node.create({
  name: 'aalWrap',
  group: 'block',
  content: 'block+',
  defining: true,
  parseHTML() {
    return [
      { tag: 'div.aal-wrap' },
      { tag: 'div.aal-wrap-top' },
    ];
  },
  renderHTML({ HTMLAttributes, node }) {
    const cls = node.attrs.variant || 'aal-wrap';
    return ['div', mergeAttributes(HTMLAttributes, { class: cls }), 0];
  },
  addAttributes() {
    return {
      variant: {
        default: 'aal-wrap',
        parseHTML: (el) =>
          el.classList.contains('aal-wrap-top') ? 'aal-wrap-top' : 'aal-wrap',
        renderHTML: () => ({}),
      },
    };
  },
});

// AalBar — the vertical accent bar at the slide's left edge.
// Atom node (no children, no editable text).
export const AalBar = Node.create({
  name: 'aalBar',
  group: 'block',
  atom: true,
  selectable: true,
  draggable: true,
  parseHTML() { return [{ tag: 'div.aal-bar' }]; },
  renderHTML({ HTMLAttributes }) {
    return ['div', mergeAttributes(HTMLAttributes, { class: 'aal-bar' })];
  },
});

// AalEyebrow — small caps teal label at the top of a slide.
// Inline-block; contains plain text.
export const AalEyebrow = Node.create({
  name: 'aalEyebrow',
  group: 'block',
  content: 'inline*',
  parseHTML() {
    return [
      { tag: 'span.aal-eyebrow' },
      { tag: 'span.aal-eyebrow-light' },
    ];
  },
  renderHTML({ HTMLAttributes, node }) {
    const cls = node.attrs.variant || 'aal-eyebrow';
    return ['span', mergeAttributes(HTMLAttributes, { class: cls }), 0];
  },
  addAttributes() {
    return {
      variant: {
        default: 'aal-eyebrow',
        parseHTML: (el) =>
          el.classList.contains('aal-eyebrow-light')
            ? 'aal-eyebrow-light' : 'aal-eyebrow',
        renderHTML: () => ({}),
      },
    };
  },
});

// AalRule — accent rule under headings.
export const AalRule = Node.create({
  name: 'aalRule',
  group: 'block',
  atom: true,
  selectable: true,
  parseHTML() { return [{ tag: 'hr.aal-rule' }]; },
  renderHTML({ HTMLAttributes }) {
    return ['hr', mergeAttributes(HTMLAttributes, { class: 'aal-rule' })];
  },
});

// AalH1 / AalH2 — hero typography. We use custom node names rather than
// extending the StarterKit Heading because the visual semantics differ
// (size, weight, color come from .aal-h1/.aal-h2 classes).
export const AalH1 = Node.create({
  name: 'aalH1',
  group: 'block',
  content: 'inline*',
  defining: true,
  parseHTML() { return [{ tag: 'h1.aal-h1' }]; },
  renderHTML({ HTMLAttributes }) {
    return ['h1', mergeAttributes(HTMLAttributes, { class: 'aal-h1' }), 0];
  },
});

export const AalH2 = Node.create({
  name: 'aalH2',
  group: 'block',
  content: 'inline*',
  defining: true,
  parseHTML() { return [{ tag: 'h2.aal-h2' }]; },
  renderHTML({ HTMLAttributes }) {
    return ['h2', mergeAttributes(HTMLAttributes, { class: 'aal-h2' }), 0];
  },
});

// AalLede — large lede paragraph following a hero heading.
export const AalLede = Node.create({
  name: 'aalLede',
  group: 'block',
  content: 'inline*',
  defining: true,
  parseHTML() {
    return [
      { tag: 'p.aal-lede' },
      { tag: 'p.aal-lede-light' },
    ];
  },
  renderHTML({ HTMLAttributes, node }) {
    const cls = node.attrs.variant || 'aal-lede';
    return ['p', mergeAttributes(HTMLAttributes, { class: cls }), 0];
  },
  addAttributes() {
    return {
      variant: {
        default: 'aal-lede',
        parseHTML: (el) =>
          el.classList.contains('aal-lede-light') ? 'aal-lede-light' : 'aal-lede',
        renderHTML: () => ({}),
      },
    };
  },
});

// AalPill — rounded outlined pill (inline-block).
export const AalPill = Node.create({
  name: 'aalPill',
  group: 'inline',
  inline: true,
  content: 'text*',
  parseHTML() { return [{ tag: 'span.aal-pill' }]; },
  renderHTML({ HTMLAttributes }) {
    return ['span', mergeAttributes(HTMLAttributes, { class: 'aal-pill' }), 0];
  },
});

// AalGrid — the layout primitive that holds Cards in a row.
// One node with a "cols" attribute (2 / 3 / 4) keeps the schema small.
export const AalGrid = Node.create({
  name: 'aalGrid',
  group: 'block',
  content: 'aalCard+',
  defining: true,
  parseHTML() {
    return [
      { tag: 'div.aal-grid-2', attrs: { cols: 2 } },
      { tag: 'div.aal-grid-3', attrs: { cols: 3 } },
      { tag: 'div.aal-grid-4', attrs: { cols: 4 } },
    ];
  },
  renderHTML({ HTMLAttributes, node }) {
    const cols = node.attrs.cols || 3;
    return ['div', mergeAttributes(HTMLAttributes, { class: `aal-grid-${cols}` }), 0];
  },
  addAttributes() {
    return {
      cols: {
        default: 3,
        parseHTML: (el) => {
          if (el.classList.contains('aal-grid-2')) return 2;
          if (el.classList.contains('aal-grid-4')) return 4;
          return 3;
        },
        renderHTML: () => ({}),
      },
    };
  },
});

// AalCard — content card. Two visual variants (dark/light) drawn from
// theme tokens; structure is the same.
export const AalCard = Node.create({
  name: 'aalCard',
  group: 'block',
  content: 'block+',
  defining: true,
  parseHTML() {
    return [
      { tag: 'div.aal-card' },
      { tag: 'div.aal-card-light' },
    ];
  },
  renderHTML({ HTMLAttributes, node }) {
    const cls = node.attrs.variant || 'aal-card';
    return ['div', mergeAttributes(HTMLAttributes, { class: cls }), 0];
  },
  addAttributes() {
    return {
      variant: {
        default: 'aal-card',
        parseHTML: (el) =>
          el.classList.contains('aal-card-light') ? 'aal-card-light' : 'aal-card',
        renderHTML: () => ({}),
      },
    };
  },
});

// AalStat — big number display. Container holds the stat value as text,
// followed by an AalStatLabel sibling for the caption.
export const AalStat = Node.create({
  name: 'aalStat',
  group: 'block',
  content: 'text*',
  parseHTML() {
    return [
      { tag: 'div.aal-stat' },
      { tag: 'div.aal-stat-light' },
    ];
  },
  renderHTML({ HTMLAttributes, node }) {
    const cls = node.attrs.variant || 'aal-stat';
    return ['div', mergeAttributes(HTMLAttributes, { class: cls }), 0];
  },
  addAttributes() {
    return {
      variant: {
        default: 'aal-stat',
        parseHTML: (el) =>
          el.classList.contains('aal-stat-light') ? 'aal-stat-light' : 'aal-stat',
        renderHTML: () => ({}),
      },
    };
  },
});

export const AalStatLabel = Node.create({
  name: 'aalStatLabel',
  group: 'block',
  content: 'inline*',
  parseHTML() { return [{ tag: 'div.aal-stat-label' }]; },
  renderHTML({ HTMLAttributes }) {
    return ['div', mergeAttributes(HTMLAttributes, { class: 'aal-stat-label' }), 0];
  },
});

// AalQuote — left-bordered pull-quote.
export const AalQuote = Node.create({
  name: 'aalQuote',
  group: 'block',
  content: 'inline*',
  defining: true,
  parseHTML() { return [{ tag: 'p.aal-quote' }]; },
  renderHTML({ HTMLAttributes }) {
    return ['p', mergeAttributes(HTMLAttributes, { class: 'aal-quote' }), 0];
  },
});

// AalFoot — footer row at the bottom of every slide.
export const AalFoot = Node.create({
  name: 'aalFoot',
  group: 'block',
  content: 'inline*',
  defining: true,
  parseHTML() {
    return [
      { tag: 'div.aal-foot' },
      { tag: 'div.aal-foot-light' },
    ];
  },
  renderHTML({ HTMLAttributes, node }) {
    const cls = node.attrs.variant || 'aal-foot';
    return ['div', mergeAttributes(HTMLAttributes, { class: cls }), 0];
  },
  addAttributes() {
    return {
      variant: {
        default: 'aal-foot',
        parseHTML: (el) =>
          el.classList.contains('aal-foot-light') ? 'aal-foot-light' : 'aal-foot',
        renderHTML: () => ({}),
      },
    };
  },
});

// AalMono — inline mono span (for command names, tokens).
export const AalMono = Node.create({
  name: 'aalMono',
  group: 'inline',
  inline: true,
  content: 'text*',
  parseHTML() { return [{ tag: 'span.aal-mono' }]; },
  renderHTML({ HTMLAttributes }) {
    return ['span', mergeAttributes(HTMLAttributes, { class: 'aal-mono' }), 0];
  },
});

// All AAL nodes, in the order they should be registered with the editor.
export const AalSchema = [
  AalSection, AalWrap, AalBar,
  AalEyebrow, AalRule,
  AalH1, AalH2, AalLede,
  AalPill,
  AalGrid, AalCard,
  AalStat, AalStatLabel,
  AalQuote, AalFoot, AalMono,
];
