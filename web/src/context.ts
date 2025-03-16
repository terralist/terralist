import { writable, type Writable } from 'svelte/store';

import { defaultIfNull } from './lib/utils';

type Theme = 'light' | 'dark';

function isTheme(arg: unknown): arg is Theme {
  return typeof arg == 'string' && ['light', 'darg'].includes(arg);
}
class Context {
  theme: Writable<Theme>;

  constructor() {
    const persistedTheme = defaultIfNull(
      localStorage.getItem('preferred.theme.style'),
      'light'
    );

    if (!isTheme(persistedTheme)) {
      throw new Error(`Unsupported theme: ${persistedTheme}`);
    }

    this.theme = writable(persistedTheme);
    this.setTheme(persistedTheme);
  }

  setTheme(theme: Theme): void {
    this.theme.set(theme);

    localStorage.setItem('preferred.theme.style', theme);

    const root = document.documentElement;

    if (theme === 'dark') {
      root.classList.add('dark');
    } else {
      root.classList.remove('dark');
    }
  }
}

const context: Context = new Context();

export default context;

export { type Theme };
