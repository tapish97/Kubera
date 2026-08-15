import { redirect } from "next/navigation";
import { AppShell } from "@/components/AppShell";
import { PageHeader } from "@/components/PageHeader";
import { SupplierManager, type Supplier } from "@/features/suppliers/SupplierManager";
import { apiFetch } from "@/lib/api";
import { auth } from "@/lib/auth/server";

export const dynamic = "force-dynamic";
export default async function SuppliersPage({ params }: PageProps<"/[locale]/suppliers">) {
  const { locale } = await params; const { data: session } = await auth.getSession(); if (!session?.user) redirect(`/${locale}/login`);
  const suppliers = await apiFetch<Supplier[]>("/suppliers");
  return <AppShell><div className="px-5 pt-[max(1.25rem,env(safe-area-inset-top))]"><PageHeader title="Marks & suppliers" back /><p className="mt-2 text-sm text-[#6f746d]">Farmer contacts, notes, and exact produce origins.</p><section className="mt-6"><SupplierManager suppliers={suppliers} locale={locale} /></section></div></AppShell>;
}
