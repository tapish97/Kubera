"use client";

import { useState } from "react";
import { createPortal } from "react-dom";
import Link from "next/link";
import { useParams, usePathname } from "next/navigation";
import { useTranslations } from "next-intl";
import { DashboardIcon } from "@/components/DashboardIcon";

const items = [
  { key: "home", icon: "home" as const, route: "dashboard" },
  { key: "inventory", icon: "inventory" as const, route: "inventory" },
  { key: "add", icon: "plus" as const, primary: true },
  { key: "sales", icon: "sale" as const, route: "sales" },
  { key: "shop", icon: "more" as const, route: "shop" },
];

function AddMenu({ locale }: { locale: string }) {
  const t = useTranslations("QuickAdd");
  const [open, setOpen] = useState(false);
  const sheet = open ? createPortal(
    <div className="fixed inset-0 z-50 flex items-end justify-center pb-24 sm:items-center sm:pb-0">
      <button type="button" aria-label={t("close")} onClick={() => setOpen(false)} className="absolute inset-0 bg-[#101a15]/55 backdrop-blur-[2px]" />
      <section role="dialog" aria-modal="true" aria-labelledby="quick-add-title" className="relative mx-2 w-full max-w-lg rounded-[28px] bg-[#fffdf8] px-5 pb-5 pt-4 shadow-2xl">
        <div className="mx-auto mb-5 h-1.5 w-11 rounded-full bg-[#d9d4ca] sm:hidden" />
        <h2 id="quick-add-title" className="text-xl font-bold text-[#20241f]">{t("title")}</h2>
        <p className="mt-1 text-sm text-[#6f746d]">{t("subtitle")}</p>
        <div className="mt-5 grid grid-cols-2 gap-3">
          <Link href={`/${locale}/stock/add`} onClick={() => setOpen(false)} className="rounded-[22px] bg-[#216148] p-5 text-white shadow-lg"><DashboardIcon name="plus" className="h-7 w-7" /><span className="mt-6 block font-bold">{t("buy")}</span><span className="mt-1 block text-xs text-white/65">{t("buyHint")}</span></Link>
          <Link href={`/${locale}/sales/new`} onClick={() => setOpen(false)} className="rounded-[22px] bg-[#f1bb5d] p-5 text-[#4e350d]"><DashboardIcon name="sale" className="h-7 w-7" /><span className="mt-6 block font-bold">{t("sell")}</span><span className="mt-1 block text-xs text-[#6d501e]">{t("sellHint")}</span></Link>
        </div>
      </section>
    </div>, document.body,
  ) : null;
  return <><button type="button" onClick={() => setOpen(true)} aria-label={t("open")} aria-expanded={open} className="flex min-h-12 flex-col items-center justify-center gap-1 text-[11px] font-semibold text-[#216148]"><span className="-mt-7 grid h-14 w-14 place-items-center rounded-full bg-[#216148] text-white shadow-[0_10px_25px_rgba(33,97,72,0.28)] ring-4 ring-[#fffdf8]"><DashboardIcon name="plus" className="h-7 w-7" /></span><span className="mt-0.5">{t("label")}</span></button>{sheet}</>;
}

export function MobileBottomNav({ pendingPriceCount = 0 }: { pendingPriceCount?: number }) {
  const { locale } = useParams<{ locale: string }>(); const pathname = usePathname(); const t = useTranslations("Navigation");
  return <nav aria-label="Primary navigation" className="fixed inset-x-0 bottom-0 z-[60] mx-auto max-w-lg border-t border-[#e7e1d5] bg-[#fffdf8]/95 px-3 pb-[max(0.75rem,env(safe-area-inset-bottom))] pt-2 backdrop-blur-xl"><div className="grid grid-cols-5 items-end">{items.map((item) => {
    if (item.primary) return <AddMenu key={item.key} locale={locale} />;
    const active = item.route ? pathname.endsWith(`/${item.route}`) : false; const className = `flex min-h-12 flex-col items-center justify-center gap-1 text-[11px] font-semibold ${active ? "text-[#216148]" : "text-[#8b877e]"}`;
    return <Link key={item.key} href={`/${locale}/${item.route}`} aria-current={active ? "page" : undefined} className={className}><span className="relative grid h-7 place-items-center"><DashboardIcon name={item.icon} className="h-5 w-5" />{item.key === "shop" && pendingPriceCount > 0 ? <span className="absolute right-0 top-0 h-2.5 w-2.5 rounded-full bg-red-500 ring-2 ring-[#fffdf8]" aria-label={t("pendingPrices", { count: pendingPriceCount })} /> : null}</span><span>{t(item.key)}</span></Link>;
  })}</div></nav>;
}
