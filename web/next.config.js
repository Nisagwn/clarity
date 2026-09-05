/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  images: {
    remotePatterns: [
      { protocol: 'http', hostname: 'localhost' },
      { protocol: 'https', hostname: 'images.unsplash.com' },
      // Open Beauty Facts ürün görselleri (CC-BY-SA; atıf ürün sayfasında)
      { protocol: 'https', hostname: 'images.openbeautyfacts.org' },
      { protocol: 'https', hostname: 'images.openfoodfacts.org' },
    ],
  },
}

module.exports = nextConfig
