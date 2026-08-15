import { redirect } from "next/navigation";
import { ShopSettingsForm } from "@/features/shop/ShopSettingsForm";
import { apiFetch } from "@/lib/api";
import { auth } from "@/lib/auth/server";
import { getTranslations } from "next-intl/server";
import { LanguageSwitcher } from "@/components/LanguageSwitcher";

export const dynamic = "force-dynamic";

type Account = {
  profile_name: string;
  shop_name: string;
  currency: string;
  timezone: string;
  onboarding_completed_at: string | null;
};

export default async function OnboardingPage({ params }: PageProps<"/[locale]/onboarding">) {
  const { locale } = await params;
  const t = await getTranslations({ locale, namespace: "Onboarding" });
  const { data: session } = await auth.getSession();
  if (!session?.user) redirect(`/${locale}/login`);

  const account = await apiFetch<Account>("/me");
  if (account.onboarding_completed_at) redirect(`/${locale}/dashboard`);

  return (
    <main className="min-h-screen bg-[#f5f1e8] px-5 pb-10 pt-[max(2rem,env(safe-area-inset-top))] text-[#20241f]">
      <div className="mx-auto max-w-md">
        <div className="flex items-center justify-between"><p className="text-xs font-bold uppercase tracking-[0.2em] text-[#216148]">{t("eyebrow")}</p><LanguageSwitcher /></div>
        <h1 className="mt-3 text-3xl font-bold tracking-[-0.04em]">{t("title")}</h1>
        <p className="mt-2 text-sm leading-6 text-[#6f746d]">{t("subtitle")}</p>

        <div className="mt-8 rounded-[26px] border border-[#e7e1d5] bg-[#fffdf8] p-5 shadow-[0_16px_45px_rgba(44,53,43,0.08)]">
          <ShopSettingsForm
            mode="onboarding"
            profileName={account.profile_name || session.user.name || ""}
            shopName={account.shop_name}
            currency={account.currency}
            timezone={account.timezone}
          />
        </div>
      </div>
    </main>
  );
}
