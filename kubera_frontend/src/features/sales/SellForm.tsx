"use client";

import { useState, useTransition } from "react";
import { useRouter } from "next/navigation";
import { useTranslations } from "next-intl";
import { StatusBanner } from "@/components/StatusBanner";
import { recordSale } from "@/features/sales/actions";
import { formatDate, formatMoney } from "@/lib/format";

type Batch = { batch_id: string; fruit: string; mark: string; quality: string | null; size: string; quantity_remaining: number; unit: string; received_at: string };

export function SellForm({ locale, currency = "INR", timezone = "UTC", batches }: { locale: string; currency?: string; timezone?: string; batches: Batch[] }) {
  const t = useTranslations("Sell");
  const router = useRouter();
  const [error, setError] = useState("");
  const [pending, startTransition] = useTransition();
  const [batchID, setBatchID] = useState(batches[0]?.batch_id ?? "");
  const [quantity, setQuantity] = useState("");
  const [price, setPrice] = useState("");
  const selected = batches.find((batch) => batch.batch_id === batchID);
  const total = Number(quantity) * Number(price);
  const input = "mt-2 min-h-14 w-full rounded-2xl border border-[#d8d2c6] bg-white px-4 text-base text-[#20241f] outline-none focus:border-[#216148]";

  function submit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    form.set("locale", locale);
    setError("");
    startTransition(async () => {
      const result = await recordSale(form);
      if (!result.ok) { setError(result.message); return; }
      router.replace(`/${locale}/sales`);
      router.refresh();
    });
  }

  if (!batches.length) return <StatusBanner tone="warning" title={t("empty")}><p>{t("emptyHint")}</p></StatusBanner>;
  return <form onSubmit={submit} className="space-y-5">
    <label className="block font-bold">1. {t("batch")}
      <select name="batch_id" value={batchID} onChange={(event) => setBatchID(event.target.value)} className={input}>
        {batches.map((batch) => <option key={batch.batch_id} value={batch.batch_id}>{batch.fruit} · {batch.mark} · {batch.quality || batch.size} · {formatDate(batch.received_at, locale, { day: "numeric", month: "short", timeZone: timezone })} · {batch.quantity_remaining} {batch.unit}</option>)}
      </select>
    </label>
    <p className="rounded-2xl bg-[#edf4ef] px-4 py-3 text-sm font-semibold text-[#216148]">{t("available", { quantity: selected?.quantity_remaining ?? 0, unit: selected?.unit ?? "" })}</p>
    <label className="block font-bold">2. {t("quantity")}<input name="quantity" value={quantity} onChange={(event) => setQuantity(event.target.value)} type="number" inputMode="decimal" min="0.01" max={selected?.quantity_remaining} step="0.01" required className={input} /></label>
    <label className="block font-bold">3. {t("price")}<input name="selling_price_per_unit" value={price} onChange={(event) => setPrice(event.target.value)} type="number" inputMode="decimal" min="0" step="0.01" required className={input} /></label>
    <details className="rounded-2xl border border-[#e7e1d5] px-4 py-3"><summary className="cursor-pointer text-sm font-bold text-[#216148]">{t("notes")}</summary><textarea name="notes" rows={3} className={`${input} py-3`} /></details>
    <div className="rounded-[22px] bg-[#fff1d9] p-4"><p className="text-xs font-bold uppercase tracking-wider text-[#7b5b27]">{t("total")}</p><p className="mt-1 text-2xl font-bold text-[#4e350d]">{formatMoney(Number.isFinite(total) ? total : 0, locale, currency)}</p></div>
    {error && <StatusBanner tone="error" title={error} />}
    <button disabled={pending} className="min-h-14 w-full rounded-2xl bg-[#f1bb5d] text-lg font-bold text-[#4e350d] shadow-lg disabled:opacity-60">{pending ? t("saving") : t("save")}</button>
  </form>;
}
