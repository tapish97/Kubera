import Link from "next/link";
import { redirect } from "next/navigation";
import { getTranslations } from "next-intl/server";
import { AppShell } from "@/components/AppShell";
import { DashboardIcon } from "@/components/DashboardIcon";
import { PageHeader } from "@/components/PageHeader";
import { apiFetch } from "@/lib/api";
import { auth } from "@/lib/auth/server";
import { formatDate, formatMoney, formatQuantity } from "@/lib/format";

export const dynamic = "force-dynamic";
type Sale = { sale_id: string; sold_at: string; fruit: string; mark: string; quantity: number; unit: string; total_sale_amount: number; gross_profit: number };
type Account = { currency: string; timezone: string };

export default async function SalesPage({ params }: PageProps<"/[locale]/sales">) {
  const { locale } = await params;
  const t = await getTranslations({ locale, namespace: "SalesHistory" });
  const { data: session } = await auth.getSession();
  if (!session?.user) redirect(`/${locale}/login`);
  const [sales, account] = await Promise.all([apiFetch<Sale[]>("/sales"), apiFetch<Account>("/me")]);
  return <AppShell><div className="px-5 pt-[max(1.25rem,env(safe-area-inset-top))]">
    <PageHeader title={t("title")} actions={<Link href={`/${locale}/sales/new`} className="grid h-11 w-11 place-items-center rounded-full bg-[#f1bb5d] text-[#4e350d]"><DashboardIcon name="plus" /></Link>} />
    <p className="mt-2 text-sm text-[#6f746d]">{t("subtitle")}</p>
    <section className="mt-6 space-y-3">{sales.length === 0 ? <div className="rounded-[24px] border border-dashed border-[#d8d2c6] px-6 py-12 text-center"><p className="font-bold">{t("empty")}</p><p className="mt-1 text-sm text-[#777b73]">{t("emptyHint")}</p></div> : sales.map((sale, index) => <article key={`${sale.sale_id}-${index}`} className="rounded-[22px] border border-[#e7e1d5] bg-white p-4">
      <div className="flex justify-between gap-3"><div><h2 className="font-bold">{sale.fruit}</h2><p className="mt-1 text-xs text-[#777b73]">{sale.mark} · {t("sold")}: {formatDate(sale.sold_at, locale, { day: "numeric", month: "short", year: "numeric", hour: "numeric", minute: "2-digit", timeZone: account.timezone })}</p></div><p className="font-bold text-[#216148]">{formatMoney(sale.total_sale_amount, locale, account.currency)}</p></div>
      <div className="mt-3 flex justify-between text-xs text-[#6f746d]"><span>{formatQuantity(sale.quantity, locale, sale.unit)}</span><span>{t("profit")}: {formatMoney(sale.gross_profit, locale, account.currency)}</span></div>
    </article>)}</section>
  </div></AppShell>;
}
