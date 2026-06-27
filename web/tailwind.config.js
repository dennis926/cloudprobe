/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{vue,js,ts,jsx,tsx}"
  ],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        primary: '#38bdf8',
        secondary: '#a78bfa',
        bg: '#0f172a',
        surface: '#1e293b',
        ink: '#f1f5f9',
        muted: '#94a3b8',
        rule: '#334155'
      }
    }
  },
  plugins: []
}
