"use client";

import { useState } from "react";

type Props = {
  label: string;
  hint: string;
  currentLocationLabel: string;
  locatingLabel: string;
  coordinatesLabel: string;
  locationLabel?: string;
  latitude?: number | null;
  longitude?: number | null;
};

export function LocationFields(props: Props) {
  const [locationLabel, setLocationLabel] = useState(props.locationLabel ?? "");
  const [latitude, setLatitude] = useState(props.latitude?.toString() ?? "");
  const [longitude, setLongitude] = useState(props.longitude?.toString() ?? "");
  const [locating, setLocating] = useState(false);
  const [error, setError] = useState("");

  function locate() {
    if (!navigator.geolocation) { setError("Location is not available on this device."); return; }
    setLocating(true); setError("");
    navigator.geolocation.getCurrentPosition(
      ({ coords }) => { setLatitude(coords.latitude.toFixed(6)); setLongitude(coords.longitude.toFixed(6)); setLocating(false); },
      () => { setError("Could not get location. Allow location access and try again."); setLocating(false); },
      { enableHighAccuracy: true, timeout: 12_000 },
    );
  }

  return <div className="rounded-2xl border border-[#ded9cf] bg-[#faf8f2] p-4">
    <label className="block text-sm font-semibold text-[#3d433d]">{props.label}
      <input name="location_label" value={locationLabel} onChange={(event) => setLocationLabel(event.target.value)} maxLength={180} placeholder="Vashi, Navi Mumbai" className="mt-2 min-h-13 w-full rounded-2xl border border-[#ded9cf] bg-white px-4 text-base outline-none focus:border-[#216148]" />
    </label>
    <p className="mt-2 text-xs leading-5 text-[#747970]">{props.hint}</p>
    <button type="button" onClick={locate} disabled={locating} className="mt-3 min-h-11 rounded-xl bg-[#e5f1e9] px-4 text-sm font-bold text-[#216148] disabled:opacity-60">{locating ? props.locatingLabel : props.currentLocationLabel}</button>
    <details className="mt-3 text-sm"><summary className="cursor-pointer font-semibold text-[#59625b]">{props.coordinatesLabel}</summary>
      <div className="mt-2 grid grid-cols-2 gap-2"><input aria-label="Latitude" name="latitude" value={latitude} onChange={(event) => setLatitude(event.target.value)} type="number" min="-90" max="90" step="any" placeholder="Latitude" className="min-h-11 rounded-xl border border-[#ded9cf] bg-white px-3" /><input aria-label="Longitude" name="longitude" value={longitude} onChange={(event) => setLongitude(event.target.value)} type="number" min="-180" max="180" step="any" placeholder="Longitude" className="min-h-11 rounded-xl border border-[#ded9cf] bg-white px-3" /></div>
      {latitude && longitude && <a className="mt-2 inline-block font-bold text-[#216148] underline" href={`https://www.openstreetmap.org/?mlat=${latitude}&mlon=${longitude}#map=16/${latitude}/${longitude}`} target="_blank" rel="noreferrer">View pin on map</a>}
    </details>
    {error && <p role="alert" className="mt-2 text-xs text-red-700">{error}</p>}
  </div>;
}
