/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{vue,js,ts,jsx,tsx}",
  ],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        bg0: 'var(--bg0)',
        bg1: 'var(--bg1)',
        bg2: 'var(--bg2)',
        bg3: 'var(--bg3)',
        bg4: 'var(--bg4)',
        glass: 'var(--glass)',
        bd: 'var(--bd)',
        bd2: 'var(--bd2)',
        bd3: 'var(--bd3)',
        text0: 'var(--text0)',
        text1: 'var(--text1)',
        text2: 'var(--text2)',
        text3: 'var(--text3)',
        acc: 'var(--acc)',
        acc2: 'var(--acc2)',
        'acc-bg': 'var(--acc-bg)',
        'acc-bg2': 'var(--acc-bg2)',
        red: 'var(--red)',
        'red-bg': 'var(--red-bg)',
        'red-glow': 'var(--red-glow)',
        ora: 'var(--ora)',
        'ora-bg': 'var(--ora-bg)',
        yel: 'var(--yel)',
        'yel-bg': 'var(--yel-bg)',
        grn: 'var(--grn)',
        'grn-bg': 'var(--grn-bg)',
        blu: 'var(--blu)',
        'blu-bg': 'var(--blu-bg)',
      },
      fontFamily: {
        sans: ['DM Sans', 'sans-serif'],
        mono: ['Geist Mono', 'monospace'],
      },
      borderRadius: {
        'sm': 'var(--r-sm)',
        'md': 'var(--r-md)',
        'lg': 'var(--r-lg)',
        'xl': 'var(--r-xl)',
      },
      animation: {
        'pulse-dot': 'pulse-dot 2.5s ease-in-out infinite',
        'slideUp': 'slideUp 0.3s both',
        'blink': 'blink 0.85s step-end infinite',
      },
      keyframes: {
        'pulse-dot': {
          '0%, 100%': { opacity: 1, boxShadow: '0 0 6px var(--grn)' },
          '50%': { opacity: 0.5, boxShadow: 'none' },
        },
        'slideUp': {
          'from': { opacity: 0, transform: 'translateY(8px)' },
          'to': { opacity: 1, transform: 'translateY(0)' },
        },
        'blink': {
          '50%': { opacity: 0 },
        }
      },
    },
  },
  plugins: [],
}
