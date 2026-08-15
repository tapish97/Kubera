"use client";

import { useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { completeOnboarding, updateSettings } from "@/features/shop/actions";
import { useTranslations } from "next-intl";
import { LocationFields } from "@/components/LocationFields";

type Props = {
  mode: "onboarding" | "settings";
  profileName: string;
  shopName: string;
  currency: string;
  timezone: string;
  locationLabel?: string;
  latitude?: number | null;
  longitude?: number | null;
  preferredLocale?: string;
};

export function ShopSettingsForm(props: Props) {
  const { locale } = useParams<{ locale: string }>();
  const router = useRouter();
  const t = useTranslations("Onboarding");
  const settings = useTranslations("Settings");
  const common = useTranslations("Common");
  const [profileName, setProfileName] = useState(props.profileName);
  const [shopName, setShopName] = useState(props.shopName);
  const [currency, setCurrency] = useState(props.currency || "INR");
  const [timezone, setTimezone] = useState(props.timezone || "Asia/Kolkata");
  const [preferredLocale, setPreferredLocale] = useState(props.preferredLocale || locale);
  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState("");
  const [error, setError] = useState("");

  async function submit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setLoading(true);
    setError("");
    setMessage("");

    const formData = new FormData();
    formData.set("profile_name", profileName);
    formData.set("shop_name", shopName);
    formData.set("currency", currency);
    formData.set("timezone", timezone);
    formData.set("preferred_locale", preferredLocale);
    const submitted = new FormData(event.currentTarget);
    formData.set("location_label", String(submitted.get("location_label") ?? ""));
    formData.set("latitude", String(submitted.get("latitude") ?? ""));
    formData.set("longitude", String(submitted.get("longitude") ?? ""));

    const result = props.mode === "onboarding"
      ? await completeOnboarding(formData)
      : await updateSettings(formData);

    if (!result.ok) {
      setError(result.message);
      setLoading(false);
      return;
    }

    if (props.mode === "onboarding") {
      router.replace(`/${preferredLocale}/dashboard`);
      return;
    }

    setMessage(settings("saved"));
    setLoading(false);
    if (preferredLocale !== locale) router.replace(`/${preferredLocale}/shop?tab=settings`);
    else router.refresh();
  }

  const inputClass = "mt-2 min-h-13 w-full rounded-2xl border border-[#ded9cf] bg-white px-4 text-base outline-none transition focus:border-[#216148] focus:ring-4 focus:ring-[#216148]/10";

  return (
    <form onSubmit={submit} className="space-y-5">
      <label className="block text-sm font-semibold text-[#3d433d]">
        {t("yourName")}
        <input className={inputClass} value={profileName} onChange={(event) => setProfileName(event.target.value)} minLength={2} maxLength={100} required autoComplete="name" />
      </label>

      <label className="block text-sm font-semibold text-[#3d433d]">
        {t("shopName")}
        <input className={inputClass} value={shopName} onChange={(event) => setShopName(event.target.value)} minLength={2} maxLength={120} required />
      </label>

      <div className="grid grid-cols-2 gap-3">
        <label className="block text-sm font-semibold text-[#3d433d]">
          {t("currency")}
          <select className={inputClass} value={currency} onChange={(event) => setCurrency(event.target.value)}>
            <option value="INR">INR · ₹</option>
            <option value="USD">USD · $</option>
            <option value="AED">AED</option>
          </select>
        </label>

        <label className="block text-sm font-semibold text-[#3d433d]">
          {t("timezone")}
          <select className={inputClass} value={timezone} onChange={(event) => setTimezone(event.target.value)}>
            <option value="Asia/Kolkata">{t("india")}</option>
            <option value="Asia/Dubai">{t("dubai")}</option>
            <option value="UTC">{t("utc")}</option>
          </select>
        </label>
      </div>

      <label className="block text-sm font-semibold text-[#3d433d]">
        {settings("defaultLanguage")}
        <select className={inputClass} value={preferredLocale} onChange={(event) => setPreferredLocale(event.target.value)}>
          <option value="en">English</option><option value="hi">हिन्दी</option><option value="mr">मराठी</option>
        </select>
        <span className="mt-1.5 block text-xs font-normal text-[#747970]">{settings("defaultLanguageHint")}</span>
      </label>

      <LocationFields label={t("shopLocation")} hint={t("shopLocationHint")} currentLocationLabel={t("useCurrentLocation")} locatingLabel={t("locating")} coordinatesLabel={t("mapCoordinates")} locationLabel={props.locationLabel} latitude={props.latitude} longitude={props.longitude} />

      {error && <p role="alert" className="rounded-2xl bg-red-50 px-4 py-3 text-sm text-red-700">{error}</p>}
      {message && <p role="status" className="rounded-2xl bg-[#e5f1e9] px-4 py-3 text-sm font-semibold text-[#216148]">{message}</p>}

      <button type="submit" disabled={loading} className="min-h-14 w-full rounded-2xl bg-[#216148] px-5 font-bold text-white shadow-[0_12px_28px_rgba(33,97,72,0.2)] disabled:cursor-not-allowed disabled:opacity-60">
        {loading ? common("saving") : props.mode === "onboarding" ? t("createShop") : common("saveChanges")}
      </button>
    </form>
  );
}
