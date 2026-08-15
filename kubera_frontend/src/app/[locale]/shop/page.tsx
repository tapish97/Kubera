import Link from "next/link";
import { redirect } from "next/navigation";
import { getTranslations } from "next-intl/server";
import { AppShell } from "@/components/AppShell";
import { LanguageSwitcher } from "@/components/LanguageSwitcher";
import { PageHeader } from "@/components/PageHeader";
import { ProfileMenu } from "@/components/ProfileMenu";
import { ShopSettingsForm } from "@/features/shop/ShopSettingsForm";
import { apiFetch } from "@/lib/api";
import { auth } from "@/lib/auth/server";
import { formatMoney } from "@/lib/format";

export const dynamic = "force-dynamic";
type Account = { profile_name:string; preferred_locale:string; shop_name:string; currency:string; timezone:string; location_label:string; latitude:number|null; longitude:number|null; onboarding_completed_at:string|null };
type DailyReport = { date:string; stock_batches_added:number; purchase_value:number; sale_count:number; sales_revenue:number; gross_profit:number; estimated_sale_item_count:number; unpriced_batch_count:number };
type Tab = "overview" | "reports" | "settings";

function todayIn(timezone:string) { const parts=new Intl.DateTimeFormat("en",{timeZone:timezone||"UTC",year:"numeric",month:"2-digit",day:"2-digit"}).formatToParts(new Date()); const value=(type:string)=>parts.find(p=>p.type===type)?.value??""; return `${value("year")}-${value("month")}-${value("day")}`; }

export default async function ShopPage({params,searchParams}:{params:Promise<{locale:string}>;searchParams:Promise<{tab?:string;date?:string;generate?:string}>}) {
  const {locale}=await params; const query=await searchParams; const t=await getTranslations({locale,namespace:"Shop"});
  const {data:session}=await auth.getSession(); if(!session?.user) redirect(`/${locale}/login`);
  const account=await apiFetch<Account>("/me"); if(!account.onboarding_completed_at) redirect(`/${locale}/onboarding`);
  const tab:Tab=query.tab==="reports"||query.tab==="settings"?query.tab:"overview";
  const date=typeof query.date==="string"&&/^\d{4}-\d{2}-\d{2}$/.test(query.date)?query.date:todayIn(account.timezone);
  const report=tab==="reports"&&query.generate==="1"?await apiFetch<DailyReport>(`/reports/daily?date=${encodeURIComponent(date)}`):null;
  const notification=await apiFetch<{unpriced_batch_count:number}>("/notifications/summary");
  return <AppShell><div className="px-5 pt-[max(1.25rem,env(safe-area-inset-top))]">
    <PageHeader eyebrow={t("eyebrow")} title={account.shop_name} actions={<><LanguageSwitcher/><div className="rounded-full bg-[#173f31] p-0.5"><ProfileMenu name={session.user.name??account.profile_name} email={session.user.email} shopName={account.shop_name}/></div></>}/>
    <nav className="mt-6 grid grid-cols-3 rounded-2xl bg-[#eee9df] p-1" aria-label={t("sections")}>{(["overview","reports","settings"] as Tab[]).map(item=><Link key={item} href={`/${locale}/shop?tab=${item}`} className={`rounded-xl px-2 py-2.5 text-center text-sm font-bold ${tab===item?"bg-white text-[#173f31] shadow-sm":"text-[#747970]"}`}>{t(item)}</Link>)}</nav>
    {tab==="overview"&&<section className="mt-5 space-y-3"><ActionLink href={`/${locale}/settlements`} title={t("buyingPrices")} hint={t("buyingPricesHint")} count={notification.unpriced_batch_count}/><ActionLink href={`/${locale}/suppliers`} title={t("marks")} hint={t("marksHint")}/><Link href={`/${locale}/shop?tab=reports`} className="flex min-h-16 items-center justify-between rounded-2xl border border-[#e7e1d5] bg-white px-4"><span><b>{t("dailyReports")}</b><small className="mt-1 block text-[#747970]">{t("dailyReportsHint")}</small></span><span>→</span></Link></section>}
    {tab==="reports"&&<section className="mt-5"><form className="flex items-end gap-2 rounded-2xl border border-[#e7e1d5] bg-white p-4"><input type="hidden" name="tab" value="reports"/><input type="hidden" name="generate" value="1"/><label className="min-w-0 flex-1 text-sm font-bold">{t("reportDate")}<input name="date" type="date" defaultValue={date} max={todayIn(account.timezone)} className="mt-2 min-h-12 w-full rounded-xl border border-[#ded9cf] px-3"/></label><button className="min-h-12 rounded-xl bg-[#216148] px-4 font-bold text-white">{t("generate")}</button></form>{!report&&<p className="mt-5 text-center text-sm text-[#747970]">{t("chooseDay")}</p>}{report&&<><div className="mt-4 grid grid-cols-2 gap-3 rounded-[24px] bg-[#173f31] p-4 text-white"><ReportTile label={t("sales")} value={formatMoney(report.sales_revenue,locale,account.currency)}/><ReportTile label={report.estimated_sale_item_count?t("estimatedProfit"):t("profit")} value={formatMoney(report.gross_profit,locale,account.currency)}/><ReportTile label={t("stockPurchased")} value={formatMoney(report.purchase_value,locale,account.currency)}/><ReportTile label={t("activity")} value={t("activityValue",{buys:report.stock_batches_added,sales:report.sale_count})} small/></div>{report.unpriced_batch_count>0&&<div className="mt-3 rounded-2xl border border-amber-300 bg-amber-50 p-4 text-sm text-amber-950"><b>{t("missingPrices",{count:report.unpriced_batch_count})}</b><p className="mt-1">{t("estimatedExplanation")}</p><Link href={`/${locale}/settlements`} className="mt-3 inline-block font-bold underline">{t("addPrices")}</Link></div>}</>}</section>}
    {tab==="settings"&&<section className="mt-5 rounded-[24px] border border-[#e7e1d5] bg-white p-5"><h2 className="mb-5 text-lg font-bold">{t("settingsTitle")}</h2><ShopSettingsForm mode="settings" profileName={account.profile_name} preferredLocale={account.preferred_locale} shopName={account.shop_name} currency={account.currency} timezone={account.timezone} locationLabel={account.location_label} latitude={account.latitude} longitude={account.longitude}/></section>}
  </div></AppShell>;
}

function ActionLink({href,title,hint,count=0}:{href:string;title:string;hint:string;count?:number}) { return <Link href={href} className="flex min-h-16 items-center justify-between rounded-2xl border border-[#e7e1d5] bg-white px-4"><span><b>{title}</b><small className="mt-1 block text-[#747970]">{hint}</small></span><span className="flex items-center gap-2">{count>0&&<span className="rounded-full bg-red-500 px-2 py-0.5 text-xs font-bold text-white">{count}</span>}→</span></Link>; }
function ReportTile({label,value,small=false}:{label:string;value:string;small?:boolean}) { return <div className="rounded-2xl bg-white/10 p-3"><p className="text-[11px] text-white/65">{label}</p><p className={`mt-1 font-bold ${small?"text-sm":"text-xl"}`}>{value}</p></div>; }
