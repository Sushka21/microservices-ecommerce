interface ImportMeta {
  readonly env: ImportMetaEnv
}

interface ImportMetaEnv {
  readonly VITE_API_URL?: string
  readonly VITE_CART_URL?: string
  readonly VITE_LOMS_URL?: string
}

declare module '*.css' {
  const src: string
  export default src
}
