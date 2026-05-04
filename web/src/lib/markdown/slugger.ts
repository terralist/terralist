/**
 * Generates URL-safe slugs for headings, ensuring uniqueness across calls
 * on the same instance. Mirrors the algorithm GitHub uses for heading anchors,
 * so links match what users expect when copying anchors from rendered docs.
 */
export class Slugger {
  private readonly seen = new Map<string, number>();

  slug(value: string): string {
    const base = value
      .toLowerCase()
      .trim()
      .replace(/<[!/a-z].*?>/gi, '')
      .replace(
        // eslint-disable-next-line no-useless-escape
        /[\u2000-\u206F\u2E00-\u2E7F\\'!"#$%&()*+,./:;<=>?@\[\]^`{|}~]/g,
        ''
      )
      .replace(/\s+/g, '-')
      .replace(/-+/g, '-')
      .replace(/^-|-$/g, '');

    const count = this.seen.get(base) ?? 0;
    this.seen.set(base, count + 1);
    return count === 0 ? base : `${base}-${count}`;
  }

  reset(): void {
    this.seen.clear();
  }
}
