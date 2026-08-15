import { redirect } from "next/navigation";
import { getTranslations } from "next-intl/server";
import { AppShell } from "@/components/AppShell";
import { PageHeader } from "@/components/PageHeader";
import { SellForm } from "@/features/sales/SellForm";
import { apiFetch } from "@/lib/api";
import { auth } from "@/lib/auth/server";

export const dynamic = "force-dynamic";
type Batch = { batch_id: string; fruit: string; mark: string; quality: string | null; size: string; quantity_remaining: number; unit: string; received_at: string };
type Account = { currency: string; timezone: string };
export default async function NewSalePage({ params }: PageProps<"/[locale]/sales/new">) { const { locale } = await params; const t = await getTranslations({ locale, namespace: "QuickAdd" }); const { data: session } = await auth.getSession(); if (!session?.user) redirect(`/${locale}/login`); const [batches, account] = await Promise.all([apiFetch<Batch[]>("/inventory"), apiFetch<Account>("/me")]); const oldestFirst = batches.sort((a, b) => new Date(a.received_at).getTime() - new Date(b.received_at).getTime()); return <AppShell><div className="px-5 pt-[max(1.25rem,env(safe-area-inset-top))]"><PageHeader title={t("sell")} back/><p className="mt-2 text-sm text-[#6f746d]">{t("sellHint")}</p><section className="mt-6 rounded-[24px] border border-[#e7e1d5] bg-white p-5 shadow-sm"><SellForm locale={locale} currency={account.currency} timezone={account.timezone} batches={oldestFirst}/></section></div></AppShell>; }
