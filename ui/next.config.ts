import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  output: 'export', // Required for Go binary embedding
  trailingSlash: true, // Better compatibility with static hosting
  images: {
    unoptimized: true,
  },
};

export default nextConfig;
