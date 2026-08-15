"use client";
import { useEffect, useState } from "react";
import { useTranslations } from "next-intl";

export function NetworkStatus() {
  const t = useTranslations("Shell");
  const [online, setOnline] = useState(true);
  useEffect(() => { const update = () => setOnline(navigator.onLine); update(); window.addEventListener("online", update); window.addEventListener("offline", update); return () => { window.removeEventListener("online", update); window.removeEventListener("offline", update); }; }, []);
  if (online) return null;
  return <div role="status" aria-live="polite" className="fixed inset-x-0 top-0 z-[60] mx-auto max-w-lg bg-[#8b4b22] px-4 py-2 text-center text-xs font-bold text-white shadow-lg">{t("offline")}</div>;
}
