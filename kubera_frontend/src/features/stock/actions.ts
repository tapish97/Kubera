"use server";

import { revalidatePath } from "next/cache";
import { ApiError, apiFetch } from "@/lib/api";

export type StockActionResult = { ok: true } | { ok: false; message: string };
export async function buyStock(formData: FormData): Promise<StockActionResult> {
  const text = (key: string) => String(formData.get(key) ?? "");
  const number = (key: string) => { const raw = text(key); return raw === "" ? null : Number(raw); };
  try {
    await apiFetch("/inventory/batches/quick", { method: "POST", body: JSON.stringify({ fruit_id: text("fruit_id"), fruit_name: text("fruit_name"), supplier_id: text("supplier_id"), supplier_name: text("supplier_name"), mark: text("mark"), phone: text("phone"), quality: text("quality") || null, size: text("size"), quantity: number("quantity"), unit: text("unit"), purchase_price_per_unit: number("purchase_price_per_unit") }) });
    const locale = text("locale"); revalidatePath(`/${locale}/dashboard`); revalidatePath(`/${locale}/fruits`); revalidatePath(`/${locale}/stock/add`);
    return { ok: true };
  } catch (error) { return { ok: false, message: error instanceof ApiError ? error.message : "Could not save purchased stock" }; }
}
