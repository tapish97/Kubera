import Link from "next/link";
import { AppShell } from "@/components/AppShell";
import { PageHeader } from "@/components/PageHeader";
import { getTranslations } from "next-intl/server";

export default async function NewSalePage({ params }: PageProps<"/[locale]/sales/new">) { const { locale } = await params; const t = await getTranslations({ locale, namespace: "QuickAdd" }); return <AppShell><div className="px-5 pt-[max(1.25rem,env(safe-area-inset-top))]"><PageHeader title={t("sell")} back/><div className="mt-8 rounded-[24px] border border-[#e7e1d5] bg-white p-6"><p className="font-bold">{t("setupTitle")}</p><p className="mt-2 text-sm leading-6 text-[#6f746d]">{t("sellSetup")}</p><Link href={`/${locale}/fruits`} className="mt-5 block min-h-12 rounded-2xl bg-[#216148] px-4 py-3 text-center font-bold text-white">{t("manageFruits")}</Link></div></div></AppShell>; }
