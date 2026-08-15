import { DashboardIcon } from "@/components/DashboardIcon";

const items = [
  { label: "Home", icon: "home" as const, active: true },
  { label: "Inventory", icon: "inventory" as const },
  { label: "Add", icon: "plus" as const, primary: true },
  { label: "Sales", icon: "sale" as const },
  { label: "More", icon: "more" as const },
];

export function MobileBottomNav() {
  return (
    <nav aria-label="Primary navigation" className="fixed inset-x-0 bottom-0 z-30 mx-auto max-w-lg border-t border-[#e7e1d5] bg-[#fffdf8]/95 px-3 pb-[max(0.75rem,env(safe-area-inset-bottom))] pt-2 backdrop-blur-xl">
      <div className="grid grid-cols-5 items-end">
        {items.map((item) => (
          <button key={item.label} type="button" aria-current={item.active ? "page" : undefined} aria-label={`${item.label}${item.active ? ", current page" : ", coming soon"}`} className={`flex min-h-12 flex-col items-center justify-center gap-1 text-[11px] font-semibold ${item.active ? "text-[#216148]" : "text-[#8b877e]"}`}>
            <span className={item.primary ? "-mt-7 grid h-14 w-14 place-items-center rounded-full bg-[#216148] text-white shadow-[0_10px_25px_rgba(33,97,72,0.28)] ring-4 ring-[#fffdf8]" : "grid h-7 place-items-center"}>
              <DashboardIcon name={item.icon} className={item.primary ? "h-7 w-7" : "h-5 w-5"} />
            </span>
            <span className={item.primary ? "mt-0.5 text-[#216148]" : ""}>{item.label}</span>
          </button>
        ))}
      </div>
    </nav>
  );
}
