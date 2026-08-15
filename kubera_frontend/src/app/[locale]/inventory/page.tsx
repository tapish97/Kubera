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
type Item = { batch_id: string; fruit: string; supplier: string; mark: string; quality: string | null; size: string; quantity_remaining: number; unit: string; purchase_price_per_unit: number | null; received_at: string };
type Account = { currency: string; timezone: string };
type Group = { fruit: string; unit: string; total: number; items: Item[] };

function groupStock(items: Item[]) {
  const groups = new Map<string, Group>();
  for (const item of items) {
    const key = `${item.fruit.toLocaleLowerCase()}::${item.unit}`;
    const group = groups.get(key) ?? { fruit: item.fruit, unit: item.unit, total: 0, items: [] };
    group.total += Number(item.quantity_remaining);
    group.items.push(item);
    groups.set(key, group);
  }
  return [...groups.values()].map((group) => ({ ...group, items: group.items.sort((a, b) => new Date(b.received_at).getTime() - new Date(a.received_at).getTime()) }));
}

function batchAge(receivedAt: string) {
  return Math.max(0, Math.floor((Date.now() - new Date(receivedAt).getTime()) / 86_400_000));
}

export default async function InventoryPage({ params }: PageProps<"/[locale]/inventory">) {
  const { locale } = await params;
  const t = await getTranslations({ locale, namespace: "Inventory" });
  const { data: session } = await auth.getSession();
  if (!session?.user) redirect(`/${locale}/login`);
  const [items, account] = await Promise.all([apiFetch<Item[]>("/inventory"), apiFetch<Account>("/me")]);
  const groups = groupStock(items);

  return <AppShell><div className="px-5 pt-[max(1.25rem,env(safe-area-inset-top))]">
    <PageHeader title={t("title")} actions={<Link href={`/${locale}/stock/add`} aria-label={t("buy")} className="grid h-11 w-11 place-items-center rounded-full bg-[#216148] text-white"><DashboardIcon name="plus" /></Link>} />
    <p className="mt-2 text-sm text-[#6f746d]">{t("subtitle")}</p>
    <section className="mt-6 space-y-4">{groups.length === 0 ? <div className="rounded-[24px] border border-dashed border-[#d8d2c6] px-6 py-12 text-center">
      <p className="font-bold">{t("empty")}</p><p className="mt-1 text-sm text-[#777b73]">{t("emptyHint")}</p><Link href={`/${locale}/stock/add`} className="mt-5 inline-block rounded-2xl bg-[#216148] px-5 py-3 font-bold text-white">{t("buy")}</Link>
    </div> : groups.map((group) => <article key={`${group.fruit}-${group.unit}`} className="overflow-hidden rounded-[24px] border border-[#dce6df] bg-white shadow-[0_8px_24px_rgba(44,53,43,0.05)]">
      <header className="flex items-center justify-between gap-3 bg-[#edf4ef] px-4 py-4"><div><p className="text-xs font-bold uppercase tracking-wider text-[#648072]">{t("fruitType")}</p><h2 className="mt-0.5 text-lg font-bold text-[#173f31]">{group.fruit}</h2></div><div className="text-right"><p className="text-xs font-bold text-[#648072]">{t("totalAvailable")}</p><p className="mt-0.5 text-xl font-bold text-[#216148]">{formatQuantity(group.total, locale, group.unit)}</p></div></header>
      <div className="divide-y divide-[#eee9df]">{group.items.map((item, index) => {
        const age = batchAge(item.received_at);
        return <div key={item.batch_id} className="p-4">
          <div className="mb-3 flex items-center justify-between gap-2"><span className="rounded-full bg-[#e7f1eb] px-2.5 py-1 text-xs font-bold text-[#216148]">{t("batch", { number: index + 1 })}</span><span className="text-xs text-[#777b73]">{age === 0 ? t("addedToday") : age === 1 ? t("oneDayOld") : t("daysOld", { count: age })}</span></div>
          <div className="flex items-start justify-between gap-3"><div><p className="font-bold text-[#20241f]">{t("quality")}: {item.quality?.trim() || t("notSpecified")}</p><p className="mt-1 text-xs text-[#777b73]">{t("mark")}: {item.mark} · {item.supplier}</p></div><p className="shrink-0 font-bold text-[#216148]">{formatQuantity(item.quantity_remaining, locale, item.unit)}</p></div>
          <p className="mt-2 text-xs text-[#777b73]">{t("added")}: {formatDate(item.received_at, locale, { day: "numeric", month: "short", year: "numeric", hour: "numeric", minute: "2-digit", timeZone: account.timezone })}</p>
          <div className="mt-3 flex flex-wrap gap-2 text-xs text-[#6f746d]"><span className="rounded-full bg-[#f1eee7] px-2.5 py-1">{t("size")}: {item.size}</span>{item.purchase_price_per_unit != null && <span className="rounded-full bg-[#fff1d9] px-2.5 py-1">{t("buyPrice")}: {formatMoney(item.purchase_price_per_unit, locale, account.currency)} / {item.unit}</span>}</div>
        </div>;
      })}</div>
    </article>)}</section>
  </div></AppShell>;
}
