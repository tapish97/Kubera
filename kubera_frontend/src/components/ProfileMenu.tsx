"use client";

import { useState } from "react";
import { useParams } from "next/navigation";
import { signOut } from "@/features/auth/actions";
import { useTranslations } from "next-intl";

export function ProfileMenu({
  name,
  email,
  shopName,
}: {
  name: string;
  email: string;
  shopName: string;
}) {
  const { locale } = useParams<{ locale: string }>();
  const t = useTranslations("Profile");
  const common = useTranslations("Common");
  const [open, setOpen] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const initial = name.trim().slice(0, 1).toUpperCase() || "K";

  async function logout() {
    setLoading(true);
    setError("");

    try {
      const result = await signOut(locale);
      if (!result.ok) {
        setError(result.message || t("logoutError"));
        return;
      }
    } catch {
      setError(t("logoutError"));
    } finally {
      setLoading(false);
    }
  }

  return (
    <>
      <button
        type="button"
        onClick={() => setOpen(true)}
        aria-label={t("open")}
        aria-expanded={open}
        className="grid h-11 w-11 place-items-center rounded-full bg-white/12 text-sm font-bold text-white ring-1 ring-white/20 transition active:scale-95"
      >
        {initial}
      </button>

      {open && (
        <div className="fixed inset-0 z-50 flex items-end justify-center sm:items-center">
          <button
            type="button"
            aria-label={t("close")}
            onClick={() => setOpen(false)}
            className="absolute inset-0 cursor-default bg-[#101a15]/55 backdrop-blur-[2px]"
          />

          <section
            role="dialog"
            aria-modal="true"
            aria-labelledby="profile-title"
            className="relative w-full max-w-lg rounded-t-[28px] bg-[#fffdf8] px-5 pb-[max(1.5rem,env(safe-area-inset-bottom))] pt-3 text-[#20241f] shadow-2xl sm:rounded-[28px]"
          >
            <div className="mx-auto mb-5 h-1.5 w-11 rounded-full bg-[#d9d4ca] sm:hidden" />

            <div className="flex items-center gap-3">
              <div className="grid h-12 w-12 place-items-center rounded-full bg-[#e5f1e9] text-base font-bold text-[#216148]">
                {initial}
              </div>
              <div className="min-w-0">
                <h2 id="profile-title" className="truncate text-base font-bold">
                  {name || t("user")}
                </h2>
                <p className="truncate text-xs text-[#777b73]">{email}</p>
              </div>
            </div>

            <div className="mt-5 rounded-2xl border border-[#e7e1d5] bg-white px-4 py-3">
              <p className="text-[10px] font-bold uppercase tracking-[0.16em] text-[#8a8d84]">
                {t("currentShop")}
              </p>
              <p className="mt-1 text-sm font-semibold">{shopName}</p>
            </div>

            {error && (
              <p role="alert" className="mt-4 rounded-xl bg-red-50 px-3 py-2 text-sm text-red-700">
                {error}
              </p>
            )}

            <button
              type="button"
              onClick={logout}
              disabled={loading}
              className="mt-5 min-h-12 w-full rounded-2xl border border-red-200 bg-red-50 px-4 font-bold text-red-700 disabled:cursor-not-allowed disabled:opacity-60"
            >
              {loading ? t("loggingOut") : t("logout")}
            </button>

            <button
              type="button"
              onClick={() => setOpen(false)}
              className="mt-2 min-h-11 w-full rounded-2xl px-4 text-sm font-semibold text-[#60655e]"
            >
              {common("cancel")}
            </button>
          </section>
        </div>
      )}
    </>
  );
}
