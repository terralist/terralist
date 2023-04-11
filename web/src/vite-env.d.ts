/// <reference types="svelte" />
/// <reference types="vite/client" />

interface ImportMetaBuildEnv {
  readonly TERRALIST_VERSION: string
}

interface ImportMeta {
  readonly build: ImportMetaBuildEnv
}