"use client";
import { useState, useTransition } from "react";
import { useTranslations } from "next-intl";
import { settlePurchasePrice } from "@/features/settlements/actions";
import { formatDate, formatMoney, formatQuantity } from "@/lib/format";

export type UnpricedBatch = { batch_id:string; fruit:string; mark:string; quality:string|null; size:string; unit:string; quantity_received:number; quantity_remaining:number; quantity_sold:number; received_at:string; suggested_purchase_price_per_unit:number|null };

export function SettlementList({batches,locale,currency,timezone}:{batches:UnpricedBatch[];locale:string;currency:string;timezone:string}) {
  const t=useTranslations("Settlements");
  if(!batches.length) return <div className="rounded-[24px] border border-dashed border-[#d8d2c6] p-10 text-center"><p className="font-bold">{t("empty")}</p><p className="mt-1 text-sm text-[#747970]">{t("emptyHint")}</p></div>;
  return <div className="space-y-4">{batches.map(batch=><SettlementCard key={batch.batch_id} batch={batch} locale={locale} currency={currency} timezone={timezone}/>)}</div>;
}

function SettlementCard({batch,locale,currency,timezone}:{batch:UnpricedBatch;locale:string;currency:string;timezone:string}) {
  const t=useTranslations("Settlements"); const [pending,startTransition]=useTransition(); const [message,setMessage]=useState("");
  function submit(event:React.FormEvent<HTMLFormElement>) { event.preventDefault(); const form=new FormData(event.currentTarget); startTransition(async()=>{const result=await settlePurchasePrice(batch.batch_id,locale,form);setMessage(result.message);}); }
  return <article className="rounded-[24px] border border-[#e7e1d5] bg-white p-4"><div className="flex justify-between gap-3"><div><h2 className="font-bold">{batch.fruit} · {batch.mark}</h2><p className="mt-1 text-xs text-[#747970]">{t("added",{date:formatDate(batch.received_at,locale,{day:"numeric",month:"short",year:"numeric",timeZone:timezone})})} · {batch.quality||batch.size}</p></div><span className="rounded-full bg-[#fff1d9] px-3 py-1 text-xs font-bold text-[#7b5b27]">{t("pending")}</span></div><div className="mt-4 grid grid-cols-2 gap-2 text-sm"><div className="rounded-xl bg-[#f3f1eb] p-3"><small className="text-[#747970]">{t("sold")}</small><p className="font-bold">{formatQuantity(batch.quantity_sold,locale,batch.unit)}</p></div><div className="rounded-xl bg-[#f3f1eb] p-3"><small className="text-[#747970]">{t("remaining")}</small><p className="font-bold">{formatQuantity(batch.quantity_remaining,locale,batch.unit)}</p></div></div><form onSubmit={submit} className="mt-4"><label className="text-sm font-bold">{t("finalPrice",{unit:batch.unit})}<input name="purchase_price_per_unit" type="number" inputMode="decimal" min="0" step="0.01" required defaultValue={batch.suggested_purchase_price_per_unit?.toFixed(2)??""} className="mt-2 min-h-13 w-full rounded-xl border border-[#d8d2c6] px-4 text-lg font-bold outline-none focus:border-[#216148]"/></label>{batch.suggested_purchase_price_per_unit!=null&&<p className="mt-2 text-xs text-[#747970]">{t("suggested",{amount:formatMoney(batch.suggested_purchase_price_per_unit,locale,currency)})}</p>}{message&&<p className="mt-2 text-sm text-[#216148]">{message}</p>}<button disabled={pending} className="mt-4 min-h-12 w-full rounded-xl bg-[#216148] font-bold text-white disabled:opacity-60">{pending?t("saving"):t("confirm")}</button></form></article>;
}
