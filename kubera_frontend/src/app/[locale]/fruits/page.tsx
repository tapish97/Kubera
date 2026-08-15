import { redirect } from "next/navigation";
import { getTranslations } from "next-intl/server";
import { AppShell } from "@/components/AppShell";
import { PageHeader } from "@/components/PageHeader";
import { FruitManager } from "@/features/fruits/FruitManager";
import type { Fruit } from "@/features/fruits/actions";
import { apiFetch } from "@/lib/api";
import { auth } from "@/lib/auth/server";

export const dynamic = "force-dynamic";
type Account = { onboarding_completed_at: string | null };

export default async function FruitsPage({ params }: PageProps<"/[locale]/fruits">) {
  const { locale } = await params; const t = await getTranslations({ locale, namespace: "Fruits" });
  const { data: session } = await auth.getSession(); if (!session?.user) redirect(`/${locale}/login`);
  const [account, fruits] = await Promise.all([apiFetch<Account>("/me"), apiFetch<Fruit[]>("/fruits?include_inactive=true")]);
  if (!account.onboarding_completed_at) redirect(`/${locale}/onboarding`);
  return <AppShell><div className="px-5 pt-[max(1.25rem,env(safe-area-inset-top))]"><PageHeader eyebrow={t("eyebrow")} title={t("title")} back /><p className="mt-2 text-sm leading-6 text-[#6f746d]">{t("subtitle")}</p><section className="mt-6"><FruitManager fruits={fruits} locale={locale} /></section></div></AppShell>;
}
