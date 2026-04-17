/**
 * Reasonably exhaustive GitHub Flavored Markdown fixture. Rendered at
 * `/markdown-test`, this exercises every feature the Terralist markdown
 * pipeline is expected to support so visual regressions are obvious.
 *
 * Sections are ordered to roughly mirror the GFM spec
 * (https://github.github.com/gfm/) followed by the extensions GitHub layers
 * on top (alerts, footnotes, emoji, mermaid).
 */
export const markdownSample = `# H1 — GitHub Flavored Markdown test

This page exercises the markdown renderer end-to-end. Compare it against the
same source rendered on github.com to spot regressions.

## Inline formatting

Plain text with **bold**, *italic*, ***bold italic***, ~~strikethrough~~,
\`inline code\`, sub<sub>script</sub>, sup<sup>erscript</sup>, and a hard
line break:
line one\\
line two.

Escapes work too: \\*not italic\\* and \\\`not code\\\`.

---

## Headings

Each heading should show a link icon when hovered. Duplicate names should get
unique anchors (e.g. \`#duplicate\`, \`#duplicate-1\`).

### Duplicate
### Duplicate
#### H4 example
##### H5 example
###### H6 example

## Links and autolinks

- Markdown link: [Terralist](https://github.com/terralist/terralist)
- Title attribute: [hover me](https://terraform.io "Go to terraform.io")
- Internal anchor: [Tables](#tables)
- Relative path: [Dashboard](/)
- Bare URL autolink: https://github.com
- Angle-bracket autolink: <https://www.hashicorp.com>
- Email autolink: <support@example.com>

External links should open in a new tab; in-doc links should not.

## Images

![Terralist logo](/terralist.svg "Terralist logo")

## Unordered lists

- Item one
- Item two
  - Nested item A
  - Nested item B
    - Deep nested item
- Item three

## Ordered lists

1. First step
2. Second step
   1. Sub-step
   2. Another sub-step
3. Third step

## Task lists

- [x] Completed task with **bold** and \`code\`
- [x] Another completed task
- [ ] Outstanding task
- [ ] Task with a [link](https://example.com)

## Tables

| Column A | Column B  | Column C |
| :------- | :-------: | -------: |
| left     | center    |    right |
| \`code\`   | **bold**  | *italic* |
| 1        | 2         |        3 |
| ~~old~~  | new       | https://example.com |

## Blockquotes

> A simple blockquote.
>
> With multiple paragraphs and **inline formatting**.
>
> > And a nested blockquote.

## GFM alerts

> [!NOTE]
> Highlights information that users should take into account, even when
> skimming.

> [!TIP]
> Optional information to help a user be more successful.

> [!IMPORTANT]
> Crucial information necessary for users to succeed.

> [!WARNING]
> Critical content demanding immediate user attention due to potential risks.

> [!CAUTION]
> Negative potential consequences of an action.

## Code blocks

Inline: \`terraform init\`.

### Bash code block
\`\`\`bash
# A bash block
echo "hello, terralist"
for i in 1 2 3; do
  echo "iteration $i"
done
\`\`\`

### HCL code block
\`\`\`hcl
resource "aws_s3_bucket" "example" {
  bucket = "my-tf-test-bucket"

  tags = {
    Name        = "My bucket"
    Environment = "Dev"
  }
}
\`\`\`

### JSON code block
\`\`\`json
{
  "name": "terralist",
  "version": "1.0.0",
  "features": ["modules", "providers"]
}
\`\`\`

### Mermaid code block
\`\`\`mermaid
flowchart LR
  A[Client] --> B{Auth?}
  B -- yes --> C[Terralist API]
  B -- no  --> D[401]
  C --> E[(Storage)]
\`\`\`

A code block without a language should still render as a plain block:

\`\`\`
no language tag here
\`\`\`

## Horizontal rule

---

## Footnotes

Here is a simple footnote[^1]. With some additional text after it[^longer]
and without disrupting the surrounding blocks.

[^1]: This is the first footnote.
[^longer]: A longer footnote with **inline formatting**, a [link](https://github.com),
    and a second paragraph.

    > Even a blockquote inside the footnote.

## HTML pass-through

<details>
<summary>Click to expand</summary>

Hidden content with **markdown** still rendered inside.

</details>

<kbd>Cmd</kbd> + <kbd>K</kbd> chord rendered with raw HTML.

## Emoji shortcodes

:rocket: launching, :sparkles: looking shiny, :white_check_mark: validated.

## Disallowed raw HTML

GFM strips dangerous tags such as \`<script>\` and \`<iframe>\`. The
following should render as escaped text rather than executing:

<script>alert('xss')</script>
`;
