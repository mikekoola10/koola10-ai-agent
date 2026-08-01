import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  // TypeScript still gates the build; ESLint 9 flag incompatibilities with
  // the Next 15 lint runner shouldn't block production deploys.
  eslint: {
    ignoreDuringBuilds: true,
  },
};

export default nextConfig;
