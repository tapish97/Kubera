"use client";

import { useMemo, useState, useTransition } from "react";
import { useRouter } from "next/navigation";
import { useTranslations } from "next-intl";
import { ConfirmSheet } from "@/components/ConfirmSheet";
import { DashboardIcon } from "@/components/DashboardIcon";
import { StatusBanner } from "@/components/StatusBanner";
import { createFruit, setFruitStatus, updateFruit, type Fruit } from "@/features/fruits/actions";

const units = ["box", "kg", "piece", "crate", "dozen"];

export function FruitManager({ fruits, locale }: { fruits: Fruit[]; locale: string }) {
  const t = useTranslations("Fruits"); const router = useRouter();
  const [query, setQuery] = useState(""); const [showInactive, setShowInactive] = useState(false);
  const [editing, setEditing] = useState<Fruit | "new" | null>(null); const [confirming, setConfirming] = useState<Fruit | null>(null);
  const [name, setName] = useState(""); const [unit, setUnit] = useState("box");
  const [error, setError] = useState(""); const [success, setSuccess] = useState(""); const [pending, startTransition] = useTransition();
  const visible = useMemo(() => fruits.filter((fruit) => (showInactive || fruit.is_active) && fruit.name.toLocaleLowerCase(locale).includes(query.trim().toLocaleLowerCase(locale))), [fruits, locale, query, showInactive]);

  function openEditor(fruit: Fruit | "new") { setEditing(fruit); setName(fruit === "new" ? "" : fruit.name); setUnit(fruit === "new" ? "box" : fruit.default_unit); setError(""); setSuccess(""); }
  function save() {
    const form = new FormData(); form.set("locale", locale); form.set("name", name); form.set("default_unit", unit);
    if (editing !== "new" && editing) form.set("id", editing.id);
    startTransition(async () => { const result = editing === "new" ? await createFruit(form) : await updateFruit(form); if (!result.ok) { setError(result.message); return; } setEditing(null); setSuccess(editing === "new" ? t("created") : t("updated")); router.refresh(); });
  }
  function changeStatus(fruit: Fruit) {
    const form = new FormData(); form.set("locale", locale); form.set("id", fruit.id); form.set("is_active", String(!fruit.is_active));
    startTransition(async () => { const result = await setFruitStatus(form); if (!result.ok) { setError(result.message); setConfirming(null); return; } setConfirming(null); setSuccess(fruit.is_active ? t("deactivated") : t("activated")); router.refresh(); });
  }

  return <>
    {success && <div className="mb-4"><StatusBanner tone="success" title={success} /></div>}
    {error && !editing && <div className="mb-4"><StatusBanner tone="error" title={error} /></div>}
    <div className="flex gap-3"><label className="relative flex-1"><span className="sr-only">{t("search")}</span><svg aria-hidden="true" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" className="absolute left-4 top-1/2 h-5 w-5 -translate-y-1/2 text-[#858980]"><circle cx="11" cy="11" r="7"/><path d="m16 16 4 4"/></svg><input value={query} onChange={(event) => setQuery(event.target.value)} placeholder={t("search")} className="min-h-12 w-full rounded-2xl border border-[#ded8cc] bg-white py-3 pl-11 pr-4 text-[#20241f] outline-none placeholder:text-[#8a8d84] focus:border-[#216148]" /></label><button type="button" onClick={() => openEditor("new")} className="grid h-12 w-12 shrink-0 place-items-center rounded-2xl bg-[#216148] text-white shadow-md" aria-label={t("add")}><DashboardIcon name="plus" /></button></div>
    <label className="mt-4 flex items-center justify-between rounded-2xl border border-[#e7e1d5] bg-white px-4 py-3 text-sm font-semibold"><span>{t("showInactive")}</span><input type="checkbox" checked={showInactive} onChange={(event) => setShowInactive(event.target.checked)} className="h-5 w-5 accent-[#216148]" /></label>
    <div className="mt-4 space-y-3">{visible.length === 0 ? <div className="rounded-[24px] border border-dashed border-[#d8d2c6] px-6 py-12 text-center"><div className="mx-auto grid h-12 w-12 place-items-center rounded-full bg-[#edf4ef] text-[#216148]"><DashboardIcon name="inventory" /></div><p className="mt-4 font-bold">{t(query ? "noResults" : "empty")}</p><p className="mt-1 text-sm text-[#777b73]">{t(query ? "noResultsHint" : "emptyHint")}</p></div> : visible.map((fruit) => <article key={fruit.id} className={`flex items-center gap-3 rounded-[22px] border bg-white p-4 ${fruit.is_active ? "border-[#e7e1d5]" : "border-[#e7e1d5] opacity-60"}`}><div className="grid h-11 w-11 shrink-0 place-items-center rounded-2xl bg-[#edf4ef] text-[#216148]"><DashboardIcon name="inventory" /></div><div className="min-w-0 flex-1"><h2 className="truncate font-bold">{fruit.name}</h2><p className="mt-0.5 text-xs text-[#777b73]">{t("unit", { unit: t(`units.${fruit.default_unit}`) })} · {fruit.is_active ? t("active") : t("inactive")}</p></div><button type="button" onClick={() => openEditor(fruit)} className="rounded-xl px-3 py-2 text-sm font-bold text-[#216148]">{t("edit")}</button><button type="button" onClick={() => fruit.is_active ? setConfirming(fruit) : changeStatus(fruit)} disabled={pending} className="rounded-xl px-2 py-2 text-xs font-bold text-[#7b5b27]">{fruit.is_active ? t("deactivate") : t("activate")}</button></article>)}</div>

    {editing && <div className="fixed inset-0 z-50 flex items-end justify-center sm:items-center"><button type="button" aria-label={t("close")} onClick={() => setEditing(null)} className="absolute inset-0 bg-[#101a15]/55 backdrop-blur-[2px]"/><section role="dialog" aria-modal="true" aria-labelledby="fruit-editor-title" className="relative w-full max-w-lg rounded-t-[28px] bg-[#fffdf8] px-5 pb-[max(1.5rem,env(safe-area-inset-bottom))] pt-5 shadow-2xl sm:rounded-[28px]"><h2 id="fruit-editor-title" className="text-xl font-bold">{editing === "new" ? t("addTitle") : t("editTitle")}</h2><label className="mt-5 block text-sm font-bold">{t("name")}<input autoFocus maxLength={80} value={name} onChange={(event) => setName(event.target.value)} className="mt-2 min-h-12 w-full rounded-2xl border border-[#d8d2c6] bg-white px-4 text-[#20241f] outline-none focus:border-[#216148]" /></label><label className="mt-4 block text-sm font-bold">{t("defaultUnit")}<select value={unit} onChange={(event) => setUnit(event.target.value)} className="mt-2 min-h-12 w-full rounded-2xl border border-[#d8d2c6] bg-white px-4 text-[#20241f] outline-none focus:border-[#216148]">{units.map((value) => <option key={value} value={value}>{t(`units.${value}`)}</option>)}</select></label>{error && <div className="mt-4"><StatusBanner tone="error" title={error} /></div>}<div className="mt-6 grid grid-cols-2 gap-3"><button type="button" onClick={() => setEditing(null)} className="min-h-12 rounded-2xl border border-[#ded8cc] bg-white font-bold text-[#5f645d]">{t("cancel")}</button><button type="button" onClick={save} disabled={pending || !name.trim()} className="min-h-12 rounded-2xl bg-[#216148] font-bold text-white disabled:opacity-60">{pending ? t("saving") : t("save")}</button></div></section></div>}
    <ConfirmSheet open={Boolean(confirming)} title={t("deactivateTitle")} description={t("deactivateDescription", { name: confirming?.name ?? "" })} confirmLabel={t("deactivate")} destructive busy={pending} onClose={() => setConfirming(null)} onConfirm={() => confirming && changeStatus(confirming)} />
  </>;
}
