"use server";
import { revalidatePath } from "next/cache";
import { ApiError, apiFetch } from "@/lib/api";

export async function settlePurchasePrice(batchID: string, locale: string, form: FormData): Promise<{ ok: boolean; message: string }> {
  const price = Number(form.get("purchase_price_per_unit"));
  try {
    await apiFetch(`/inventory/batches/${batchID}/purchase-price`, { method: "PATCH", body: JSON.stringify({ purchase_price_per_unit: price }) });
    for (const path of ["settlements", "inventory", "sales", "dashboard", "shop"]) revalidatePath(`/${locale}/${path}`);
    return { ok: true, message: "Buying price saved" };
  } catch (error) { return { ok: false, message: error instanceof ApiError ? error.message : "Could not save buying price" }; }
}
