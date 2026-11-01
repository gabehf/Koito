import { redirect, type LoaderFunctionArgs } from "react-router";
import type { ListeningProfile } from "api/api";
import { profilePath } from "../../api/profile-paths";
export async function clientLoader({ request }: LoaderFunctionArgs) {
  const url = new URL(request.url);
  const response = await fetch("/apis/web/v1/profiles", {
    signal: request.signal,
  });
  if (!response.ok)
    throw new Response("Unable to resolve listening profile", {
      status: response.status,
    });
  const profiles = (await response.json()) as ListeningProfile[];
  const destination = url.pathname + url.search;
  if (profiles.length === 1)
    return redirect(profilePath(profiles[0].slug, destination));
  return redirect(`/?next=${encodeURIComponent(destination)}`);
}
export default function LegacyRedirect() {
  return null;
}
