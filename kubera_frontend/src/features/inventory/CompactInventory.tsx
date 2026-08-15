"use client";
import { useState, useTransition } from "react";
import { useTranslations } from "next-intl";
import { updateBatch } from "@/features/inventory/actions";
import { formatDate, formatMoney, formatQuantity } from "@/lib/format";

export type InventoryItem = { batch_id: string; fruit: string; supplier: string; mark: string; quality: string | null; size: string; quantity_remaining: number; unit: string; purchase_price_per_unit: number | null; received_at: string };
type Group = { fruit: string; unit: string; total: number; items: InventoryItem[] };

function groupsFor(items: InventoryItem[]) {
  const groups = new Map<string, Group>();
  for (const item of items) { const key = `${item.fruit.toLowerCase()}::${item.unit}`; const group = groups.get(key) ?? { fruit:item.fruit, unit:item.unit, total:0, items:[] }; group.total += Number(item.quantity_remaining); group.items.push(item); groups.set(key,group); }
  return [...groups.values()];
}

export function CompactInventory({ items, locale, currency, timezone }: { items: InventoryItem[]; locale: string; currency: string; timezone: string }) {
  const [editing, setEditing] = useState<string | null>(null); const groups = groupsFor(items);
  return <div className="space-y-3">{groups.map((group) => <section key={`${group.fruit}-${group.unit}`} className="overflow-hidden rounded-[18px] border border-[#e1e7e2] bg-white">
    <header className="flex items-center justify-between bg-[#edf4ef] px-3 py-2.5"><h2 className="font-bold text-[#173f31]">{group.fruit}</h2><span className="text-sm font-bold text-[#216148]">{formatQuantity(group.total,locale,group.unit)}</span></header>
    <div className="divide-y divide-[#eee9df]">{group.items.map((item) => <BatchRow key={item.batch_id} item={item} locale={locale} currency={currency} timezone={timezone} open={editing === item.batch_id} setOpen={(open) => setEditing(open ? item.batch_id : null)} />)}</div>
  </section>)}</div>;
}

function BatchRow({ item, locale, currency, timezone, open, setOpen }: { item: InventoryItem; locale:string; currency:string; timezone:string; open:boolean; setOpen(open:boolean):void }) {
  const t = useTranslations("Inventory");
  const [pending,startTransition] = useTransition(); const [message,setMessage] = useState("");
  function submit(event:React.FormEvent<HTMLFormElement>) { event.preventDefault(); const form = new FormData(event.currentTarget); startTransition(async()=>{ const result=await updateBatch(item.batch_id,locale,form); setMessage(result.message); if(result.ok) setOpen(false); }); }
  const details = [item.mark, item.quality || t("notSpecified"), item.size !== "normal" ? item.size : ""].filter(Boolean).join(" · ");
  return <div><div className="flex min-h-16 items-center gap-2 px-3 py-2">
    <button type="button" onClick={()=>setOpen(!open)} className="min-w-0 flex-1 text-left"><p className="truncate text-sm font-bold">{details}</p><p className="mt-0.5 truncate text-[11px] text-[#777b73]">{formatDate(item.received_at,locale,{day:"numeric",month:"short",hour:"numeric",minute:"2-digit",timeZone:timezone})}{item.purchase_price_per_unit != null ? ` · ${formatMoney(item.purchase_price_per_unit,locale,currency)}/${item.unit}` : ` · ${t("pricePending")}`}</p></button>
    <strong className="shrink-0 text-sm text-[#216148]">{formatQuantity(item.quantity_remaining,locale,item.unit)}</strong><button type="button" onClick={()=>setOpen(!open)} className="min-h-9 rounded-lg bg-[#f1eee7] px-2.5 text-xs font-bold text-[#59625b]">{open?t("close"):t("edit")}</button>
  </div>{open && <form onSubmit={submit} className="grid grid-cols-2 gap-2 border-t border-[#eee9df] bg-[#faf8f2] p-3">
    <label className="text-xs font-bold">{t("remaining")}<input name="quantity_remaining" type="number" inputMode="decimal" min="0" step="0.01" required defaultValue={item.quantity_remaining} className="mt-1 min-h-11 w-full rounded-lg border border-[#d8d2c6] px-3" /></label>
    <label className="text-xs font-bold">{t("buyPrice")}<input name="purchase_price_per_unit" type="number" inputMode="decimal" min="0" step="0.01" defaultValue={item.purchase_price_per_unit ?? ""} className="mt-1 min-h-11 w-full rounded-lg border border-[#d8d2c6] px-3" /></label>
    <label className="text-xs font-bold">{t("quality")}<input name="quality" defaultValue={item.quality ?? ""} className="mt-1 min-h-11 w-full rounded-lg border border-[#d8d2c6] px-3" /></label>
    <label className="text-xs font-bold">{t("size")}<select name="size" defaultValue={item.size} className="mt-1 min-h-11 w-full rounded-lg border border-[#d8d2c6] bg-white px-3"><option value="small">Small</option><option value="normal">Normal</option><option value="large">Large</option></select></label>
    {message && <p className="col-span-2 text-xs text-[#216148]">{message}</p>}<button disabled={pending} className="col-span-2 min-h-11 rounded-xl bg-[#216148] text-sm font-bold text-white disabled:opacity-60">{pending?t("saving"):t("save")}</button>
  </form>}</div>;
}
