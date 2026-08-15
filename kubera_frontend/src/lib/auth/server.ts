import { createNeonAuth } from "@neondatabase/auth/next/server";

const baseUrl = process.env.NEON_AUTH_BASE_URL;
const cookieSecret = process.env.NEON_AUTH_COOKIE_SECRET;

console.log("Auth URL exists:", Boolean(baseUrl));
console.log("Cookie secret exists:", Boolean(cookieSecret));

if (!baseUrl) {
  throw new Error("NEON_AUTH_BASE_URL is missing");
}

if (!cookieSecret) {
  throw new Error("NEON_AUTH_COOKIE_SECRET is missing");
}

export const auth = createNeonAuth({
  baseUrl,
  cookies: {
    secret: cookieSecret,
  },
});