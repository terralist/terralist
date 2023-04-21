/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./src/**/*.svelte",
    "./*.html"
  ],
  theme: {
    extend: {
      transitionProperty: {
        width: "width",
      }
    }
  },
  plugins: [],
  darkMode: "class",
}
