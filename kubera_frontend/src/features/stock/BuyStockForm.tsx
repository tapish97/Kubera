"use client";
import { useState, useTransition } from "react";
import { useRouter } from "next/navigation";
import { useTranslations } from "next-intl";
import { SearchSelect } from "@/components/SearchSelect";
import { StatusBanner } from "@/components/StatusBanner";
import { buyStock } from "@/features/stock/actions";
import { formatMoney } from "@/lib/format";
import { LocationFields } from "@/components/LocationFields";

type Supplier = { id: string; name: string; mark: string };
type Fruit = {
  id: string;
  name: string;
  default_unit: string;
  suppliers: Supplier[];
};
export function BuyStockForm({
  locale,
  currency = "INR",
  fruits,
}: {
  locale: string;
  currency?: string;
  fruits: Fruit[];
}) {
  const t = useTranslations("BuyStock");
  const router = useRouter();
  const [pending, startTransition] = useTransition();
  const [error, setError] = useState("");
  const [fruitName, setFruitName] = useState("");
  const selectedFruit = fruits.find(
    (fruit) =>
      fruit.name.toLocaleLowerCase() === fruitName.trim().toLocaleLowerCase(),
  );
  const [mark, setMark] = useState("");
  const selectedSupplier = selectedFruit?.suppliers.find(
    (supplier) =>
      supplier.mark.toLocaleLowerCase() === mark.trim().toLocaleLowerCase(),
  );
  const [unit, setUnit] = useState("box");
  const [quantity, setQuantity] = useState("");
  const [price, setPrice] = useState("");
  const total = Number(quantity) * Number(price);
  const input =
    "mt-2 min-h-14 w-full rounded-2xl border border-[#d8d2c6] bg-white px-4 text-base text-[#20241f] outline-none focus:border-[#216148]";
  function chooseFruit(value: string) {
    setFruitName(value);
    setMark("");
    const fruit = fruits.find(
      (item) =>
        item.name.toLocaleLowerCase() === value.trim().toLocaleLowerCase(),
    );
    if (fruit) setUnit(fruit.default_unit);
  }
  function submit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!fruitName.trim() || !mark.trim()) return;
    const form = new FormData(event.currentTarget);
    form.set("locale", locale);
    form.set("fruit_id", selectedFruit?.id ?? "");
    form.set("fruit_name", fruitName);
    form.set("supplier_id", selectedSupplier?.id ?? "");
    form.set("mark", mark);
    setError("");
    startTransition(async () => {
      const result = await buyStock(form);
      if (!result.ok) {
        setError(result.message);
        return;
      }
      router.replace(`/${locale}/inventory`);
      router.refresh();
    });
  }
  return (
    <form onSubmit={submit} className="space-y-5">
      <label className="block font-bold">
        1. {t("fruit")}
        <SearchSelect
          value={fruitName}
          onChange={chooseFruit}
          options={fruits.map((fruit) => fruit.name)}
          placeholder={t("fruitSearch")}
          label={t("fruit")}
        />
        <input type="hidden" name="fruit_name" value={fruitName} />
      </label>
      <label className="block font-bold">
        2. {t("supplierMark")}
        <SearchSelect
          value={mark}
          onChange={setMark}
          options={
            selectedFruit?.suppliers.map((supplier) => supplier.mark) ?? []
          }
          placeholder={selectedFruit ? t("markSearch") : t("markNew")}
          label={t("supplierMark")}
        />
        <input type="hidden" name="mark" value={mark} />
        {selectedFruit && selectedFruit.suppliers.length === 0 && (
          <p className="mt-2 text-xs text-[#777b73]">{t("noMarks")}</p>
        )}
      </label>
      {mark.trim() && !selectedSupplier && (
        <LocationFields
          required
          label={t("markLocation")}
          hint={t("markLocationHint")}
          currentLocationLabel={t("useCurrentLocation")}
          locatingLabel={t("locating")}
          coordinatesLabel={t("mapCoordinates")}
        />
      )}
      <div className="grid grid-cols-2 gap-3">
        <label className="text-sm font-bold">
          3. {t("quantity")}
          <input
            name="quantity"
            value={quantity}
            onChange={(e) => setQuantity(e.target.value)}
            inputMode="decimal"
            type="number"
            min="0.01"
            step="0.01"
            required
            className={input}
          />
        </label>
        <label className="text-sm font-bold">
          {t("unit")}
          <select
            name="unit"
            value={unit}
            onChange={(e) => setUnit(e.target.value)}
            className={input}
          >
            {["box", "kg", "piece", "crate", "dozen"].map((value) => (
              <option key={value} value={value}>
                {t(`units.${value}`)}
              </option>
            ))}
          </select>
        </label>
      </div>
      <label className="block font-bold">
        4. {t("price")}
        <input
          name="purchase_price_per_unit"
          value={price}
          onChange={(e) => setPrice(e.target.value)}
          inputMode="decimal"
          type="number"
          min="0"
          step="0.01"
          required
          className={input}
        />
      </label>
      <details className="rounded-2xl border border-[#e7e1d5] px-4 py-3">
        <summary className="cursor-pointer text-sm font-bold text-[#216148]">
          {t("optional")}
        </summary>
        <div className="mt-3 space-y-3">
          {!selectedSupplier && (
            <>
              <input
                name="supplier_name"
                placeholder={t("supplierName")}
                className={input}
              />
              <input
                name="phone"
                inputMode="tel"
                placeholder={t("phone")}
                className={input}
              />
            </>
          )}
          <select name="size" className={input}>
            <option value="normal">{t("normal")}</option>
            <option value="small">{t("small")}</option>
            <option value="large">{t("large")}</option>
          </select>
          <input
            name="quality"
            placeholder={t("qualityHint")}
            className={input}
          />
        </div>
      </details>
      <div className="rounded-[22px] bg-[#edf4ef] p-4">
        <p className="text-xs font-bold uppercase tracking-wider text-[#587264]">
          {t("total")}
        </p>
        <p className="mt-1 text-2xl font-bold text-[#173f31]">
          {formatMoney(Number.isFinite(total) ? total : 0, locale, currency)}
        </p>
      </div>
      {error && <StatusBanner tone="error" title={error} />}
      <button
        disabled={pending || !fruitName.trim() || !mark.trim()}
        className="min-h-14 w-full rounded-2xl bg-[#216148] text-lg font-bold text-white shadow-lg disabled:opacity-60"
      >
        {pending ? t("saving") : t("save")}
      </button>
    </form>
  );
}
