"use server";
import { revalidatePath } from "next/cache";
import { ApiError, apiFetch } from "@/lib/api";

export async function updateBatch(batchID: string, locale: string, form: FormData): Promise<{ ok: boolean; message: string }> {
  const text = (key: string) => String(form.get(key) ?? "").trim();
  try {
    await apiFetch(`/inventory/batches/${batchID}`, { method: "PATCH", body: JSON.stringify({ quality: text("quality"), size: text("size"), quantity_remaining: Number(text("quantity_remaining")), purchase_price_per_unit: text("purchase_price_per_unit") === "" ? null : Number(text("purchase_price_per_unit")) }) });
    for (const path of ["inventory", "dashboard", "sales", "shop", "settlements"]) revalidatePath(`/${locale}/${path}`);
    return { ok: true, message: "Saved" };
  } catch (error) { return { ok: false, message: error instanceof ApiError ? error.message : "Could not update stock" }; }
}
