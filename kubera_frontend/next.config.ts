import type { NextConfig } from "next";
import createNextIntlPlugin from "next-intl/plugin";

const withNextIntl = createNextIntlPlugin("./src/i18n/request.ts");

const nextConfig: NextConfig = {
  // Keep Webpack development artifacts separate from production/Turbopack
  // output. Mixing both in .next can leave the browser subscribed to a stale
  // HMR runtime after a build or bundler switch.
  distDir: process.env.NODE_ENV === "development" ? ".next-dev" : ".next",
};

export default withNextIntl(nextConfig);
