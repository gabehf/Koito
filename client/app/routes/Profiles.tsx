import {
  Link,
  useLoaderData,
  useLocation,
  type LoaderFunctionArgs,
} from "react-router";
import type { ListeningProfile } from "api/api";
import ProfileAvatar from "~/components/ProfileAvatar";
import AppShell from "~/components/AppShell";
import { profilePath } from "../../api/profile-paths";

export async function clientLoader({ request }: LoaderFunctionArgs) {
  const response = await fetch("/apis/web/v1/profiles", {
    signal: request.signal,
  });
  if (!response.ok)
    throw new Response("Unable to load listening profiles", {
      status: response.status,
    });
  return { profiles: (await response.json()) as ListeningProfile[] };
}
export function meta() {
  return [{ title: "Listening profiles - Koito" }];
}
export default function Profiles() {
  const { profiles } = useLoaderData<typeof clientLoader>();
  const location = useLocation();
  const legacyPath = new URLSearchParams(location.search).get("next");
  // Only known legacy routes can be restored; never redirect to an external URL.
  const next =
    legacyPath &&
    /^\/(artist\/\d+|album\/\d+|track\/\d+|listens|rewind|chart\/top\/(artists|albums|tracks))(\?|$)/.test(
      legacyPath,
    )
      ? legacyPath
      : "/";
  return (
    <AppShell>
      <main className="w-full max-w-4xl mx-auto px-6 py-12 sm:py-20 min-h-[80vh]">
        <p className="text-sm uppercase tracking-widest text-(--color-fg-secondary)">
          Koito
        </p>
        <h1 className="mt-3">Choose a listening profile</h1>
        <p className="mt-3 mb-10 text-(--color-fg-secondary)">
          Explore a listening history. No login needed.
        </p>
        {profiles.length ? (
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
            {profiles.map((profile) => (
              <Link
                key={profile.id}
                to={profilePath(profile.slug, next)}
                className="card p-6 flex items-center gap-4 hover:bg-(--color-bg-tertiary) focus-visible:outline-2 focus-visible:outline-offset-4"
              >
                <ProfileAvatar profile={profile} size={56} />
                <div className="min-w-0">
                  <span className="font-semibold block truncate">
                    {profile.display_name}
                  </span>
                  <span className="text-sm text-(--color-fg-secondary)">
                    /{profile.slug}
                  </span>
                </div>
              </Link>
            ))}
          </div>
        ) : (
          <div className="card p-8">
            <h2>No listening profiles yet</h2>
            <p className="mt-3 text-(--color-fg-secondary)">
              An administrator can prepare a listening profile for this
              instance.
            </p>
          </div>
        )}
        <Link
          to="/admin"
          className="inline-block mt-10 text-sm underline text-(--color-fg-secondary)"
        >
          Administration
        </Link>
      </main>
    </AppShell>
  );
}
