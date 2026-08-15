import { redirect } from "next/navigation";
import { AppShell } from "@/components/AppShell";
import { PageHeader } from "@/components/PageHeader";
import { SettlementList, type UnpricedBatch } from "@/features/settlements/SettlementList";
import { apiFetch } from "@/lib/api";
import { auth } from "@/lib/auth/server";
import { getTranslations } from "next-intl/server";

export const dynamic = "force-dynamic";
type Account = { currency: string; timezone: string };
export default async function SettlementsPage({ params }: PageProps<"/[locale]/settlements">) { const { locale } = await params; const t=await getTranslations({locale,namespace:"Settlements"}); const { data: session } = await auth.getSession(); if (!session?.user) redirect(`/${locale}/login`); const [batches, account] = await Promise.all([apiFetch<UnpricedBatch[]>("/inventory/unpriced"), apiFetch<Account>("/me")]); return <AppShell><div className="px-5 pt-[max(1.25rem,env(safe-area-inset-top))]"><PageHeader title={t("title")} back /><p className="mt-2 text-sm leading-6 text-[#6f746d]">{t("subtitle")}</p><section className="mt-6"><SettlementList batches={batches} locale={locale} currency={account.currency} timezone={account.timezone} /></section></div></AppShell>; }
