"use client";

import { useRouter } from "next/navigation";

export default function LogoutButton() {
  const router = useRouter();

  async function logout() {
    await fetch("/api/auth/sign-out", {
      method: "POST",
    });

    router.push("/en/login");
    router.refresh();
  }

  return (
    <button
      onClick={logout}
      className="rounded-xl border px-4 py-2"
    >
      Logout
    </button>
  );
}