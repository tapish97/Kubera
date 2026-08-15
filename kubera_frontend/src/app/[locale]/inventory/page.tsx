import Link from "next/link";
import { redirect } from "next/navigation";
import { getTranslations } from "next-intl/server";
import { AppShell } from "@/components/AppShell";
import { PageHeader } from "@/components/PageHeader";
import { DashboardIcon } from "@/components/DashboardIcon";
import { apiFetch } from "@/lib/api";
import { auth } from "@/lib/auth/server";
import { formatMoney, formatQuantity } from "@/lib/format";

export const dynamic = "force-dynamic";
type Item = { batch_id: string; fruit: string; supplier: string; mark: string; quality: string | null; size: string; quantity_remaining: number; unit: string; purchase_price_per_unit: number | null };
type Account = { currency: string };
export default async function InventoryPage({ params }: PageProps<"/[locale]/inventory">) { const { locale } = await params; const t = await getTranslations({ locale, namespace: "Inventory" }); const { data: session } = await auth.getSession(); if (!session?.user) redirect(`/${locale}/login`); const [items, account] = await Promise.all([apiFetch<Item[]>("/inventory"), apiFetch<Account>("/me")]); return <AppShell><div className="px-5 pt-[max(1.25rem,env(safe-area-inset-top))]"><PageHeader title={t("title")} actions={<Link href={`/${locale}/stock/add`} className="grid h-11 w-11 place-items-center rounded-full bg-[#216148] text-white"><DashboardIcon name="plus"/></Link>}/><p className="mt-2 text-sm text-[#6f746d]">{t("subtitle")}</p><section className="mt-6 space-y-3">{items.length === 0 ? <div className="rounded-[24px] border border-dashed border-[#d8d2c6] px-6 py-12 text-center"><p className="font-bold">{t("empty")}</p><p className="mt-1 text-sm text-[#777b73]">{t("emptyHint")}</p><Link href={`/${locale}/stock/add`} className="mt-5 inline-block rounded-2xl bg-[#216148] px-5 py-3 font-bold text-white">{t("buy")}</Link></div> : items.map((item) => <article key={item.batch_id} className="rounded-[22px] border border-[#e7e1d5] bg-white p-4"><div className="flex items-start justify-between gap-3"><div><h2 className="font-bold">{item.fruit}</h2><p className="mt-1 text-xs text-[#777b73]">{item.mark} · {item.supplier}</p></div><p className="text-lg font-bold text-[#216148]">{formatQuantity(item.quantity_remaining, locale, item.unit)}</p></div><div className="mt-3 flex flex-wrap gap-2 text-xs text-[#6f746d]">{item.quality && <span className="rounded-full bg-[#edf4ef] px-2.5 py-1">{item.quality}</span>}<span className="rounded-full bg-[#f1eee7] px-2.5 py-1">{item.size}</span>{item.purchase_price_per_unit != null && <span className="rounded-full bg-[#fff1d9] px-2.5 py-1">{formatMoney(item.purchase_price_per_unit, locale, account.currency)} / {item.unit}</span>}</div></article>)}</section></div></AppShell>; }
