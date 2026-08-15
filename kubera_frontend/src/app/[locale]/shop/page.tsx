import Link from "next/link";
import { redirect } from "next/navigation";
import { AppShell } from "@/components/AppShell";
import { LanguageSwitcher } from "@/components/LanguageSwitcher";
import { PageHeader } from "@/components/PageHeader";
import { ProfileMenu } from "@/components/ProfileMenu";
import { ShopSettingsForm } from "@/features/shop/ShopSettingsForm";
import { apiFetch } from "@/lib/api";
import { auth } from "@/lib/auth/server";
import { formatMoney } from "@/lib/format";

export const dynamic = "force-dynamic";
type Account = { profile_name: string; shop_name: string; currency: string; timezone: string; location_label: string; latitude: number | null; longitude: number | null; onboarding_completed_at: string | null };
type DailyReport = { date: string; stock_batches_added: number; purchase_value: number; sale_count: number; sales_revenue: number; gross_profit: number };

function todayIn(timezone: string) {
  const parts = new Intl.DateTimeFormat("en", { timeZone: timezone || "UTC", year: "numeric", month: "2-digit", day: "2-digit" }).formatToParts(new Date());
  const value = (type: string) => parts.find((part) => part.type === type)?.value ?? "";
  return `${value("year")}-${value("month")}-${value("day")}`;
}

export default async function ShopPage({ params, searchParams }: { params: Promise<{ locale: string }>; searchParams: Promise<{ date?: string | string[] }> }) {
  const { locale } = await params; const query = await searchParams;
  const { data: session } = await auth.getSession(); if (!session?.user) redirect(`/${locale}/login`);
  const account = await apiFetch<Account>("/me"); if (!account.onboarding_completed_at) redirect(`/${locale}/onboarding`);
  const requestedDate = typeof query.date === "string" && /^\d{4}-\d{2}-\d{2}$/.test(query.date) ? query.date : todayIn(account.timezone);
  const report = await apiFetch<DailyReport>(`/reports/daily?date=${encodeURIComponent(requestedDate)}`);

  return <AppShell><div className="px-5 pt-[max(1.25rem,env(safe-area-inset-top))]">
    <PageHeader eyebrow="Your business" title={account.shop_name} actions={<><LanguageSwitcher /><div className="rounded-full bg-[#173f31] p-0.5"><ProfileMenu name={session.user.name ?? account.profile_name} email={session.user.email} shopName={account.shop_name} /></div></>} />

    <section className="mt-7 rounded-[26px] bg-[#173f31] p-5 text-white">
      <div className="flex items-end justify-between gap-3"><div><p className="text-xs font-bold uppercase tracking-wider text-[#b9d0c3]">Daily report</p><h2 className="mt-1 text-xl font-bold">Shop activity</h2></div><form><input aria-label="Report date" name="date" type="date" defaultValue={report.date} max={todayIn(account.timezone)} className="min-h-11 rounded-xl bg-white px-3 text-sm font-bold text-[#173f31]" /><button className="ml-2 min-h-11 rounded-xl bg-[#f1bb5d] px-3 text-sm font-bold text-[#4e350d]">View</button></form></div>
      <div className="mt-5 grid grid-cols-2 gap-3">
        <ReportTile label="Sales" value={formatMoney(report.sales_revenue, locale, account.currency)} />
        <ReportTile label="Profit" value={formatMoney(report.gross_profit, locale, account.currency)} />
        <ReportTile label="Stock purchased" value={formatMoney(report.purchase_value, locale, account.currency)} />
        <ReportTile label="Activity" value={`${report.stock_batches_added} buys · ${report.sale_count} sales`} small />
      </div>
    </section>

    <Link href={`/${locale}/suppliers`} className="mt-4 flex min-h-16 items-center justify-between rounded-[22px] border border-[#dce6df] bg-[#edf4ef] px-5 font-bold text-[#173f31]"><span>Marks & suppliers<small className="mt-1 block font-normal text-[#648072]">Edit farmer, phone, notes and origin</small></span><span>→</span></Link>
    <section className="mt-4 rounded-[26px] border border-[#e7e1d5] bg-white p-5 shadow-[0_12px_35px_rgba(44,53,43,0.06)]"><h2 className="mb-5 text-lg font-bold">Shop settings</h2><ShopSettingsForm mode="settings" profileName={account.profile_name} shopName={account.shop_name} currency={account.currency} timezone={account.timezone} locationLabel={account.location_label} latitude={account.latitude} longitude={account.longitude} /></section>
  </div></AppShell>;
}

function ReportTile({ label, value, small = false }: { label: string; value: string; small?: boolean }) {
  return <div className="rounded-2xl bg-white/10 p-3"><p className="text-[11px] font-semibold text-white/60">{label}</p><p className={`mt-1 font-bold ${small ? "text-sm" : "text-xl"}`}>{value}</p></div>;
}
