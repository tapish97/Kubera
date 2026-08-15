"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { OpenStreetMapPicker } from "@/components/OpenStreetMapPicker";

type Props = { label: string; hint: string; currentLocationLabel: string; locatingLabel: string; coordinatesLabel: string; locationLabel?: string; latitude?: number | null; longitude?: number | null; required?: boolean };

type NominatimResult = { address?: Record<string, string>; display_name?: string };
const locationCache = new Map<string, string>();
let reverseQueue: Promise<void> = Promise.resolve();
let lastReverseLookup = 0;

function compactLocation(result: NominatimResult) {
  const address = result.address ?? {};
  const first = address.suburb || address.neighbourhood || address.village || address.town || address.city_district || address.city;
  const secondCandidates = [address.city, address.county, address.state_district, address.state];
  const second = secondCandidates.find((value) => value && value.toLocaleLowerCase() !== first?.toLocaleLowerCase());
  return [first, second].filter(Boolean).join(", ") || result.display_name?.split(",").slice(0, 2).join(",").trim() || "";
}

function reverseLocation(position: { lat: number; lng: number }) {
  const cacheKey = `${position.lat.toFixed(5)},${position.lng.toFixed(5)}`;
  const cached = locationCache.get(cacheKey);
  if (cached) return Promise.resolve(cached);
  let resolved = "";
  reverseQueue = reverseQueue.catch(() => undefined).then(async () => {
    const wait = Math.max(0, 1_050 - (Date.now() - lastReverseLookup));
    if (wait) await new Promise((resolve) => window.setTimeout(resolve, wait));
    lastReverseLookup = Date.now();
    const query = new URLSearchParams({ format: "jsonv2", lat: String(position.lat), lon: String(position.lng), zoom: "12", addressdetails: "1", "accept-language": navigator.language || "en" });
    const response = await fetch(`https://nominatim.openstreetmap.org/reverse?${query}`);
    if (!response.ok) throw new Error("Location lookup failed");
    resolved = compactLocation(await response.json() as NominatimResult);
    if (resolved) locationCache.set(cacheKey, resolved);
  });
  return reverseQueue.then(() => resolved);
}

export function LocationFields(props: Props) {
  const [locationLabel, setLocationLabel] = useState(props.locationLabel ?? "");
  const [latitude, setLatitude] = useState<number | null>(props.latitude ?? null);
  const [longitude, setLongitude] = useState<number | null>(props.longitude ?? null);
  const [locating, setLocating] = useState(false);
  const [lookingUp, setLookingUp] = useState(false);
  const [error, setError] = useState("");
  const lookupTimer = useRef<number | null>(null);
  const setPin = useCallback((position: { lat: number; lng: number }) => {
    setLatitude(Number(position.lat.toFixed(6))); setLongitude(Number(position.lng.toFixed(6)));
    if (lookupTimer.current) window.clearTimeout(lookupTimer.current);
    setLookingUp(true);
    lookupTimer.current = window.setTimeout(() => {
      reverseLocation(position).then((place) => { if (place) setLocationLabel(place); setError(""); }).catch(() => setError("Pin saved, but the place name could not be found. Enter it manually.")).finally(() => setLookingUp(false));
    }, 450);
  }, []);

  useEffect(() => () => { if (lookupTimer.current) window.clearTimeout(lookupTimer.current); }, []);

  function locate() {
    if (!navigator.geolocation) { setError("Location is not available on this device."); return; }
    setLocating(true); setError("");
    navigator.geolocation.getCurrentPosition(
      ({ coords }) => { setPin({ lat: coords.latitude, lng: coords.longitude }); setLocating(false); },
      () => { setError("Could not get location. Allow location access and try again."); setLocating(false); },
      { enableHighAccuracy: true, timeout: 12_000 },
    );
  }

  return <div className="rounded-2xl border border-[#ded9cf] bg-[#faf8f2] p-4">
    <label className="block text-sm font-semibold text-[#3d433d]">{props.label}<input name="location_label" value={locationLabel} onChange={(event) => setLocationLabel(event.target.value)} maxLength={180} required={props.required} placeholder="Vashi, Navi Mumbai" className="mt-2 min-h-13 w-full rounded-2xl border border-[#ded9cf] bg-white px-4 text-base outline-none focus:border-[#216148]" /></label>
    <p className="mt-2 text-xs leading-5 text-[#747970]">{props.hint}</p>
    <button type="button" onClick={locate} disabled={locating} className="my-3 min-h-11 rounded-xl bg-[#e5f1e9] px-4 text-sm font-bold text-[#216148] disabled:opacity-60">{locating ? props.locatingLabel : props.currentLocationLabel}</button>
    <OpenStreetMapPicker latitude={latitude} longitude={longitude} onChange={setPin} />
    <input type="hidden" name="latitude" value={latitude ?? ""} /><input type="hidden" name="longitude" value={longitude ?? ""} />
    <p className={`mt-3 text-xs font-semibold ${latitude == null ? "text-amber-700" : "text-[#216148]"}`}>{lookingUp ? "Finding place name…" : latitude == null ? props.coordinatesLabel : "✓ Map pin selected"}</p>
    {error && <p role="alert" className="mt-2 text-xs text-red-700">{error}</p>}
  </div>;
}
