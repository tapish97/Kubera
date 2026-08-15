"use client";

import { useEffect, useRef, useState } from "react";

type Position = { lat: number; lng: number };
type LeafletEvent = { latlng: Position; target: { getLatLng(): Position } };
type Layer = { addTo(map: MapInstance): Layer };
type Marker = Layer & { setLatLng(position: [number, number]): Marker; on(name: string, handler: (event: LeafletEvent) => void): Marker };
type MapInstance = { setView(position: [number, number], zoom: number): MapInstance; on(name: string, handler: (event: LeafletEvent) => void): MapInstance; panTo(position: [number, number]): void; invalidateSize(): void; remove(): void };
type LeafletAPI = {
  map(element: HTMLElement): MapInstance;
  tileLayer(url: string, options: { attribution: string; maxZoom: number }): Layer;
  marker(position: [number, number], options: { draggable: boolean; icon: unknown }): Marker;
  divIcon(options: { className: string; html: string; iconSize: [number, number]; iconAnchor: [number, number] }): unknown;
};

declare global { interface Window { L?: LeafletAPI } }
let loader: Promise<LeafletAPI> | null = null;

function loadLeaflet() {
  if (window.L) return Promise.resolve(window.L);
  if (loader) return loader;
  loader = new Promise((resolve, reject) => {
    if (!document.querySelector('link[data-kubera-leaflet]')) {
      const style = document.createElement("link"); style.rel = "stylesheet"; style.href = "https://unpkg.com/leaflet@1.9.4/dist/leaflet.css"; style.dataset.kuberaLeaflet = "true"; document.head.appendChild(style);
    }
    const script = document.createElement("script");
    script.src = "https://unpkg.com/leaflet@1.9.4/dist/leaflet.js"; script.async = true;
    script.onload = () => window.L ? resolve(window.L) : reject(new Error("Map did not load"));
    script.onerror = () => reject(new Error("Map did not load")); document.head.appendChild(script);
  });
  return loader;
}

export function OpenStreetMapPicker({ latitude, longitude, onChange }: { latitude: number | null; longitude: number | null; onChange(position: Position): void }) {
  const container = useRef<HTMLDivElement>(null); const map = useRef<MapInstance | null>(null); const marker = useRef<Marker | null>(null); const [error, setError] = useState("");
  const initialPosition = useRef<{ position: [number, number]; zoom: number }>({ position: latitude != null && longitude != null ? [latitude, longitude] : [19.076, 72.8777], zoom: latitude == null ? 10 : 15 });
  useEffect(() => {
    if (!container.current || map.current) return;
    let active = true;
    loadLeaflet().then((leaflet) => {
      if (!active || !container.current) return;
      const initial = initialPosition.current;
      const instance = leaflet.map(container.current).setView(initial.position, initial.zoom);
      leaflet.tileLayer("https://tile.openstreetmap.org/{z}/{x}/{y}.png", { maxZoom: 19, attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors' }).addTo(instance);
      const icon = leaflet.divIcon({ className: "kubera-map-pin", html: '<span style="display:block;width:24px;height:24px;border-radius:50% 50% 50% 0;background:#216148;border:3px solid white;box-shadow:0 2px 8px #0005;transform:rotate(-45deg)"></span>', iconSize: [30, 30], iconAnchor: [15, 28] });
      const pin = leaflet.marker(initial.position, { draggable: true, icon }).addTo(instance) as Marker;
      instance.on("click", (event) => { pin.setLatLng([event.latlng.lat, event.latlng.lng]); onChange(event.latlng); });
      pin.on("dragend", (event) => onChange(event.target.getLatLng()));
      map.current = instance; marker.current = pin; window.setTimeout(() => instance.invalidateSize(), 0);
    }).catch(() => setError("The map could not load. You can still use current location."));
    return () => { active = false; if (map.current) { map.current.remove(); map.current = null; marker.current = null; } };
  }, [onChange]);

  useEffect(() => { if (latitude == null || longitude == null || !map.current || !marker.current) return; marker.current.setLatLng([latitude, longitude]); map.current.panTo([latitude, longitude]); }, [latitude, longitude]);
  return <><div ref={container} className="h-64 w-full overflow-hidden rounded-2xl bg-[#e9ece8]" aria-label="Choose location on OpenStreetMap" />{error && <p className="mt-2 text-xs text-red-700">{error}</p>}</>;
}
