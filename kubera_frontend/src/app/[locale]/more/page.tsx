import { redirect } from "next/navigation";
import { AppShell } from "@/components/AppShell";
import { PageHeader } from "@/components/PageHeader";
import { ProfileMenu } from "@/components/ProfileMenu";
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

export default async function MorePage({ params }: PageProps<"/[locale]/more">) {
  const { locale } = await params;
  const t = await getTranslations({ locale, namespace: "Settings" });
  const { data: session } = await auth.getSession();
  if (!session?.user) redirect(`/${locale}/login`);

  const account = await apiFetch<Account>("/me");
  if (!account.onboarding_completed_at) redirect(`/${locale}/onboarding`);

  return (
    <AppShell>
      <div className="px-5 pt-[max(1.25rem,env(safe-area-inset-top))]">
        <PageHeader eyebrow={t("eyebrow")} title={t("title")} actions={<><LanguageSwitcher /><div className="rounded-full bg-[#173f31] p-0.5"><ProfileMenu name={session.user.name ?? account.profile_name} email={session.user.email} shopName={account.shop_name} /></div></>} />

        <section className="mt-7 rounded-[26px] border border-[#e7e1d5] bg-white p-5 shadow-[0_12px_35px_rgba(44,53,43,0.06)]">
          <ShopSettingsForm mode="settings" profileName={account.profile_name} shopName={account.shop_name} currency={account.currency} timezone={account.timezone} />
        </section>
      </div>
    </AppShell>
  );
}
