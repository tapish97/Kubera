"use server";
import { revalidatePath } from "next/cache";
import { ApiError, apiFetch } from "@/lib/api";
export async function recordSale(formData: FormData): Promise<{ ok: true } | { ok: false; message: string }> { const text = (key: string) => String(formData.get(key) ?? ""); try { await apiFetch("/sales", { method: "POST", body: JSON.stringify({ items: [{ batch_id: text("batch_id"), quantity: Number(text("quantity")), selling_price_per_unit: Number(text("selling_price_per_unit")) }], notes: text("notes") }) }); const locale = text("locale"); revalidatePath(`/${locale}/dashboard`); revalidatePath(`/${locale}/sales/new`); return { ok: true }; } catch (error) { return { ok: false, message: error instanceof ApiError ? error.message : "Could not record sale" }; } }
