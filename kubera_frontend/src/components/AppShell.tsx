import type { ReactNode } from "react";
import { MobileBottomNav } from "@/components/MobileBottomNav";
import { NetworkStatus } from "@/components/NetworkStatus";
import { apiFetch } from "@/lib/api";

export async function AppShell({ children }: { children: ReactNode }) {
  let pendingPriceCount = 0;
  try { pendingPriceCount = (await apiFetch<{ unpriced_batch_count: number }>("/notifications/summary")).unpriced_batch_count; } catch {}
  return <div className="min-h-screen bg-[#f5f1e8] text-[#20241f]"><NetworkStatus /><main className="mx-auto min-h-screen max-w-lg overflow-hidden bg-[#fffdf8] pb-28 shadow-[0_0_60px_rgba(44,53,43,0.08)]">{children}</main><MobileBottomNav pendingPriceCount={pendingPriceCount} /></div>;
}
