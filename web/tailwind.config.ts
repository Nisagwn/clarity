import type { Config } from 'tailwindcss'

// Marka paleti kullanıcının verdiği renk kartelasından türetildi.
// Endişe seviyesi renkleri (sage → clay → wine) bilinçli olarak paletin
// içinden seçildi: hem markayla uyumlu hem de sıcaklık arttıkça uyarı
// şiddeti artacak şekilde sıralı.
const config: Config = {
  content: [
    './app/**/*.{js,ts,jsx,tsx,mdx}',
    './components/**/*.{js,ts,jsx,tsx,mdx}',
  ],
  theme: {
    extend: {
      colors: {
        brand: {
          50: '#fdf8f8',  // ipeksi sütlü pudra (ikondan)
          100: '#f9ecec',
          200: '#e7babb', // açık pembe
          300: '#e98b96', // gül
          400: '#c16978', // gül kurusu
          500: '#933d4c', // şarap
          600: '#7a3140',
        },
        clay: '#d48d7b',    // terrakota
        cocoa: '#8f6344',   // kahve
        espresso: '#413014',// koyu kahve (ana metin)
        mauve: '#93777d',   // soluk leylak-gri
        mist: '#b2a3aa',    // pus
        sage: '#646c61',    // adaçayı
      },
      fontFamily: {
        display: ['var(--font-display)', 'Georgia', 'serif'],
      },
    },
  },
  plugins: [],
}
export default config
