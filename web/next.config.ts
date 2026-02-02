import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  output: "standalone",
  generateBuildId: async () => {
    return process.env.GIT_HASH || "build-id-not-set";
  },
};

export default nextConfig;
