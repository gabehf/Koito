import {
  Outlet,
  useLoaderData,
  Link,
  isRouteErrorResponse,
  useRouteError,
  type LoaderFunctionArgs,
} from "react-router";
import type { ListeningProfile } from "api/api";
import { profileApiPath } from "../../api/profile-paths";
import { ProfileProvider } from "~/providers/ProfileProvider";
import AppShell from "~/components/AppShell";
import ProfileAvatar from "~/components/ProfileAvatar";
export async function clientLoader({ params, request }: LoaderFunctionArgs) {
  const response = await fetch(profileApiPath(params.slug!), {
    signal: request.signal,
  });
  if (!response.ok)
    throw new Response(
      response.status === 404
        ? "Listening profile not found"
        : "Unable to load profile",
      { status: response.status },
    );
  return (await response.json()) as ListeningProfile;
}
export default function ProfileLayout() {
  const profile = useLoaderData<typeof clientLoader>();
  return (
    <ProfileProvider key={`${profile.id}:${profile.slug}`} profile={profile}>
      <AppShell>
        <header className="w-full border-b border-(--color-border) px-5 sm:px-10 py-4 flex items-center gap-3">
          <ProfileAvatar profile={profile} size={32} />
          <div className="min-w-0">
            <span className="font-semibold block truncate">
              {profile.display_name}
            </span>
            <span className="text-xs text-(--color-fg-secondary) block truncate">
              Listening profile · /{profile.slug}
            </span>
          </div>
        </header>
        <Outlet />
      </AppShell>
    </ProfileProvider>
  );
}
export function ErrorBoundary() {
  const error = useRouteError();
  const missing = isRouteErrorResponse(error) && error.status === 404;
  return (
    <AppShell>
      <main className="mx-auto max-w-xl px-6 py-16">
        <h1>
          {missing
            ? "Profile or page not found"
            : "Unable to load this profile"}
        </h1>
        <p className="mt-4 text-(--color-fg-secondary)">
          {missing
            ? "This address may have changed, or this music is not in the selected profile’s history."
            : "Please try again, or choose another listening profile."}
        </p>
        <Link className="inline-block mt-6 underline" to="/">
          Choose a profile
        </Link>
      </main>
    </AppShell>
  );
}
