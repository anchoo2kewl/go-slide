// Default theme schema — TipTap Node extensions matching the .slide-* classes
// in the Default theme CSS. Mirrors the structure of schema-aal.mjs so authors
// can switch themes without rewriting decks; only class names differ.

import { Node, mergeAttributes } from '@tiptap/core';

const passthroughAttr = (name, def = null) => ({
  default: def,
  parseHTML: (el) => el.getAttribute(name),
  renderHTML: (attrs) => (attrs[name] != null ? { [name]: attrs[name] } : {}),
});

export const SlideSection = Node.create({
  name: 'slideSection',
  group: 'block',
  content: 'block+',
  defining: true,
  parseHTML() { return [{ tag: 'section' }]; },
  renderHTML({ HTMLAttributes }) {
    return ['section', mergeAttributes(HTMLAttributes), 0];
  },
  addAttributes() {
    return {
      class: {
        default: 'slide pptx-slide',
        parseHTML: (el) => el.getAttribute('class') || 'slide pptx-slide',
        renderHTML: (a) => (a.class ? { class: a.class } : {}),
      },
      'data-background-color': passthroughAttr('data-background-color'),
      'data-background-image': passthroughAttr('data-background-image'),
      'data-transition': passthroughAttr('data-transition'),
    };
  },
});

export const SlideWrap = Node.create({
  name: 'slideWrap',
  group: 'block', content: 'block+', defining: true,
  parseHTML() { return [{ tag: 'div.slide-wrap' }, { tag: 'div.slide-wrap-top' }]; },
  renderHTML({ HTMLAttributes, node }) {
    return ['div', mergeAttributes(HTMLAttributes, { class: node.attrs.variant || 'slide-wrap' }), 0];
  },
  addAttributes() {
    return {
      variant: {
        default: 'slide-wrap',
        parseHTML: (el) => (el.classList.contains('slide-wrap-top') ? 'slide-wrap-top' : 'slide-wrap'),
        renderHTML: () => ({}),
      },
    };
  },
});

export const SlideBar = Node.create({
  name: 'slideBar', group: 'block', atom: true, selectable: true, draggable: true,
  parseHTML() { return [{ tag: 'div.slide-bar' }]; },
  renderHTML({ HTMLAttributes }) { return ['div', mergeAttributes(HTMLAttributes, { class: 'slide-bar' })]; },
});

export const SlideEyebrow = Node.create({
  name: 'slideEyebrow', group: 'block', content: 'inline*',
  parseHTML() { return [{ tag: 'span.slide-eyebrow' }]; },
  renderHTML({ HTMLAttributes }) { return ['span', mergeAttributes(HTMLAttributes, { class: 'slide-eyebrow' }), 0]; },
});

export const SlideRule = Node.create({
  name: 'slideRule', group: 'block', atom: true, selectable: true,
  parseHTML() { return [{ tag: 'hr.slide-rule' }]; },
  renderHTML({ HTMLAttributes }) { return ['hr', mergeAttributes(HTMLAttributes, { class: 'slide-rule' })]; },
});

export const SlideH1 = Node.create({
  name: 'slideH1', group: 'block', content: 'inline*', defining: true,
  parseHTML() { return [{ tag: 'h1.slide-h1' }]; },
  renderHTML({ HTMLAttributes }) { return ['h1', mergeAttributes(HTMLAttributes, { class: 'slide-h1' }), 0]; },
});
export const SlideH2 = Node.create({
  name: 'slideH2', group: 'block', content: 'inline*', defining: true,
  parseHTML() { return [{ tag: 'h2.slide-h2' }]; },
  renderHTML({ HTMLAttributes }) { return ['h2', mergeAttributes(HTMLAttributes, { class: 'slide-h2' }), 0]; },
});
export const SlideLede = Node.create({
  name: 'slideLede', group: 'block', content: 'inline*', defining: true,
  parseHTML() { return [{ tag: 'p.slide-lede' }]; },
  renderHTML({ HTMLAttributes }) { return ['p', mergeAttributes(HTMLAttributes, { class: 'slide-lede' }), 0]; },
});

export const SlidePill = Node.create({
  name: 'slidePill', group: 'inline', inline: true, content: 'text*',
  parseHTML() { return [{ tag: 'span.slide-pill' }]; },
  renderHTML({ HTMLAttributes }) { return ['span', mergeAttributes(HTMLAttributes, { class: 'slide-pill' }), 0]; },
});

export const SlideGrid = Node.create({
  name: 'slideGrid', group: 'block', content: 'slideCard+', defining: true,
  parseHTML() {
    return [
      { tag: 'div.slide-grid-2', attrs: { cols: 2 } },
      { tag: 'div.slide-grid-3', attrs: { cols: 3 } },
      { tag: 'div.slide-grid-4', attrs: { cols: 4 } },
    ];
  },
  renderHTML({ HTMLAttributes, node }) {
    const cols = node.attrs.cols || 3;
    return ['div', mergeAttributes(HTMLAttributes, { class: `slide-grid-${cols}` }), 0];
  },
  addAttributes() {
    return {
      cols: {
        default: 3,
        parseHTML: (el) => {
          if (el.classList.contains('slide-grid-2')) return 2;
          if (el.classList.contains('slide-grid-4')) return 4;
          return 3;
        },
        renderHTML: () => ({}),
      },
    };
  },
});

export const SlideCard = Node.create({
  name: 'slideCard', group: 'block', content: 'block+', defining: true,
  parseHTML() { return [{ tag: 'div.slide-card' }]; },
  renderHTML({ HTMLAttributes }) { return ['div', mergeAttributes(HTMLAttributes, { class: 'slide-card' }), 0]; },
});

export const SlideStat = Node.create({
  name: 'slideStat', group: 'block', content: 'text*',
  parseHTML() { return [{ tag: 'div.slide-stat' }]; },
  renderHTML({ HTMLAttributes }) { return ['div', mergeAttributes(HTMLAttributes, { class: 'slide-stat' }), 0]; },
});
export const SlideStatLabel = Node.create({
  name: 'slideStatLabel', group: 'block', content: 'inline*',
  parseHTML() { return [{ tag: 'div.slide-stat-label' }]; },
  renderHTML({ HTMLAttributes }) { return ['div', mergeAttributes(HTMLAttributes, { class: 'slide-stat-label' }), 0]; },
});

export const SlideQuote = Node.create({
  name: 'slideQuote', group: 'block', content: 'inline*', defining: true,
  parseHTML() { return [{ tag: 'p.slide-quote' }]; },
  renderHTML({ HTMLAttributes }) { return ['p', mergeAttributes(HTMLAttributes, { class: 'slide-quote' }), 0]; },
});

export const SlideFoot = Node.create({
  name: 'slideFoot', group: 'block', content: 'inline*', defining: true,
  parseHTML() { return [{ tag: 'div.slide-foot' }]; },
  renderHTML({ HTMLAttributes }) { return ['div', mergeAttributes(HTMLAttributes, { class: 'slide-foot' }), 0]; },
});

export const SlideMono = Node.create({
  name: 'slideMono', group: 'inline', inline: true, content: 'text*',
  parseHTML() { return [{ tag: 'span.slide-mono' }]; },
  renderHTML({ HTMLAttributes }) { return ['span', mergeAttributes(HTMLAttributes, { class: 'slide-mono' }), 0]; },
});

export const DefaultSchema = [
  SlideSection, SlideWrap, SlideBar,
  SlideEyebrow, SlideRule,
  SlideH1, SlideH2, SlideLede,
  SlidePill,
  SlideGrid, SlideCard,
  SlideStat, SlideStatLabel,
  SlideQuote, SlideFoot, SlideMono,
];
