// Conditionally load PWA if available
let withPWA = (config) => config
try {
  const pwa = require('next-pwa')
  withPWA = pwa({
    dest: 'public',
    register: true,
    skipWaiting: true,
    disable: process.env.NODE_ENV === 'development',
  })
} catch (e) {
  console.warn('next-pwa not found, skipping PWA configuration')
}

// Bundle analyzer (run with ANALYZE=true next build or npm run build:analyze)
let withBundleAnalyzer = (config) => config
if (process.env.ANALYZE === 'true') {
  try {
    const withBundleAnalyzerPkg = require('@next/bundle-analyzer')
    withBundleAnalyzer = withBundleAnalyzerPkg({ enabled: true })
  } catch (e) {
    console.warn('@next/bundle-analyzer not found')
  }
}

/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  // output: 'standalone', // Only use in production builds, not dev
  
  // Performance optimizations
  swcMinify: true,
  compress: true,
  
  // Enable standalone output for Docker
  output: process.env.NODE_ENV === 'production' ? 'standalone' : undefined,
  
  // Note: optimizeCss requires critters package which may not be available
  // experimental: {
  //   optimizeCss: true,
  // },
  
  // Image optimization
  images: {
    formats: ['image/avif', 'image/webp'],
    deviceSizes: [640, 750, 828, 1080, 1200, 1920, 2048, 3840],
    imageSizes: [16, 32, 48, 64, 96, 128, 256, 384],
  },
  
  // Experimental features for performance
  // Note: optimizeCss requires critters which may not be available
  // experimental: {
  //   optimizeCss: true,
  // },
  
  // Webpack optimizations (only in production)
  webpack: (config, { isServer, dev }) => {
    // Only apply custom chunk splitting in production
    if (!isServer && !dev) {
      config.optimization = {
        ...config.optimization,
        splitChunks: {
          chunks: 'all',
          maxInitialRequests: 25,
          minSize: 20000,
          cacheGroups: {
            default: false,
            vendors: false,
            // Vendor chunk for core libraries
            vendor: {
              name: 'vendor',
              chunks: 'all',
              test: /node_modules/,
              priority: 20,
            },
            // Chart libraries chunk (heavy)
            charts: {
              name: 'charts',
              test: /[\\/]node_modules[\\/](recharts|@nivo|react-plotly|d3)[\\/]/,
              chunks: 'all',
              priority: 25,
              reuseExistingChunk: true,
            },
            // Monaco editor chunk (very heavy)
            monaco: {
              name: 'monaco',
              test: /[\\/]node_modules[\\/]@monaco-editor[\\/]/,
              chunks: 'all',
              priority: 30,
              reuseExistingChunk: true,
            },
            // React and React DOM
            react: {
              name: 'react',
              test: /[\\/]node_modules[\\/](react|react-dom|scheduler)[\\/]/,
              chunks: 'all',
              priority: 15,
              reuseExistingChunk: true,
            },
            // Common chunk
            common: {
              name: 'common',
              minChunks: 2,
              chunks: 'all',
              priority: 10,
              reuseExistingChunk: true,
              enforce: true,
            },
          },
        },
      }
    }
    return config
  },
}

const applied = withPWA(nextConfig)
module.exports = withBundleAnalyzer(applied)
