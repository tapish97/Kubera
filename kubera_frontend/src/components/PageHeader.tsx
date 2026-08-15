"use client";
import type { ReactNode } from "react";
import { useRouter } from "next/navigation";
import { useTranslations } from "next-intl";

export function PageHeader({ eyebrow, title, actions, back = false }: { eyebrow?: string; title: string; actions?: ReactNode; back?: boolean }) {
  const router = useRouter(); const t = useTranslations("Shell");
  return <header className="flex min-h-12 items-center justify-between gap-3"><div className="flex min-w-0 items-center gap-3">{back && <button type="button" onClick={() => router.back()} aria-label={t("back")} className="grid h-11 w-11 shrink-0 place-items-center rounded-full border border-[#e7e1d5] bg-white text-[#216148] active:scale-95"><svg aria-hidden="true" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" className="h-5 w-5"><path d="m15 18-6-6 6-6" /></svg></button>}<div className="min-w-0">{eyebrow && <p className="text-xs font-bold uppercase tracking-[0.18em] text-[#216148]">{eyebrow}</p>}<h1 className="truncate text-2xl font-bold tracking-tight">{title}</h1></div></div>{actions && <div className="flex shrink-0 items-center gap-2">{actions}</div>}</header>;
}
