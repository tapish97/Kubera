import { DashboardIcon } from "@/components/DashboardIcon";
import { MobileBottomNav } from "@/components/MobileBottomNav";
import { ApiError, apiFetch } from "@/lib/api";
import { auth } from "@/lib/auth/server";
import { redirect } from "next/navigation";

export const dynamic = "force-dynamic";

type CurrentAccount = { profile_id: string; shop_id: string; shop_name: string };
type Summary = { total_inventory_quantity: number; today_sales: number; today_gross_profit: number };
type RecentSale = { sale_id: string; sold_at: string; fruit: string; mark: string; quantity: number; unit: string; total_sale_amount: number };

const emptySummary: Summary = { total_inventory_quantity: 0, today_sales: 0, today_gross_profit: 0 };

function money(value: number) {
  return new Intl.NumberFormat("en-IN", { style: "currency", currency: "INR", maximumFractionDigits: 0 }).format(value);
}

async function loadShopData() {
  try {
    const account = await apiFetch<CurrentAccount>("/me");
    const [summary, recentSales] = await Promise.all([
      apiFetch<Summary>("/dashboard/summary"),
      apiFetch<RecentSale[]>("/dashboard/recent-sales"),
    ]);
    return { account, summary, recentSales, connected: true };
  } catch (error) {
    if (error instanceof ApiError && error.status === 401) return { unauthorized: true } as const;
    return { account: null, summary: emptySummary, recentSales: [] as RecentSale[], connected: false };
  }
}

export default async function DashboardPage({ params }: PageProps<"/[locale]/dashboard">) {
  const { locale } = await params;
  const { data: session } = await auth.getSession();
  if (!session?.user) redirect(`/${locale}/login`);

  const data = await loadShopData();
  if ("unauthorized" in data) redirect(`/${locale}/login`);

  const firstName = session.user.name?.trim().split(/\s+/)[0] || "there";
  const initial = firstName.slice(0, 1).toUpperCase();
  const date = new Intl.DateTimeFormat(locale, { weekday: "long", day: "numeric", month: "short" }).format(new Date());

  return (
    <div className="min-h-screen bg-[#f5f1e8] text-[#20241f]">
      <main className="mx-auto min-h-screen max-w-lg overflow-hidden bg-[#fffdf8] pb-28 shadow-[0_0_60px_rgba(44,53,43,0.08)]">
        <header className="relative overflow-hidden bg-[#173f31] px-5 pb-8 pt-[max(1.25rem,env(safe-area-inset-top))] text-white">
          <div className="absolute -right-16 -top-20 h-52 w-52 rounded-full border-[35px] border-white/5" />
          <div className="absolute -bottom-20 left-10 h-40 w-40 rounded-full bg-[#e4a94b]/10 blur-2xl" />
          <div className="relative flex items-center justify-between">
            <div>
              <p className="text-[11px] font-bold uppercase tracking-[0.22em] text-[#b9d0c3]">Kubera</p>
              <p className="mt-1 max-w-64 truncate text-sm font-medium text-white/80">{data.account?.shop_name ?? "My Shop"}</p>
            </div>
            <div className="grid h-11 w-11 place-items-center rounded-full bg-white/12 text-sm font-bold ring-1 ring-white/20">{initial}</div>
          </div>
          <div className="relative mt-8">
            <p className="text-sm capitalize text-[#b9d0c3]">{date}</p>
            <h1 className="mt-1 text-[28px] font-semibold tracking-[-0.03em]">Good day, {firstName}</h1>
            <p className="mt-2 text-sm text-white/65">Here&apos;s how your shop is doing today.</p>
          </div>
        </header>

        {!data.connected && (
          <div className="mx-5 -mt-3 rounded-2xl border border-[#ecd7ae] bg-[#fff8e8] px-4 py-3 text-sm text-[#75551c] shadow-sm">
            <p className="font-semibold">Shop data is connecting</p>
            <p className="mt-0.5 text-xs leading-5 text-[#8a6a32]">Your dashboard is ready. Live figures will appear when the latest API is online.</p>
          </div>
        )}

        <section aria-labelledby="overview-heading" className="px-5 pt-6">
          <div className="flex items-end justify-between">
            <div><p className="text-xs font-bold uppercase tracking-[0.16em] text-[#8a8d84]">Today</p><h2 id="overview-heading" className="mt-1 text-xl font-semibold tracking-tight">Shop overview</h2></div>
            <span className={`rounded-full px-3 py-1 text-xs font-bold ${data.connected ? "bg-[#e5f1e9] text-[#216148]" : "bg-[#fff1d9] text-[#8b6427]"}`}>{data.connected ? "Live" : "Waiting"}</span>
          </div>
          <div className="mt-4 grid grid-cols-2 gap-3">
            <article className="rounded-[22px] bg-[#edf4ef] p-4">
              <div className="flex items-center gap-2 text-[#3f6655]"><DashboardIcon name="trend" className="h-4 w-4" /><p className="text-xs font-semibold">Sales</p></div>
              <p className="mt-4 text-2xl font-bold tracking-[-0.04em] text-[#173f31]">{money(data.summary.today_sales)}</p><p className="mt-1 text-xs text-[#718078]">Today&apos;s revenue</p>
            </article>
            <article className="rounded-[22px] bg-[#fff1d9] p-4">
              <div className="flex items-center gap-2 text-[#8b6427]"><DashboardIcon name="sparkle" className="h-4 w-4" /><p className="text-xs font-semibold">Profit</p></div>
              <p className="mt-4 text-2xl font-bold tracking-[-0.04em] text-[#634411]">{money(data.summary.today_gross_profit)}</p><p className="mt-1 text-xs text-[#8f7957]">Gross profit today</p>
            </article>
            <article className="col-span-2 flex items-center justify-between rounded-[22px] border border-[#e7e1d5] bg-white p-4 shadow-[0_8px_24px_rgba(56,57,48,0.05)]">
              <div><p className="text-xs font-semibold text-[#777b73]">Current inventory</p><p className="mt-1 text-2xl font-bold tracking-[-0.04em]">{data.summary.total_inventory_quantity.toLocaleString("en-IN")}</p><p className="mt-0.5 text-xs text-[#8a8d84]">Across active batches</p></div>
              <div className="grid h-12 w-12 place-items-center rounded-2xl bg-[#edf4ef] text-[#216148]"><DashboardIcon name="box" className="h-6 w-6" /></div>
            </article>
          </div>
        </section>

        <section aria-labelledby="actions-heading" className="px-5 pt-8">
          <h2 id="actions-heading" className="text-lg font-semibold tracking-tight">Quick actions</h2>
          <div className="mt-3 grid grid-cols-2 gap-3">
            <button type="button" className="col-span-2 flex min-h-20 items-center justify-between rounded-[22px] bg-[#216148] px-5 text-left text-white shadow-[0_12px_30px_rgba(33,97,72,0.2)]">
              <span><span className="block text-base font-bold">Add stock</span><span className="mt-1 block text-xs text-white/65">Record a new incoming batch</span></span><span className="grid h-10 w-10 place-items-center rounded-full bg-white/15"><DashboardIcon name="plus" /></span>
            </button>
            <button type="button" className="flex min-h-24 flex-col justify-between rounded-[22px] bg-[#f1bb5d] p-4 text-left text-[#4e350d]"><DashboardIcon name="sale" className="h-6 w-6" /><span><span className="block font-bold">Record sale</span><span className="mt-0.5 block text-xs text-[#6d501e]">Sell from a batch</span></span></button>
            <button type="button" className="flex min-h-24 flex-col justify-between rounded-[22px] border border-[#e7e1d5] bg-white p-4 text-left"><DashboardIcon name="inventory" className="h-6 w-6 text-[#216148]" /><span><span className="block font-bold">Inventory</span><span className="mt-0.5 block text-xs text-[#777b73]">View all stock</span></span></button>
          </div>
        </section>

        <section aria-labelledby="recent-heading" className="px-5 pt-8">
          <div className="flex items-center justify-between"><h2 id="recent-heading" className="text-lg font-semibold tracking-tight">Recent sales</h2><button type="button" className="flex items-center gap-1 text-xs font-bold text-[#216148]">View all <DashboardIcon name="arrow" className="h-4 w-4" /></button></div>
          <div className="mt-3 overflow-hidden rounded-[22px] border border-[#e7e1d5] bg-white">
            {data.recentSales.length === 0 ? (
              <div className="px-5 py-8 text-center"><div className="mx-auto grid h-11 w-11 place-items-center rounded-full bg-[#f1eee7] text-[#8a8d84]"><DashboardIcon name="sale" /></div><p className="mt-3 text-sm font-semibold">No sales recorded yet</p><p className="mt-1 text-xs text-[#8a8d84]">Your latest sales will appear here.</p></div>
            ) : data.recentSales.slice(0, 4).map((sale) => (
              <article key={`${sale.sale_id}-${sale.fruit}`} className="flex items-center justify-between border-b border-[#eee9df] px-4 py-3.5 last:border-0"><div className="min-w-0"><p className="truncate text-sm font-bold">{sale.mark} · {sale.fruit}</p><p className="mt-0.5 text-xs text-[#858980]">{sale.quantity} {sale.unit}</p></div><p className="ml-3 text-sm font-bold text-[#216148]">{money(sale.total_sale_amount)}</p></article>
            ))}
          </div>
        </section>

        <section className="mx-5 mt-8 rounded-[22px] bg-[#232823] p-5 text-white">
          <div className="flex items-start gap-3"><div className="grid h-10 w-10 shrink-0 place-items-center rounded-2xl bg-white/10 text-[#f1bb5d]"><DashboardIcon name="sparkle" /></div><div><p className="text-sm font-bold">Kubera insights</p><p className="mt-1 text-xs leading-5 text-white/55">Smart shop insights will live here once enough inventory and sales history has been collected.</p><span className="mt-3 inline-block rounded-full bg-white/8 px-2.5 py-1 text-[10px] font-bold uppercase tracking-wider text-white/60">Coming later</span></div></div>
        </section>
      </main>
      <MobileBottomNav />
    </div>
  );
}
