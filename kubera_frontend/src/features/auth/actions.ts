"use server";

import { auth } from "@/lib/auth/server";
import { redirect } from "next/navigation";

type AuthActionResult =
  | { ok: true }
  | { ok: false; message: string };

export async function signInWithEmail(formData: FormData): Promise<AuthActionResult> {
  try {
    const { error } = await auth.signIn.email({
      email: String(formData.get("email") ?? ""),
      password: String(formData.get("password") ?? ""),
    });
    if (error) {
      return { ok: false, message: error.message || "Login failed" };
    }
    return { ok: true };
  } catch {
    return { ok: false, message: "Could not connect. Please try again." };
  }
}

export async function signUpWithEmail(formData: FormData): Promise<AuthActionResult> {
  try {
    const { error } = await auth.signUp.email({
      name: String(formData.get("name") ?? ""),
      email: String(formData.get("email") ?? ""),
      password: String(formData.get("password") ?? ""),
    });
    if (error) {
      return { ok: false, message: error.message || "Registration failed" };
    }
    return { ok: true };
  } catch {
    return { ok: false, message: "Could not connect. Please try again." };
  }
}

export async function signOut(locale: string): Promise<AuthActionResult> {
  try {
    const { error } = await auth.signOut();
    if (error) {
      return { ok: false, message: error.message || "Could not log out" };
    }
  } catch {
    return { ok: false, message: "Could not connect. Please try again." };
  }

  const safeLocale = ["en", "hi", "mr"].includes(locale) ? locale : "en";
  redirect(`/${safeLocale}/login`);
}
