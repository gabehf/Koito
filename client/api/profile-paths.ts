// URL construction is explicit: a request never discovers its profile from the
// current window location, so old callbacks cannot write into a newly-selected profile.
export function profileApiPath(slug: string, path = ""): string {
  if (!slug) throw new Error("A listening profile is required");
  const suffix = path.replace(/^\/apis\/web\/v1\/?/, "").replace(/^\//, "");
  return `/apis/web/v1/profiles/${encodeURIComponent(slug)}${suffix ? `/${suffix}` : ""}`;
}

export function profilePath(slug: string, path = "/"): string {
  if (!slug || !path.startsWith("/") || path.startsWith("//")) return path;
  return `/${encodeURIComponent(slug)}${path === "/" ? "" : path}`;
}

export function profileRequestPath(
  slug: string | undefined,
  path: string,
  method = "GET",
): string {
  if (!path.startsWith("/apis/web/v1/")) throw new Error("Invalid API path");
  const suffix = path.slice("/apis/web/v1".length);
  if (["/login", "/logout", "/config", "/user"].includes(suffix.split("?")[0]))
    return path;
  const personal =
    method.toUpperCase() === "GET" ||
    /^\/(listens|export|user\/apikeys)([/?]|$)/.test(suffix);
  if (!personal) return path; // Shared catalog mutations remain administrator-only.
  if (!slug) throw new Error("Choose a listening profile first");
  return profileApiPath(slug, suffix);
}

export function switchedProfilePath(
  slug: string,
  pathname: string,
  search: string,
  currentSlug?: string,
): string {
  if (!currentSlug) return profilePath(slug);
  const prefix = `/${encodeURIComponent(currentSlug)}`;
  const suffix = pathname.startsWith(`${prefix}/`)
    ? pathname.slice(prefix.length)
    : "/";
  // Media IDs may not exist in the next person's history; preserve general screens only.
  const reusable =
    /^\/(listens|rewind|chart\/top\/(artists|albums|tracks))\/?$/.test(suffix);
  const params = new URLSearchParams(search);
  for (const key of ["artist_id", "album_id", "track_id", "page"])
    params.delete(key);
  const query = params.toString();
  return (
    profilePath(slug, reusable ? suffix : "/") + (query ? `?${query}` : "")
  );
}
