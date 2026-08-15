"use client";

import { useCallback, useState } from "react";
import { OpenStreetMapPicker } from "@/components/OpenStreetMapPicker";

type Props = { label: string; hint: string; currentLocationLabel: string; locatingLabel: string; coordinatesLabel: string; locationLabel?: string; latitude?: number | null; longitude?: number | null; required?: boolean };

export function LocationFields(props: Props) {
  const [locationLabel, setLocationLabel] = useState(props.locationLabel ?? "");
  const [latitude, setLatitude] = useState<number | null>(props.latitude ?? null);
  const [longitude, setLongitude] = useState<number | null>(props.longitude ?? null);
  const [locating, setLocating] = useState(false);
  const [error, setError] = useState("");
  const setPin = useCallback((position: { lat: number; lng: number }) => { setLatitude(Number(position.lat.toFixed(6))); setLongitude(Number(position.lng.toFixed(6))); }, []);

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
    <p className={`mt-3 text-xs font-semibold ${latitude == null ? "text-amber-700" : "text-[#216148]"}`}>{latitude == null ? props.coordinatesLabel : "✓ Map pin selected"}</p>
    {error && <p role="alert" className="mt-2 text-xs text-red-700">{error}</p>}
  </div>;
}
