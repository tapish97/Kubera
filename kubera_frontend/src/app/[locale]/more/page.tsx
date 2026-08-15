import { redirect } from "next/navigation";

export default async function LegacyMorePage({ params }: PageProps<"/[locale]/more">) {
  const { locale } = await params;
  redirect(`/${locale}/shop`);
}
