"use client";

import { useState } from "react";
import Link from "next/link";
import { useParams } from "next/navigation";
import { useTranslations } from "next-intl";
import { DashboardIcon } from "@/components/DashboardIcon";

export function QuickActionMenu() {
  const { locale } = useParams<{ locale: string }>(); const t = useTranslations("QuickAdd"); const [open, setOpen] = useState(false);
  return <>
    <button type="button" onClick={() => setOpen(true)} aria-label={t("open")} aria-expanded={open} className="flex min-h-12 flex-col items-center justify-center gap-1 text-[11px] font-semibold text-[#216148]"><span className="-mt-7 grid h-14 w-14 place-items-center rounded-full bg-[#216148] text-white shadow-[0_10px_25px_rgba(33,97,72,0.28)] ring-4 ring-[#fffdf8]"><DashboardIcon name="plus" className="h-7 w-7" /></span><span className="mt-0.5">{t("label")}</span></button>
    {open && <div className="fixed inset-0 z-50 flex items-end justify-center sm:items-center"><button type="button" aria-label={t("close")} onClick={() => setOpen(false)} className="absolute inset-0 bg-[#101a15]/55 backdrop-blur-[2px]"/><section role="dialog" aria-modal="true" aria-labelledby="quick-add-title" className="relative w-full max-w-lg rounded-t-[28px] bg-[#fffdf8] px-5 pb-[max(1.5rem,env(safe-area-inset-bottom))] pt-4 shadow-2xl sm:rounded-[28px]"><div className="mx-auto mb-5 h-1.5 w-11 rounded-full bg-[#d9d4ca] sm:hidden"/><h2 id="quick-add-title" className="text-xl font-bold text-[#20241f]">{t("title")}</h2><p className="mt-1 text-sm text-[#6f746d]">{t("subtitle")}</p><div className="mt-5 grid grid-cols-2 gap-3"><Link href={`/${locale}/stock/add`} onClick={() => setOpen(false)} className="rounded-[22px] bg-[#216148] p-5 text-white shadow-lg"><DashboardIcon name="plus" className="h-7 w-7"/><span className="mt-6 block font-bold">{t("buy")}</span><span className="mt-1 block text-xs text-white/65">{t("buyHint")}</span></Link><Link href={`/${locale}/sales/new`} onClick={() => setOpen(false)} className="rounded-[22px] bg-[#f1bb5d] p-5 text-[#4e350d]"><DashboardIcon name="sale" className="h-7 w-7"/><span className="mt-6 block font-bold">{t("sell")}</span><span className="mt-1 block text-xs text-[#6d501e]">{t("sellHint")}</span></Link></div><button type="button" onClick={() => setOpen(false)} className="mt-4 min-h-12 w-full rounded-2xl border border-[#ded8cc] bg-white font-bold text-[#60655e]">{t("close")}</button></section></div>}
  </>;
}
