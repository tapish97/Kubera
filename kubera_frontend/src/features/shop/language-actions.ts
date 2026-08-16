"use server";
import {ApiError,apiFetch} from "@/lib/api";
export async function savePreferredLanguage(locale:string){if(!["en","hi","mr"].includes(locale))return{ok:false,message:"Invalid language"};try{const account=await apiFetch<{profile_name:string}>("/me");await apiFetch("/me/profile",{method:"PATCH",body:JSON.stringify({name:account.profile_name,preferred_locale:locale})});return{ok:true};}catch(error){return{ok:false,message:error instanceof ApiError?error.message:"Could not save language"};}}
