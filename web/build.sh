# Build script for O2OChat Web Client
# Compiles Go code to WebAssembly

set -e

echo "🔨 Building O2OChat Web Client..."

# Create output directory
mkdir -p web/dist

# Copy wasm_exec.js from Go installation
cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" web/dist/

# Compile Go to WebAssembly
echo "📦 Compiling Go to WebAssembly..."
GOOS=js GOARCH=wasm go build -o web/dist/main.wasm ./web/main.go

# Copy HTML
echo "📄 Copying HTML..."
cp web/index.html web/dist/index.html

# Create manifest.json for PWA
cat > web/dist/manifest.json << 'EOF'
{
  "name": "O2OChat Web",
  "short_name": "O2OChat",
  "description": "P2P Instant Messaging Web Client",
  "start_url": "/",
  "display": "standalone",
  "background_color": "#667eea",
  "theme_color": "#667eea",
  "icons": [
    {
      "src": "icon-192.png",
      "sizes": "192x192",
      "type": "image/png"
    },
    {
      "src": "icon-512.png",
      "sizes": "512x512",
      "type": "image/png"
    }
  ]
}
EOF

# Create service worker for PWA
cat > web/dist/sw.js << 'EOF'
const CACHE_NAME = 'o2ochat-v1';
const urlsToCache = [
  '/',
  '/index.html',
  '/wasm_exec.js',
  '/main.wasm',
  '/manifest.json'
];

self.addEventListener('install', event => {
  event.waitUntil(
    caches.open(CACHE_NAME)
      .then(cache => cache.addAll(urlsToCache))
  );
});

self.addEventListener('fetch', event => {
  event.respondWith(
    caches.match(event.request)
      .then(response => response || fetch(event.request))
  );
});

self.addEventListener('activate', event => {
  event.waitUntil(
    caches.keys().then(cacheNames => {
      return Promise.all(
        cacheNames.map(cacheName => {
          if (cacheName !== CACHE_NAME) {
            return caches.delete(cacheName);
          }
        })
      );
    })
  );
});
EOF

echo "✅ Build complete!"
echo ""
echo "📁 Output directory: web/dist/"
echo "🌐 Open web/dist/index.html in browser"
echo "📱 PWA support enabled"
echo ""
echo "🚀 To run locally:"
echo "   cd web/dist && python3 -m http.server 8080"
echo "   Open http://localhost:8080"
