"use server";

import { ApiError, apiFetch } from "@/lib/api";

export type ShopActionResult =
  | { ok: true }
  | { ok: false; message: string };

function value(formData: FormData, key: string) {
  return String(formData.get(key) ?? "").trim();
}

function coordinate(formData: FormData, key: string) {
  const raw = value(formData, key);
  return raw === "" ? null : Number(raw);
}

export async function completeOnboarding(formData: FormData): Promise<ShopActionResult> {
  try {
    await apiFetch("/me/onboarding", {
      method: "POST",
      body: JSON.stringify({
        profile_name: value(formData, "profile_name"),
        shop_name: value(formData, "shop_name"),
        currency: value(formData, "currency"),
        timezone: value(formData, "timezone"),
        location_label: value(formData, "location_label"),
        latitude: coordinate(formData, "latitude"),
        longitude: coordinate(formData, "longitude"),
      }),
    });
    return { ok: true };
  } catch (error) {
    return {
      ok: false,
      message: error instanceof ApiError ? error.message : "Could not save your shop. Please try again.",
    };
  }
}

export async function updateSettings(formData: FormData): Promise<ShopActionResult> {
  try {
    await apiFetch("/me/profile", {
      method: "PATCH",
      body: JSON.stringify({ name: value(formData, "profile_name") }),
    });
    await apiFetch("/me/shop", {
      method: "PATCH",
      body: JSON.stringify({
        name: value(formData, "shop_name"),
        currency: value(formData, "currency"),
        timezone: value(formData, "timezone"),
        location_label: value(formData, "location_label"),
        latitude: coordinate(formData, "latitude"),
        longitude: coordinate(formData, "longitude"),
      }),
    });
    return { ok: true };
  } catch (error) {
    return {
      ok: false,
      message: error instanceof ApiError ? error.message : "Could not update settings. Please try again.",
    };
  }
}
