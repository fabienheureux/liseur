/**
 * Miniflux Reader - Client-side utilities
 */

// Cookie helper functions
function setCookie(name, value, days = 365) {
  const date = new Date();
  date.setTime(date.getTime() + days * 24 * 60 * 60 * 1000);
  const expires = "expires=" + date.toUTCString();
  document.cookie = name + "=" + value + ";" + expires + ";path=/";
}

// Density selector
document.addEventListener("DOMContentLoaded", () => {
  const densitySelector = document.getElementById("density-selector");

  if (!densitySelector) return;

  const root = document.documentElement;

  // Update density on change
  densitySelector.addEventListener("change", (e) => {
    const density = e.target.value;
    root.style.setProperty("--density", density);
    setCookie("density", density);
  });
});
