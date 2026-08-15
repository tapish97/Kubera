"use client";

import { useState, useTransition } from "react";
import { LocationFields } from "@/components/LocationFields";
import { updateSupplier } from "@/features/suppliers/actions";

export type Supplier = { id: string; name: string; mark: string; phone: string; notes: string; location_label: string; latitude: number | null; longitude: number | null };

export function SupplierManager({ suppliers, locale }: { suppliers: Supplier[]; locale: string }) {
  if (!suppliers.length) return <div className="rounded-2xl border border-dashed border-[#d8d2c6] p-8 text-center"><p className="font-bold">No marks yet</p><p className="mt-1 text-sm text-[#747970]">Marks appear here after you add stock.</p></div>;
  return <div className="space-y-3">{suppliers.map((supplier) => <SupplierForm key={supplier.id} supplier={supplier} locale={locale} />)}</div>;
}

function SupplierForm({ supplier, locale }: { supplier: Supplier; locale: string }) {
  const [open, setOpen] = useState(false); const [pending, startTransition] = useTransition(); const [message, setMessage] = useState("");
  const input = "mt-2 min-h-12 w-full rounded-xl border border-[#ded9cf] bg-white px-3 outline-none focus:border-[#216148]";
  function submit(event: React.FormEvent<HTMLFormElement>) { event.preventDefault(); const form = new FormData(event.currentTarget); setMessage(""); startTransition(async () => { const result = await updateSupplier(supplier.id, locale, form); setMessage(result.message); if (result.ok) setOpen(false); }); }
  return <article className="overflow-hidden rounded-[22px] border border-[#e7e1d5] bg-white">
    <button type="button" onClick={() => setOpen(!open)} className="flex min-h-16 w-full items-center justify-between px-4 text-left"><span><strong className="block">{supplier.mark}</strong><small className="text-[#747970]">{supplier.name} · {supplier.location_label || "Location needed"}</small></span><span className="font-bold text-[#216148]">{open ? "Close" : "Edit"}</span></button>
    {open && <form onSubmit={submit} className="space-y-4 border-t border-[#eee9df] p-4">
      <label className="block text-sm font-bold">Mark<input required name="mark" defaultValue={supplier.mark} maxLength={80} className={input} /></label>
      <label className="block text-sm font-bold">Farmer / supplier name<input required name="name" defaultValue={supplier.name} maxLength={120} className={input} /></label>
      <label className="block text-sm font-bold">Phone<input name="phone" defaultValue={supplier.phone} inputMode="tel" className={input} /></label>
      <LocationFields required label="Farm / mark origin" hint="Enter a small locality and tap the exact farm or collection area on the map." currentLocationLabel="Use current location" locatingLabel="Finding location…" coordinatesLabel="Select a pin on the map" locationLabel={supplier.location_label} latitude={supplier.latitude} longitude={supplier.longitude} />
      <label className="block text-sm font-bold">Notes<textarea name="notes" defaultValue={supplier.notes} maxLength={1000} rows={4} placeholder="Farmer details, produce quality, payment terms…" className={`${input} py-3`} /></label>
      {message && <p role="status" className="text-sm text-[#216148]">{message}</p>}
      <button disabled={pending} className="min-h-13 w-full rounded-xl bg-[#216148] font-bold text-white disabled:opacity-60">{pending ? "Saving…" : "Save mark"}</button>
    </form>}
  </article>;
}
