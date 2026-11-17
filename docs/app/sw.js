"use strict";

// The phone recorder's service worker: it keeps every file of the page, so the
// recorder opens with the network off (FR-078). build-recorder.ps1 fills in the
// tokens: the cache is named from VERSION, the list is every file it shipped.
const CACHE = "symdiary-recorder-1.4.0";
const FILES = ["./", "assets/application-icon-1EQn0oGa.png", "assets/recorder-BMIb2EeW.css", "assets/recorder-COJPPb0W.js", "icon-192.png", "icon-512.png", "index.html", "manifest.webmanifest", "symdiary-wasm.js", "wasm_exec.js"];

self.addEventListener("install", (event) => {
  // cache: "reload" skips the HTTP cache, so a new version never keeps old files.
  const fresh = FILES.map((url) => new Request(url, { cache: "reload" }));
  event.waitUntil(caches.open(CACHE).then((cache) => cache.addAll(fresh)));
  self.skipWaiting();
});

self.addEventListener("activate", (event) => {
  event.waitUntil(
    caches
      .keys()
      .then((keys) => Promise.all(keys.filter((k) => k !== CACHE).map((k) => caches.delete(k))))
      .then(() => self.clients.claim())
  );
});

self.addEventListener("fetch", (event) => {
  event.respondWith(
    caches.match(event.request, { ignoreSearch: true }).then((hit) => hit || fetch(event.request))
  );
});
