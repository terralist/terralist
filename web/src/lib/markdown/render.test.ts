import { describe, expect, it } from 'vitest';

import { CODE_PLACEHOLDER_CLASS, decodeBase64Utf8, render } from './render';

describe('render', () => {
  it('renders heading anchors with duplicate GitHub-style ids', (): void => {
    const html = render('# Title\n\n## Title', { routePath: '/markdown-test' });

    expect(html).toContain('<h1 id="title" class="markdown-heading">');
    expect(html).toContain(
      '<a class="heading-anchor" href="#/markdown-test?md-anchor=title"'
    );
    expect(html).toContain('<h2 id="title-1" class="markdown-heading">');
    expect(html).toContain(
      '<a class="heading-anchor" href="#/markdown-test?md-anchor=title-1"'
    );
  });

  it('renders common GFM blocks and inline extensions', (): void => {
    const html = render(`
| Feature | Status |
| ------- | ------ |
| Tables | yes |

- [x] checked
- [ ] unchecked

~~removed~~
https://example.com
`);

    expect(html).toContain('<table>');
    expect(html).toContain('<th>Feature</th>');
    expect(html).toMatch(/<input[^>]*checked[^>]*>/);
    expect(html).toContain('type="checkbox"');
    expect(html).toContain('<del>removed</del>');
    expect(html).toContain('<a href="https://example.com"');
  });

  it('renders GitHub alert and footnote extensions', (): void => {
    const html = render(`
> [!NOTE]
> Useful context.

Footnote reference[^one].

[^one]: Footnote body.
`);

    expect(html).toContain('markdown-alert markdown-alert-note');
    expect(html).toContain('markdown-alert-title');
    expect(html).toContain('data-footnote-ref');
    expect(html).toContain('class="footnotes"');
    expect(html).toContain('Footnote body.');
  });

  it('sanitizes unsafe links and annotates external links', (): void => {
    const html = render(
      '[bad](javascript:alert(1)) [ok](https://example.com "Example")'
    );

    expect(html).not.toContain('javascript:');
    expect(html).toContain('bad');
    expect(html).toContain(
      '<a href="https://example.com" title="Example" target="_blank" rel="noopener noreferrer">ok</a>'
    );
  });

  it('escapes GFM disallowed raw HTML tags', (): void => {
    const html = render("<script>alert('xss')</script>");

    expect(html).toContain("&lt;script>alert('xss')&lt;/script>");
    expect(html).not.toContain("<script>alert('xss')</script>");
  });

  it('emits UTF-8-safe code placeholders for Svelte mounting', (): void => {
    const html = render('```ts copy\nconsole.log("sparkles ✨");\n```');
    const codeAttr = html.match(/data-md-code="([^"]+)"/)?.[1];

    expect(html).toContain(`class="${CODE_PLACEHOLDER_CLASS}"`);
    expect(html).toContain('data-md-lang="ts"');
    expect(codeAttr).toBeDefined();
    expect(decodeBase64Utf8(codeAttr ?? '')).toBe(
      'console.log("sparkles ✨");'
    );
  });

  it('rewrites in-document fragments for SPA hash routing', (): void => {
    const html = render('[Tables](#tables) [Route](#/settings)', {
      routePath: 'markdown-test'
    });

    expect(html).toContain('href="#/markdown-test?md-anchor=tables"');
    expect(html).toContain('href="#/settings"');
  });
});
