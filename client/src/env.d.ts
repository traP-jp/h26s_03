/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_API_BASE?: string;
  readonly VITE_API_MOCKING?: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}

declare module "*.vue";
