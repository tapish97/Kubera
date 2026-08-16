"use client";
import {useTransition} from "react";
import {useLocale,useTranslations} from "next-intl";
import {usePathname,useRouter} from "next/navigation";
import {savePreferredLanguage} from "@/features/shop/language-actions";
const locales=[{code:"en",label:"EN"},{code:"hi",label:"हिं"},{code:"mr",label:"मर"}];
export function LanguageSwitcher({inverted=false}:{inverted?:boolean}){const locale=useLocale(),pathname=usePathname(),router=useRouter(),t=useTranslations("Common"),[pending,startTransition]=useTransition();function change(next:string){if(next===locale)return;const parts=pathname.split("/");parts[1]=next;startTransition(async()=>{const result=await savePreferredLanguage(next);if(result.ok)router.replace(parts.join("/")||`/${next}`);});}return <div aria-label={t("language")} className={`inline-flex rounded-full p-1 ${inverted?"bg-white/10":"bg-[#eee9df]"}`}>{locales.map(item=><button key={item.code} type="button" disabled={pending} onClick={()=>change(item.code)} aria-pressed={locale===item.code} className={`min-h-8 rounded-full px-2.5 text-[11px] font-bold transition disabled:opacity-60 ${locale===item.code?inverted?"bg-white text-[#173f31]":"bg-white text-[#216148] shadow-sm":inverted?"text-white/65":"text-[#777b73]"}`}>{item.label}</button>)}</div>}
