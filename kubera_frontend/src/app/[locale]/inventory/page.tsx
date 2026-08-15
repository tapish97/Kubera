import Link from "next/link";
import { redirect } from "next/navigation";
import { getTranslations } from "next-intl/server";
import { AppShell } from "@/components/AppShell";
import { DashboardIcon } from "@/components/DashboardIcon";
import { PageHeader } from "@/components/PageHeader";
import { CompactInventory, type InventoryItem } from "@/features/inventory/CompactInventory";
import { apiFetch } from "@/lib/api";
import { auth } from "@/lib/auth/server";

export const dynamic = "force-dynamic";
type Account = { currency:string; timezone:string };

export default async function InventoryPage({ params }:PageProps<"/[locale]/inventory">) {
  const {locale}=await params; const t=await getTranslations({locale,namespace:"Inventory"}); const {data:session}=await auth.getSession(); if(!session?.user) redirect(`/${locale}/login`);
  const [items,account]=await Promise.all([apiFetch<InventoryItem[]>("/inventory"),apiFetch<Account>("/me")]);
  return <AppShell><div className="px-4 pt-[max(1rem,env(safe-area-inset-top))]">
    <PageHeader title={t("title")} actions={<Link href={`/${locale}/stock/add`} aria-label={t("buy")} className="grid h-10 w-10 place-items-center rounded-full bg-[#216148] text-white"><DashboardIcon name="plus" /></Link>} />
    <p className="mt-1 text-xs text-[#6f746d]">{t("compactHint")}</p>
    <section className="mt-4">{items.length ? <CompactInventory items={items} locale={locale} currency={account.currency} timezone={account.timezone} /> : <div className="rounded-[20px] border border-dashed border-[#d8d2c6] px-6 py-10 text-center"><p className="font-bold">{t("empty")}</p><p className="mt-1 text-sm text-[#777b73]">{t("emptyHint")}</p><Link href={`/${locale}/stock/add`} className="mt-4 inline-block rounded-xl bg-[#216148] px-4 py-2.5 font-bold text-white">{t("buy")}</Link></div>}</section>
  </div></AppShell>;
}
