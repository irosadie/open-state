import { dirname, join } from "node:path"
import { fileURLToPath } from "node:url"
import type { NextConfig } from "next"

const currentDirectory = dirname(fileURLToPath(import.meta.url))

const nextConfig: NextConfig = {
  output: "standalone",
  outputFileTracingRoot: join(currentDirectory, "../../"),
  images: {
    remotePatterns: [
      { protocol: "https", hostname: "lh3.googleusercontent.com" },
      { protocol: "https", hostname: "avatars.githubusercontent.com" },
      { protocol: "https", hostname: "kinchay.b-cdn.net" },
    ],
  },
  turbopack: {
    resolveAlias: {
      $: ".",
    },
  },
}

export default nextConfig
