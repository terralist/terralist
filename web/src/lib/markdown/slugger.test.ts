import { describe, expect, it } from 'vitest';

import { Slugger } from './slugger';

describe('Slugger', () => {
  it('generates GitHub-style heading slugs', (): void => {
    const slugger = new Slugger();

    expect(slugger.slug('  GitHub Flavored Markdown!  ')).toBe(
      'github-flavored-markdown'
    );
    expect(slugger.slug('Multiple    spaces---collapse')).toBe(
      'multiple-spaces-collapse'
    );
  });

  it('deduplicates repeated slugs on the same instance', (): void => {
    const slugger = new Slugger();

    expect(slugger.slug('Duplicate')).toBe('duplicate');
    expect(slugger.slug('Duplicate')).toBe('duplicate-1');
    expect(slugger.slug('Duplicate')).toBe('duplicate-2');
  });

  it('can reset duplicate tracking', (): void => {
    const slugger = new Slugger();

    expect(slugger.slug('Duplicate')).toBe('duplicate');
    expect(slugger.slug('Duplicate')).toBe('duplicate-1');

    slugger.reset();

    expect(slugger.slug('Duplicate')).toBe('duplicate');
  });
});
