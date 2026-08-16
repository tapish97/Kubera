import type { ReactNode } from "react";
import { MobileBottomNav } from "@/components/MobileBottomNav";
import { NetworkStatus } from "@/components/NetworkStatus";
import { apiFetch } from "@/lib/api";
import { getLocale } from "next-intl/server";
import { redirect } from "next/navigation";

type NavigationData = { pendingPriceCount: number; preferredLocale?: string };

export async function AppShell({ children, navigationData }: { children: ReactNode; navigationData?: NavigationData }) {
	let pendingPriceCount = navigationData?.pendingPriceCount ?? 0;
	let preferredLocale = navigationData?.preferredLocale;
	if (!navigationData) {
		try { const [notification,account]=await Promise.all([apiFetch<{unpriced_batch_count:number}>("/notifications/summary"),apiFetch<{preferred_locale:string}>("/me")]);pendingPriceCount=notification.unpriced_batch_count;preferredLocale=account.preferred_locale; } catch {}
	}
  const currentLocale=await getLocale();
  if(preferredLocale&&preferredLocale!==currentLocale)redirect(`/${preferredLocale}/dashboard`);
  return <div className="min-h-screen bg-[#f5f1e8] text-[#20241f]"><NetworkStatus /><main className="mx-auto min-h-screen max-w-lg overflow-hidden bg-[#fffdf8] pb-28 shadow-[0_0_60px_rgba(44,53,43,0.08)]">{children}</main><MobileBottomNav pendingPriceCount={pendingPriceCount} /></div>;
}
