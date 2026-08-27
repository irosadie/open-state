#!/bin/sh
set -e

# ============================================================
# Runtime env replacement for NEXT_PUBLIC_* variables.
#
# Next.js inlines NEXT_PUBLIC_* at BUILD time into the JS bundle.
# We build with unique placeholders (e.g. __NEXT_PUBLIC_API_URL__)
# and replace them here at container startup with real values
# from the .env passed via docker-compose.
# ============================================================

echo "🔧 Replacing build-time placeholders with runtime env values..."

# Find all JS files in the .next directory and replace placeholders
find /app/.next -type f -name "*.js" -exec sed -i \
  "s|__NEXT_PUBLIC_API_URL__|${NEXT_PUBLIC_API_URL}|g" {} +

echo "✅ Runtime env replacement complete."

# Start the Next.js server
exec node server.js
