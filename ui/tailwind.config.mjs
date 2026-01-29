/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{ts,tsx}'],
  theme: {
    extend: {
      colors: {
        nintendo: '#e60012',
        sony: '#003087',
      },
    },
  },
  plugins: [],
}
