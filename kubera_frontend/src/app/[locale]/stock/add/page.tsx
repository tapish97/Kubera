import { redirect } from "next/navigation";
import { getTranslations } from "next-intl/server";
import { AppShell } from "@/components/AppShell";
import { PageHeader } from "@/components/PageHeader";
import { BuyStockForm } from "@/features/stock/BuyStockForm";
import { apiFetch } from "@/lib/api";
import { auth } from "@/lib/auth/server";

export const dynamic = "force-dynamic";
type Fruit = { id: string; name: string; default_unit: string };
type Supplier = { id: string; name: string; mark: string };
type Account = { currency: string };
export default async function AddStockPage({ params }: PageProps<"/[locale]/stock/add">) {
  const { locale } = await params; const t = await getTranslations({ locale, namespace: "QuickAdd" }); const { data: session } = await auth.getSession(); if (!session?.user) redirect(`/${locale}/login`);
  const [fruits, suppliers, account] = await Promise.all([apiFetch<Fruit[]>("/fruits"), apiFetch<Supplier[]>("/suppliers"), apiFetch<Account>("/me")]);
  return <AppShell><div className="px-5 pt-[max(1.25rem,env(safe-area-inset-top))]"><PageHeader title={t("buy")} back/><p className="mt-2 text-sm text-[#6f746d]">{t("buyHint")}</p><section className="mt-6 rounded-[24px] border border-[#e7e1d5] bg-white p-5 shadow-sm"><BuyStockForm locale={locale} currency={account.currency} fruits={fruits} suppliers={suppliers}/></section></div></AppShell>;
}
