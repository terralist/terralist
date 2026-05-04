import { test, expect, type Locator, type Page } from '@playwright/test';

const gotoMarkdownFixture = async (page: Page): Promise<Locator> => {
  await page.goto('/#/markdown-test');

  await expect(page.getByText('Markdown rendering test')).toBeVisible({
    timeout: 10_000,
  });

  const markdown = page.locator('.markdown-body');
  await expect(markdown).toContainText('GitHub Flavored Markdown test');
  return markdown;
};

const cssNumber = async (
  locator: Locator,
  property: string
): Promise<number> => {
  const value = await locator.evaluate(
    (el, prop) => getComputedStyle(el).getPropertyValue(prop),
    property
  );

  return Number.parseFloat(value);
};

test.describe('GitHub Flavored Markdown rendering', () => {
  test('renders heading anchors, links, and duplicate slugs', async ({
    page,
  }) => {
    const markdown = await gotoMarkdownFixture(page);

    const h1 = markdown.locator('h1#h1-github-flavored-markdown-test');
    await expect(h1).toContainText('H1 — GitHub Flavored Markdown test');
    await expect(h1.locator('> a.heading-anchor')).toHaveAttribute(
      'href',
      '#/markdown-test?md-anchor=h1-github-flavored-markdown-test'
    );

    await expect(markdown.locator('h3#duplicate')).toBeVisible();
    await expect(markdown.locator('h3#duplicate-1')).toBeVisible();

    await expect(
      markdown.getByRole('link', { name: 'Tables', exact: true })
    ).toHaveAttribute('href', '#/markdown-test?md-anchor=tables');

    const external = markdown.locator(
      'a[href="https://github.com/terralist/terralist"]'
    );
    await expect(external).toHaveAttribute('target', '_blank');
    await expect(external).toHaveAttribute('rel', 'noopener noreferrer');
  });

  test('renders GFM blocks and GitHub extensions', async ({ page }) => {
    const markdown = await gotoMarkdownFixture(page);

    await expect(markdown.locator('table')).toBeVisible();
    await expect(markdown.locator('table th').first()).toContainText(
      'Column A'
    );
    await expect(markdown.locator('table td code')).toContainText('code');
    await expect(markdown.locator('table del')).toContainText('old');

    const taskInputs = markdown.locator('input[type="checkbox"]');
    await expect(taskInputs).toHaveCount(4);
    await expect(taskInputs.nth(0)).toBeChecked();
    await expect(taskInputs.nth(2)).not.toBeChecked();

    for (const variant of [
      'note',
      'tip',
      'important',
      'warning',
      'caution',
    ]) {
      await expect(
        markdown.locator(`.markdown-alert-${variant}`)
      ).toBeVisible();
    }

    await expect(markdown.locator('.footnotes')).toBeVisible();
    await expect(markdown.locator('[data-footnote-ref]').first()).toBeVisible();
    await expect(markdown.locator('.footnotes')).toContainText(
      'This is the first footnote.'
    );

    await expect(markdown.locator('details summary')).toContainText(
      'Click to expand'
    );
    await expect(markdown.locator('kbd').first()).toContainText('Cmd');
    await expect(markdown).toContainText('🚀 launching');

    await expect(markdown).toContainText("<script>alert('xss')</script>");
    await expect(markdown.locator('script')).toHaveCount(0);
  });

  test('keeps markdown-specific CSS overrides applied', async ({ page }) => {
    const markdown = await gotoMarkdownFixture(page);

    await expect(markdown).toHaveCSS('border-radius', '4px');
    expect(await cssNumber(markdown, 'padding-left')).toBeGreaterThan(0);

    await expect(markdown.locator('h2#unordered-lists + ul')).toHaveCSS(
      'list-style-type',
      'disc'
    );
    await expect(markdown.locator('h2#ordered-lists + ol')).toHaveCSS(
      'list-style-type',
      'decimal'
    );
    await expect(
      markdown.locator('h2#unordered-lists + ul ul').first()
    ).toHaveCSS('list-style-type', 'circle');

    const note = markdown.locator('.markdown-alert-note').first();
    await expect(note).toHaveCSS('border-left-style', 'solid');
    expect(await cssNumber(note, 'border-left-width')).toBeGreaterThanOrEqual(
      3
    );

    const heading = markdown.locator('h2#headings');
    const anchor = heading.locator('> a.heading-anchor');
    await expect(anchor).toHaveCSS('opacity', '0');
    await heading.hover();
    await expect(anchor).toHaveCSS('opacity', '0.6');
  });

  test('mounts syntax-highlighted code blocks and Mermaid diagrams', async ({
    page,
  }) => {
    const markdown = await gotoMarkdownFixture(page);

    await expect(markdown.locator('.md-code-block')).toHaveCount(4, {
      timeout: 10_000,
    });
    await expect(markdown.locator('.md-code-block .md-copy-btn')).toHaveCount(
      4
    );
    await expect(markdown.locator('.md-code-block .shiki').first()).toBeVisible(
      { timeout: 10_000 }
    );
    await expect(markdown.locator('.md-code-block').first()).toContainText(
      'echo "hello, terralist"'
    );

    await expect(markdown.locator('.md-mermaid svg')).toBeVisible({
      timeout: 10_000,
    });
  });
});
