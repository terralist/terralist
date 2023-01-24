import { writable, type Writable } from "svelte/store";

import { defaultIfNull } from "./lib/utils";

type Theme = "light" | "dark";

class Context {
  theme: Writable<Theme>;

  constructor() {
    let persistedTheme = defaultIfNull(localStorage.getItem("preferred.theme.style"), "light");
    this.theme = writable(persistedTheme);
    this.setTheme(persistedTheme);
  }

  setTheme(theme: Theme) {
    this.theme.set(theme);

    localStorage.setItem("preferred.theme.style", theme);
    
    const root = document.documentElement;

    if (theme === "dark") {
      root.classList.add("dark");
    } else {
      root.classList.remove("dark");
    }
  }
};

const context: Context = new Context();

export default context;

export {
  type Theme
};