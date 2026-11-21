declare module '*.css?inline' {
  const css: string;
  export default css;
}

// server/static/src/util.ts reads this global, which the webapp's index.html
// defines in an inline script (FINDINGS F4). Declared so the imported source
// type-checks; the guest never defines it and never reaches the code that
// reads it.
declare const API_HOST: string;
