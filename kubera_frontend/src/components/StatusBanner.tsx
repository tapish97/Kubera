import type { ReactNode } from "react";
export function StatusBanner({ tone, title, children }: { tone: "success" | "error" | "warning"; title: string; children?: ReactNode }) {
  const styles = { success: "border-[#bad7c6] bg-[#edf7f0] text-[#216148]", error: "border-red-200 bg-red-50 text-red-700", warning: "border-[#ecd7ae] bg-[#fff8e8] text-[#75551c]" }[tone];
  return <div role={tone === "error" ? "alert" : "status"} className={`rounded-2xl border px-4 py-3 text-sm ${styles}`}><p className="font-bold">{title}</p>{children && <div className="mt-1 text-xs leading-5 opacity-85">{children}</div>}</div>;
}
