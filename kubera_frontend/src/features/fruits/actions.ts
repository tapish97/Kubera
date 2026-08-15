"use server";

import { revalidatePath } from "next/cache";
import { ApiError, apiFetch } from "@/lib/api";

export type Fruit = { id: string; name: string; default_unit: string; is_active: boolean };
export type FruitActionResult = { ok: true } | { ok: false; message: string };

function value(formData: FormData, key: string) { return String(formData.get(key) ?? ""); }
function failure(error: unknown): FruitActionResult { return { ok: false, message: error instanceof ApiError ? error.message : "Could not save fruit" }; }

export async function createFruit(formData: FormData): Promise<FruitActionResult> {
  try {
    await apiFetch<Fruit>("/fruits", { method: "POST", body: JSON.stringify({ name: value(formData, "name"), default_unit: value(formData, "default_unit") }) });
    revalidatePath(`/${value(formData, "locale")}/fruits`);
    return { ok: true };
  } catch (error) { return failure(error); }
}

export async function updateFruit(formData: FormData): Promise<FruitActionResult> {
  try {
    await apiFetch<Fruit>(`/fruits/${encodeURIComponent(value(formData, "id"))}`, { method: "PATCH", body: JSON.stringify({ name: value(formData, "name"), default_unit: value(formData, "default_unit") }) });
    revalidatePath(`/${value(formData, "locale")}/fruits`);
    return { ok: true };
  } catch (error) { return failure(error); }
}

export async function setFruitStatus(formData: FormData): Promise<FruitActionResult> {
  try {
    await apiFetch<Fruit>(`/fruits/${encodeURIComponent(value(formData, "id"))}/status`, { method: "PATCH", body: JSON.stringify({ is_active: value(formData, "is_active") === "true" }) });
    revalidatePath(`/${value(formData, "locale")}/fruits`);
    return { ok: true };
  } catch (error) { return failure(error); }
}
