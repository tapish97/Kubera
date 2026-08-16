"use server";
import { revalidatePath } from "next/cache";
import { ApiError,apiFetch } from "@/lib/api";
export async function closeDay(date:string,locale:string,allowEstimates:boolean){try{await apiFetch(`/closings/${date}`,{method:"POST",body:JSON.stringify({allow_estimates:allowEstimates})});for(const path of ["closing","shop","dashboard"])revalidatePath(`/${locale}/${path}`);return{ok:true,message:"Day closed"};}catch(error){return{ok:false,message:error instanceof ApiError?error.message:"Could not close day"};}}
