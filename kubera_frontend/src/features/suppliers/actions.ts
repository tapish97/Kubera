"use server";

import { revalidatePath } from "next/cache";
import { ApiError, apiFetch } from "@/lib/api";

const text = (form: FormData, key: string) => String(form.get(key) ?? "").trim();
const coordinate = (form: FormData, key: string) => text(form, key) ? Number(text(form, key)) : null;

export async function updateSupplier(id: string, locale: string, form: FormData): Promise<{ ok: boolean; message: string }> {
  try {
    await apiFetch(`/suppliers/${id}`, { method: "PATCH", body: JSON.stringify({ name: text(form, "name"), mark: text(form, "mark"), phone: text(form, "phone"), notes: text(form, "notes"), location_label: text(form, "location_label"), latitude: coordinate(form, "latitude"), longitude: coordinate(form, "longitude") }) });
    revalidatePath(`/${locale}/suppliers`);
    revalidatePath(`/${locale}/stock/add`);
    return { ok: true, message: "Saved" };
  } catch (error) {
    return { ok: false, message: error instanceof ApiError ? error.message : "Could not update mark" };
  }
}
