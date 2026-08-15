import { hasLocale } from "next-intl";
import { getRequestConfig } from "next-intl/server";
import { routing } from "@/i18n/routing";

function isMessageNamespace(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

function mergeMessages(defaults: Record<string, unknown>, localized: Record<string, unknown>) {
  const namespaces = new Set([...Object.keys(defaults), ...Object.keys(localized)]);
  return Object.fromEntries([...namespaces].map((namespace) => {
    const fallback = defaults[namespace];
    const translation = localized[namespace];
    if (isMessageNamespace(fallback) && isMessageNamespace(translation)) {
      return [namespace, { ...fallback, ...translation }];
    }
    return [namespace, translation ?? fallback];
  }));
}

export default getRequestConfig(async ({ requestLocale }) => {
  const requested = await requestLocale;
  const locale = hasLocale(routing.locales, requested) ? requested : routing.defaultLocale;
  const [messages, defaults] = await Promise.all([
    import(`../../messages/${locale}.json`).then((module) => module.default),
    import("../../messages/en.json").then((module) => module.default),
  ]);

  return {
    locale,
    messages: mergeMessages(defaults, messages),
  };
});
