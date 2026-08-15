"use client";

import { useState } from "react";
import Link from "next/link";
import { useParams, useRouter } from "next/navigation";
import { signInWithEmail } from "@/features/auth/actions";
import { useTranslations } from "next-intl";
import { LanguageSwitcher } from "@/components/LanguageSwitcher";

export default function LoginPage() {
  const { locale } = useParams<{ locale: string }>();
  const router = useRouter();
  const t = useTranslations("Auth");

  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  async function handleLogin(e: React.FormEvent) {
    e.preventDefault();

    setLoading(true);
    setError("");

    try {
      const formData = new FormData();
      formData.set("email", email);
      formData.set("password", password);
      const result = await signInWithEmail(formData);
      if (!result.ok) {
        setError(result.message);
        return;
      }

      router.replace(`/${result.locale || locale}/dashboard`);
    } catch {
      setError(t("connectError"));
    } finally {
      setLoading(false);
    }
  }

  return (
    <main className="min-h-screen bg-[#f5f1e8] px-4 pb-10 pt-[max(1rem,env(safe-area-inset-top))] text-[#20241f]">
      <div className="mx-auto w-full max-w-sm">
        <div className="mb-4 flex justify-end"><LanguageSwitcher /></div>
        <section className="relative overflow-hidden rounded-[30px] bg-[#173f31] px-6 py-8 text-white shadow-[0_18px_45px_rgba(23,63,49,0.2)]">
          <div className="absolute -right-12 -top-16 h-40 w-40 rounded-full border-[28px] border-white/5" />
          <p className="relative text-xs font-bold uppercase tracking-[0.24em] text-[#f1bb5d]">Kubera</p>
          <h1 className="relative mt-3 text-3xl font-bold tracking-[-0.04em] text-white">{t("loginTitle")}</h1>
          <p className="relative mt-2 text-sm text-[#c7d9cf]">{t("loginSubtitle")}</p>
        </section>

        <form
          onSubmit={handleLogin}
          className="mt-5 space-y-4 rounded-[26px] border border-[#e3ddcf] bg-[#fffdf8] p-5 shadow-[0_12px_35px_rgba(44,53,43,0.07)]"
        >
          <input
            type="email"
            placeholder={t("email")}
            value={email}
            onChange={(e) =>
              setEmail(e.target.value)
            }
            className="w-full rounded-2xl border border-[#d8d2c6] bg-white p-4 text-[#20241f] outline-none placeholder:text-[#8a8d84] focus:border-[#216148] focus:ring-2 focus:ring-[#216148]/15"
            required
          />

          <input
            type="password"
            placeholder={t("password")}
            value={password}
            onChange={(e) =>
              setPassword(e.target.value)
            }
            className="w-full rounded-2xl border border-[#d8d2c6] bg-white p-4 text-[#20241f] outline-none placeholder:text-[#8a8d84] focus:border-[#216148] focus:ring-2 focus:ring-[#216148]/15"
            required
          />

          {error && (
            <p className="text-sm text-red-600">
              {error}
            </p>
          )}

          <button
            type="submit"
            disabled={loading}
            className="min-h-14 w-full rounded-2xl bg-[#216148] p-4 font-bold text-white shadow-[0_10px_24px_rgba(33,97,72,0.2)] disabled:cursor-not-allowed disabled:opacity-60"
          >
            {loading ? t("loggingIn") : t("login")}
          </button>
        </form>

        <p className="mt-6 text-center text-sm text-[#6f746d]">
          {t("newUser")} {" "}
          <Link href={`/${locale}/register`} className="font-semibold text-[#216148]">
            {t("createAccountLink")}
          </Link>
        </p>
      </div>
    </main>
  );
}
