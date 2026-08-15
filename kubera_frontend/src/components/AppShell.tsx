import type { ReactNode } from "react";
import { MobileBottomNav } from "@/components/MobileBottomNav";
import { NetworkStatus } from "@/components/NetworkStatus";

export function AppShell({ children }: { children: ReactNode }) {
  return <div className="min-h-screen bg-[#f5f1e8] text-[#20241f]"><NetworkStatus /><main className="mx-auto min-h-screen max-w-lg overflow-hidden bg-[#fffdf8] pb-28 shadow-[0_0_60px_rgba(44,53,43,0.08)]">{children}</main><MobileBottomNav /></div>;
}
