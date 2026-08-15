"use client";

import { useLocale, useTranslations } from "next-intl";
import { usePathname, useRouter } from "next/navigation";

const locales = [
  { code: "en", label: "EN" },
  { code: "hi", label: "हिं" },
  { code: "mr", label: "मर" },
];

export function LanguageSwitcher({ inverted = false }: { inverted?: boolean }) {
  const locale = useLocale();
  const pathname = usePathname();
  const router = useRouter();
  const t = useTranslations("Common");

  function changeLocale(nextLocale: string) {
    const segments = pathname.split("/");
    segments[1] = nextLocale;
    router.replace(segments.join("/") || `/${nextLocale}`);
  }

  return (
    <div aria-label={t("language")} className={`inline-flex rounded-full p-1 ${inverted ? "bg-white/10" : "bg-[#eee9df]"}`}>
      {locales.map((item) => (
        <button key={item.code} type="button" onClick={() => changeLocale(item.code)} aria-pressed={locale === item.code} className={`min-h-8 rounded-full px-2.5 text-[11px] font-bold transition ${locale === item.code ? inverted ? "bg-white text-[#173f31]" : "bg-white text-[#216148] shadow-sm" : inverted ? "text-white/65" : "text-[#777b73]"}`}>
          {item.label}
        </button>
      ))}
    </div>
  );
}
