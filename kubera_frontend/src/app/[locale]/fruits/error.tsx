"use client";
import { RouteError } from "@/components/RouteError";
export default function Error({ retry }: { error: Error & { digest?: string }; retry: () => void }) { return <RouteError retry={retry} />; }
