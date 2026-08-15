"use client";

import { DashboardIcon } from "@/components/DashboardIcon";
import Link from "next/link";
import { useParams, usePathname } from "next/navigation";
import { useTranslations } from "next-intl";

const items = [
  { key: "home", icon: "home" as const, route: "dashboard" },
  { key: "inventory", icon: "inventory" as const },
  { key: "add", icon: "plus" as const, primary: true },
  { key: "sales", icon: "sale" as const },
  { key: "more", icon: "more" as const, route: "more" },
];

export function MobileBottomNav() {
  const { locale } = useParams<{ locale: string }>();
  const pathname = usePathname();
  const t = useTranslations("Navigation");

  return (
    <nav aria-label="Primary navigation" className="fixed inset-x-0 bottom-0 z-30 mx-auto max-w-lg border-t border-[#e7e1d5] bg-[#fffdf8]/95 px-3 pb-[max(0.75rem,env(safe-area-inset-bottom))] pt-2 backdrop-blur-xl">
      <div className="grid grid-cols-5 items-end">
        {items.map((item) => {
          const active = item.route ? pathname.endsWith(`/${item.route}`) : false;
          const className = `flex min-h-12 flex-col items-center justify-center gap-1 text-[11px] font-semibold ${active ? "text-[#216148]" : "text-[#8b877e]"}`;
          const content = <>
            <span className={item.primary ? "-mt-7 grid h-14 w-14 place-items-center rounded-full bg-[#216148] text-white shadow-[0_10px_25px_rgba(33,97,72,0.28)] ring-4 ring-[#fffdf8]" : "grid h-7 place-items-center"}>
              <DashboardIcon name={item.icon} className={item.primary ? "h-7 w-7" : "h-5 w-5"} />
            </span>
            <span className={item.primary ? "mt-0.5 text-[#216148]" : ""}>{t(item.key)}</span>
          </>;

          return item.route ? (
            <Link key={item.key} href={`/${locale}/${item.route}`} aria-current={active ? "page" : undefined} className={className}>{content}</Link>
          ) : (
            <button key={item.key} type="button" disabled aria-label={`${t(item.key)}, coming soon`} className={className}>{content}</button>
          );
        })}
      </div>
    </nav>
  );
}
