import { hasLocale } from "next-intl";
import { getRequestConfig } from "next-intl/server";
import { routing } from "@/i18n/routing";

export default getRequestConfig(async ({ requestLocale }) => {
  const requested = await requestLocale;
  const locale = hasLocale(routing.locales, requested) ? requested : routing.defaultLocale;
  const [messages, defaults] = await Promise.all([
    import(`../../messages/${locale}.json`).then((module) => module.default),
    import("../../messages/en.json").then((module) => module.default),
  ]);

  return {
    locale,
    messages: { ...defaults, ...messages, Shell: defaults.Shell },
  };
});
