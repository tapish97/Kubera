import { DashboardIcon } from "@/components/DashboardIcon";
import { AppShell } from "@/components/AppShell";
import { ProfileMenu } from "@/components/ProfileMenu";
import { LanguageSwitcher } from "@/components/LanguageSwitcher";
import { ApiError, apiFetch } from "@/lib/api";
import { auth } from "@/lib/auth/server";
import { redirect } from "next/navigation";
import { getTranslations } from "next-intl/server";
import { formatDate, formatMoney, formatQuantity } from "@/lib/format";
import Link from "next/link";

export const dynamic = "force-dynamic";

type CurrentAccount = { profile_id: string; shop_id: string; shop_name: string; currency: string; onboarding_completed_at: string | null };
type Summary = { total_inventory_quantity: number; today_sales: number; today_gross_profit: number };
type RecentSale = { sale_id: string; sold_at: string; fruit: string; mark: string; quantity: number; unit: string; total_sale_amount: number };
type InventoryItem = { batch_id: string; fruit: string; quality: string | null; size: string; quantity_remaining: number; unit: string };
type StockHighlight = { label: string; quantity: number; unit: string };

const emptySummary: Summary = { total_inventory_quantity: 0, today_sales: 0, today_gross_profit: 0 };

function stockHighlights(items: InventoryItem[]): StockHighlight[] {
  const grouped = new Map<string, StockHighlight>();
  for (const item of items) {
    const details = [item.quality?.trim(), item.size !== "normal" ? item.size : ""].filter(Boolean).join(", ");
    const label = details ? `${item.fruit} (${details})` : item.fruit;
    const key = `${label.toLocaleLowerCase()}::${item.unit}`;
    const current = grouped.get(key) ?? { label, quantity: 0, unit: item.unit };
    current.quantity += Number(item.quantity_remaining); grouped.set(key, current);
  }
  return [...grouped.values()];
}

async function loadShopData() {
  try {
    const account = await apiFetch<CurrentAccount>("/me");
    const [summary, recentSales, inventory] = await Promise.all([
      apiFetch<Summary>("/dashboard/summary"),
      apiFetch<RecentSale[]>("/dashboard/recent-sales"),
      apiFetch<InventoryItem[]>("/inventory"),
    ]);
    return { account, summary, recentSales, stock: stockHighlights(inventory), connected: true };
  } catch (error) {
	const detail = error instanceof Error ? error.message : "";
    return {
      account: null,
      summary: emptySummary,
      recentSales: [] as RecentSale[],
	  stock: [] as StockHighlight[],
      connected: false,
      connectionError: error instanceof ApiError && error.status === 401 ? "shopSession" : "shopUnavailable",
	  connectionDetail: detail,
    };
  }
}

export default async function DashboardPage({ params }: PageProps<"/[locale]/dashboard">) {
  const { locale } = await params;
  const t = await getTranslations({ locale, namespace: "Dashboard" });
  const common = await getTranslations({ locale, namespace: "Common" });
  const errors = await getTranslations({ locale, namespace: "Errors" });
  const { data: session } = await auth.getSession();
  if (!session?.user) redirect(`/${locale}/login`);

  const data = await loadShopData();
  if (data.account && !data.account.onboarding_completed_at) {
    redirect(`/${locale}/onboarding`);
  }

  const firstName = session.user.name?.trim().split(/\s+/)[0] || "there";
  const currency = data.account?.currency || "INR";
  const date = formatDate(new Date(), locale, { weekday: "long", day: "numeric", month: "short" });

  return (
    <AppShell>
        <header className="relative overflow-hidden bg-[#173f31] px-5 pb-8 pt-[max(1.25rem,env(safe-area-inset-top))] text-white">
          <div className="absolute -right-16 -top-20 h-52 w-52 rounded-full border-[35px] border-white/5" />
          <div className="absolute -bottom-20 left-10 h-40 w-40 rounded-full bg-[#e4a94b]/10 blur-2xl" />
          <div className="relative flex items-center justify-between">
            <div>
              <p className="text-[11px] font-bold uppercase tracking-[0.22em] text-[#b9d0c3]">{common("kubera")}</p>
              <p className="mt-1 max-w-64 truncate text-sm font-medium text-white/80">{data.account?.shop_name ?? "My Shop"}</p>
            </div>
            <div className="flex items-center gap-2"><LanguageSwitcher inverted /><ProfileMenu
              name={session.user.name ?? ""}
              email={session.user.email}
              shopName={data.account?.shop_name ?? "My Shop"}
            /></div>
          </div>
          <div className="relative mt-8">
            <p className="text-sm capitalize text-[#b9d0c3]">{date}</p>
            <h1 className="mt-1 text-[28px] font-semibold tracking-[-0.03em]">{t("goodDay", { name: firstName })}</h1>
            <p className="mt-2 text-sm text-white/65">{t("subtitle")}</p>
          </div>
        </header>

        {!data.connected && (
          <div className="mx-5 -mt-3 rounded-2xl border border-[#ecd7ae] bg-[#fff8e8] px-4 py-3 text-sm text-[#75551c] shadow-sm">
            <p className="font-semibold">{t("connectingTitle")}</p>
            <p className="mt-0.5 text-xs leading-5 text-[#8a6a32]">{errors(data.connectionError ?? "shopUnavailable")}</p>
			{data.connectionDetail && <p className="mt-1 text-[11px] leading-4 text-[#8a6a32]">{data.connectionDetail}</p>}
          </div>
        )}

        <section aria-labelledby="overview-heading" className="px-5 pt-6">
          <div className="flex items-end justify-between">
            <div><p className="text-xs font-bold uppercase tracking-[0.16em] text-[#8a8d84]">{t("today")}</p><h2 id="overview-heading" className="mt-1 text-xl font-semibold tracking-tight">{t("overview")}</h2></div>
            <span className={`rounded-full px-3 py-1 text-xs font-bold ${data.connected ? "bg-[#e5f1e9] text-[#216148]" : "bg-[#fff1d9] text-[#8b6427]"}`}>{data.connected ? common("live") : common("waiting")}</span>
          </div>
          <div className="mt-4 grid grid-cols-2 gap-3">
            <article className="rounded-[22px] bg-[#edf4ef] p-4">
              <div className="flex items-center gap-2 text-[#3f6655]"><DashboardIcon name="trend" className="h-4 w-4" /><p className="text-xs font-semibold">{t("sales")}</p></div>
              <p className="mt-4 text-2xl font-bold tracking-[-0.04em] text-[#173f31]">{formatMoney(data.summary.today_sales, locale, currency)}</p><p className="mt-1 text-xs text-[#718078]">{t("revenue")}</p>
            </article>
            <article className="rounded-[22px] bg-[#fff1d9] p-4">
              <div className="flex items-center gap-2 text-[#8b6427]"><DashboardIcon name="sparkle" className="h-4 w-4" /><p className="text-xs font-semibold">{t("profit")}</p></div>
              <p className="mt-4 text-2xl font-bold tracking-[-0.04em] text-[#634411]">{formatMoney(data.summary.today_gross_profit, locale, currency)}</p><p className="mt-1 text-xs text-[#8f7957]">{t("grossProfit")}</p>
            </article>
          </div>
        </section>

        <section aria-labelledby="remaining-heading" className="px-5 pt-8">
          <div className="flex items-center justify-between"><div><p className="text-xs font-bold uppercase tracking-[0.16em] text-[#8a8d84]">{t("quickView")}</p><h2 id="remaining-heading" className="mt-1 text-lg font-semibold tracking-tight">{t("remainingStock")}</h2></div><Link href={`/${locale}/inventory`} className="text-xs font-bold text-[#216148]">{t("viewStock")} →</Link></div>
          <div className="mt-3 overflow-hidden rounded-[22px] border border-[#e7e1d5] bg-white">
            {data.stock.length === 0 ? <div className="px-5 py-7 text-center"><p className="text-sm font-semibold">{t("noStock")}</p><Link href={`/${locale}/stock/add`} className="mt-3 inline-block text-sm font-bold text-[#216148]">{t("buyFirstStock")} →</Link></div> : data.stock.slice(0, 6).map((item) => <div key={`${item.label}-${item.unit}`} className="flex items-center justify-between border-b border-[#eee9df] px-4 py-3.5 last:border-0"><p className="min-w-0 truncate font-bold">{item.label}</p><p className="ml-3 shrink-0 font-bold text-[#216148]">{formatQuantity(item.quantity, locale, item.unit)}</p></div>)}
          </div>
        </section>

        <section aria-labelledby="recent-heading" className="px-5 pt-8">
          <div className="flex items-center justify-between"><h2 id="recent-heading" className="text-lg font-semibold tracking-tight">{t("recentSales")}</h2><Link href={`/${locale}/sales`} className="flex items-center gap-1 text-xs font-bold text-[#216148]">{t("viewAll")} <DashboardIcon name="arrow" className="h-4 w-4" /></Link></div>
          <div className="mt-3 overflow-hidden rounded-[22px] border border-[#e7e1d5] bg-white">
            {data.recentSales.length === 0 ? (
              <div className="px-5 py-8 text-center"><div className="mx-auto grid h-11 w-11 place-items-center rounded-full bg-[#f1eee7] text-[#8a8d84]"><DashboardIcon name="sale" /></div><p className="mt-3 text-sm font-semibold">{t("noSales")}</p><p className="mt-1 text-xs text-[#8a8d84]">{t("noSalesHint")}</p></div>
            ) : data.recentSales.slice(0, 4).map((sale) => (
              <article key={`${sale.sale_id}-${sale.fruit}`} className="flex items-center justify-between border-b border-[#eee9df] px-4 py-3.5 last:border-0"><div className="min-w-0"><p className="truncate text-sm font-bold">{sale.mark} · {sale.fruit}</p><p className="mt-0.5 text-xs text-[#858980]">{formatQuantity(sale.quantity, locale, sale.unit)}</p></div><p className="ml-3 text-sm font-bold text-[#216148]">{formatMoney(sale.total_sale_amount, locale, currency)}</p></article>
            ))}
          </div>
        </section>

        <section className="mx-5 mt-8 rounded-[22px] bg-[#232823] p-5 text-white">
          <div className="flex items-start gap-3"><div className="grid h-10 w-10 shrink-0 place-items-center rounded-2xl bg-white/10 text-[#f1bb5d]"><DashboardIcon name="sparkle" /></div><div><p className="text-sm font-bold">{t("insights")}</p><p className="mt-1 text-xs leading-5 text-white/55">{t("insightsHint")}</p><span className="mt-3 inline-block rounded-full bg-white/8 px-2.5 py-1 text-[10px] font-bold uppercase tracking-wider text-white/60">{t("comingLater")}</span></div></div>
        </section>
    </AppShell>
  );
}
