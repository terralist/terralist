/// <reference types="svelte" />
/// <reference types="vite/client" />

interface ImportMetaBuildEnv {
  readonly VITE_TERRALIST_VERSION: string
}

interface ImportMeta {
  readonly build: ImportMetaBuildEnv
}