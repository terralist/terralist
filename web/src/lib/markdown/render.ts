import { Marked, type Token } from 'marked';
import markedAlert from 'marked-alert';
import markedFootnote from 'marked-footnote';

import { Slugger } from './slugger';

/**
 * Subset of the marked renderer `this` we rely on inside our overrides.
 * marked v15's renderer functions are bound to the parser via `this`; the
 * official typings infer this only when classes extend `Renderer`, so we
 * spell it out for our object-literal overrides.
 */
type RendererThis = { parser: { parseInline: (tokens: Token[]) => string } };

/**
 * SVG octicon GitHub uses for the heading anchor link icon.
 */
const ANCHOR_SVG =
  '<svg viewBox="0 0 16 16" width="16" height="16" aria-hidden="true" focusable="false">' +
  '<path fill="currentColor" d="m7.775 3.275 1.25-1.25a3.5 3.5 0 1 1 4.95 4.95l-2.5 2.5a3.5 3.5 0 0 1-4.95 0 .751.751 0 0 1 .018-1.042.751.751 0 0 1 1.042-.018 1.998 1.998 0 0 0 2.83 0l2.5-2.5a2.002 2.002 0 0 0-2.83-2.83l-1.25 1.25a.751.751 0 0 1-1.042-.018.751.751 0 0 1-.018-1.042Zm-4.69 9.64a1.998 1.998 0 0 0 2.83 0l1.25-1.25a.751.751 0 0 1 1.042.018.751.751 0 0 1 .018 1.042l-1.25 1.25a3.5 3.5 0 1 1-4.95-4.95l2.5-2.5a3.5 3.5 0 0 1 4.95 0 .751.751 0 0 1-.018 1.042.751.751 0 0 1-1.042.018 1.998 1.998 0 0 0-2.83 0l-2.5 2.5a1.998 1.998 0 0 0 0 2.83Z"/>' +
  '</svg>';

const escapeAttr = (value: string): string =>
  value
    .replace(/&/g, '&amp;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;');

const stripHtml = (html: string): string => html.replace(/<[^>]*>/g, '').trim();

/**
 * GFM "Disallowed Raw HTML" extension (spec section 6.11). Replaces the
 * leading `<` of any disallowed tag with `&lt;` so it renders as literal
 * text rather than executing. Operates on the final HTML, after marked has
 * already escaped content inside code blocks and inline code.
 */
const DISALLOWED_RAW_HTML_RE =
  /<(\/?)(title|textarea|style|xmp|iframe|noembed|noframes|script|plaintext)(?=[\s/>])/gi;

const stripDisallowedRawHtml = (html: string): string =>
  html.replace(DISALLOWED_RAW_HTML_RE, '&lt;$1$2');

/**
 * Allowlist of URL schemes considered safe for `<a href>`. Prevents
 * `javascript:` / `data:` / `vbscript:` URLs from sneaking through user
 * markdown.
 */
const SAFE_URL_RE =
  /^(?:https?:|mailto:|tel:|ftp:|sftp:|ssh:|git:|#|\/|\.\/|\.\.\/)/i;

const sanitizeHref = (href: string): string => {
  const trimmed = href.trim();
  if (!trimmed) return '';
  // Treat scheme-relative protocol as http/https, which is safe.
  if (trimmed.startsWith('//')) return trimmed;
  if (SAFE_URL_RE.test(trimmed)) return trimmed;
  // Allow bare relative paths like "foo/bar" or "foo.md".
  if (!/^[a-z][a-z0-9+.-]*:/i.test(trimmed)) return trimmed;
  return '';
};

/**
 * Encode arbitrary UTF-8 code into a safe base64 string for embedding inside
 * an HTML attribute. Plain `btoa` would throw on non-Latin-1 input.
 */
const encodeBase64Utf8 = (text: string): string => {
  const bytes = new TextEncoder().encode(text);
  let binary = '';
  for (const byte of bytes) {
    binary += String.fromCharCode(byte);
  }
  return btoa(binary);
};

/**
 * Inverse of {@link encodeBase64Utf8}, used by the Markdown component when it
 * mounts a `MarkdownCode` Svelte component into a code-block placeholder.
 */
export const decodeBase64Utf8 = (encoded: string): string => {
  const binary = atob(encoded);
  const bytes = new Uint8Array(binary.length);
  for (let i = 0; i < binary.length; i++) {
    bytes[i] = binary.charCodeAt(i);
  }
  return new TextDecoder().decode(bytes);
};

/**
 * CSS class used as a marker on the placeholder elements emitted by the code
 * renderer. The Markdown component finds them and mounts a `MarkdownCode`
 * Svelte component (which handles Shiki + Mermaid) into each one.
 */
export const CODE_PLACEHOLDER_CLASS = 'md-code-placeholder';

/**
 * Query key used with svelte-spa-router's hash URLs (`#/path?...`) so
 * in-document anchors are shareable. HTML only allows one `#fragment`; the
 * app route already occupies the hash (`#/markdown-test`), so heading and
 * footnote links are rewritten to `#/current/path?md-anchor=<id>`.
 */
export const MD_ANCHOR_QUERY = 'md-anchor';

export type RenderOptions = {
  /** SPA path from `location` (e.g. `/markdown-test`). When set, `href="#id"` links become shareable hash+query URLs. */
  routePath?: string | null;
};

const normalizeRoutePath = (routePath: string): string =>
  routePath.startsWith('/') ? routePath : `/${routePath}`;

/**
 * Rewrite `<a href="#fragment">` (not `#/…` routes) to
 * `#<routePath>?md-anchor=<encoded-fragment>` so links survive hash routing.
 */
const rewriteFragmentLinksForSpaHash = (
  html: string,
  routePath: string
): string => {
  const path = normalizeRoutePath(routePath.trim());
  const prefix = `#${path}?${MD_ANCHOR_QUERY}=`;
  return html.replace(/href="#(?!\/)([^"#]*)"/gi, (_m, fragment: string) =>
    fragment ? `href="${prefix}${encodeURIComponent(fragment)}"` : 'href="#"'
  );
};

/**
 * Render Markdown to HTML using the full GitHub Flavored Markdown feature
 * set:
 *
 *   - CommonMark base (paragraphs, emphasis, code, lists, links, images,
 *     blockquotes, raw HTML, hard breaks, etc.).
 *   - GFM extensions: tables, strikethrough, task list items, autolinks,
 *     disallowed raw HTML (handled by marked's default escaping).
 *   - GitHub additions on top of GFM that READMEs commonly use:
 *       - Heading anchor links (custom heading renderer + Slugger).
 *       - Alerts (`> [!NOTE]`, `> [!TIP]`, ...) via `marked-alert`.
 *       - Footnotes (`[^1]` / `[^1]: ...`) via `marked-footnote`.
 *       - External link safety (`target="_blank" rel="noopener noreferrer"`).
 *   - Code blocks are emitted as placeholder elements that the Svelte
 *     wrapper replaces with `MarkdownCode` instances so Shiki and Mermaid
 *     can run client-side.
 *
 * A fresh `Marked` instance and `Slugger` are created on every call so that
 * heading IDs and footnote numbering are scoped to a single document and do
 * not leak across renders.
 *
 * @param options.routePath When set (typically `$location` from svelte-spa-router),
 *   same-document anchors become `#/path?md-anchor=id` so shared URLs open the
 *   correct route and scroll to the target.
 */
export const render = (source: string, options?: RenderOptions): string => {
  const routePath = options?.routePath?.trim() ?? '';
  const slugger = new Slugger();
  const marked = new Marked({ gfm: true, breaks: false });

  marked.use(
    markedAlert(),
    markedFootnote({ refMarkers: true, footnoteDivider: true })
  );

  marked.use({
    hooks: {
      postprocess(html: string): string {
        let out = stripDisallowedRawHtml(html);
        if (routePath) {
          out = rewriteFragmentLinksForSpaHash(out, routePath);
        }
        return out;
      }
    },
    renderer: {
      heading(this: RendererThis, { tokens, depth }) {
        // Render the heading's inline content first so we can derive a slug
        // from the plain text and use the rendered HTML as the visible label.
        const innerHtml = this.parser.parseInline(tokens);
        const rawText = stripHtml(innerHtml);
        const id = slugger.slug(rawText);
        const aria = escapeAttr(`Permalink to ${rawText}`);
        const safeId = escapeAttr(id);
        return (
          `<h${depth} id="${safeId}" class="markdown-heading">` +
          `<a class="heading-anchor" href="#${safeId}" aria-label="${aria}">${ANCHOR_SVG}</a>` +
          innerHtml +
          `</h${depth}>\n`
        );
      },
      link(this: RendererThis, { href, title, tokens }) {
        const text = this.parser.parseInline(tokens);
        const safeHref = sanitizeHref(href);
        if (!safeHref) {
          // Drop the link entirely if the URL was unsafe; render plain text.
          return text;
        }
        const isExternal = /^(https?:)?\/\//i.test(safeHref);
        const titleAttr = title ? ` title="${escapeAttr(title)}"` : '';
        const externalAttrs = isExternal
          ? ' target="_blank" rel="noopener noreferrer"'
          : '';
        return `<a href="${escapeAttr(safeHref)}"${titleAttr}${externalAttrs}>${text}</a>`;
      },
      code({ text, lang }) {
        // Take just the first whitespace-delimited token; some sources write
        // ```ts copy``` etc. to attach extra metadata.
        const safeLang = (lang ?? '').split(/\s+/)[0] ?? '';
        const encoded = encodeBase64Utf8(text);
        return (
          `<div class="${CODE_PLACEHOLDER_CLASS}"` +
          ` data-md-lang="${escapeAttr(safeLang)}"` +
          ` data-md-code="${encoded}"></div>\n`
        );
      }
    }
  });

  return marked.parse(source, { async: false });
};
