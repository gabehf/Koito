export function initWeekStartCookie() {
  if (typeof window === "undefined") return;

  // Unlike tz.ts's cookie, this one is recomputed on every load (not cached
  // once) so it stays in sync with the browser's live locale detection.
  let firstDay: number;
  try {
    // This doesn't work in Firefox
    // https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Intl/Locale/getWeekInfo
    firstDay = new Intl.Locale(navigator.language).getWeekInfo().firstDay;
  } catch (err) {
    return;
  }
  if (!firstDay) return;

  document.cookie = `week_start_locale=${firstDay}; Path=/; Max-Age=31536000; SameSite=Lax`;
}
