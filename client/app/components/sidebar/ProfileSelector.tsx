import { useEffect, useState } from "react";
import { Link, useLocation } from "react-router";
import { UsersRound, Check } from "lucide-react";
import type { ListeningProfile } from "api/api";
import { useOptionalProfile } from "~/providers/ProfileProvider";
import ProfileAvatar from "../ProfileAvatar";
import { Modal } from "../modals/Modal";
import { switchedProfilePath } from "../../../api/profile-paths";
export default function ProfileSelector() {
  const active = useOptionalProfile()?.profile;
  const location = useLocation();
  const [open, setOpen] = useState(false);
  const [profiles, setProfiles] = useState<ListeningProfile[]>([]);
  const [status, setStatus] = useState("Loading profiles…");
  useEffect(() => {
    if (!open) return;
    const controller = new AbortController();
    setStatus("Loading profiles…");
    fetch("/apis/web/v1/profiles", { signal: controller.signal })
      .then(async (response) => {
        if (!response.ok) throw new Error("Unable to load profiles");
        const profiles = (await response.json()) as ListeningProfile[];
        setProfiles(profiles);
        setStatus(profiles.length ? "" : "No listening profiles yet.");
      })
      .catch((error) => {
        if (error.name !== "AbortError")
          setStatus(
            "Unable to load profiles. Close this dialog and try again.",
          );
      });
    return () => controller.abort();
  }, [open]);
  return (
    <>
      <button
        type="button"
        aria-label={`Choose listening profile${active ? `: ${active.display_name}` : ""}`}
        title="Choose listening profile"
        onClick={() => setOpen(true)}
        className="p-2 rounded-md hover:bg-(--color-bg-tertiary) inline-flex items-center justify-center focus-visible:outline-2"
      >
        {active ? (
          <ProfileAvatar profile={active} size={18} />
        ) : (
          <UsersRound size={18} />
        )}
      </button>
      <Modal
        isOpen={open}
        onClose={() => setOpen(false)}
        maxW={480}
        label="Choose a listening profile"
      >
        <h2 className="mb-5">Listening profiles</h2>
        {status ? (
          <p role="status" className="text-(--color-fg-secondary)">
            {status}
          </p>
        ) : (
          <div className="flex flex-col gap-2">
            {profiles.map((profile) => (
              <Link
                key={profile.id}
                to={switchedProfilePath(
                  profile.slug,
                  location.pathname,
                  location.search,
                  active?.slug,
                )}
                onClick={() => setOpen(false)}
                aria-current={profile.id === active?.id ? "true" : undefined}
                className="flex items-center gap-3 p-3 rounded-md hover:bg-(--color-bg-tertiary) focus-visible:outline-2"
              >
                <ProfileAvatar profile={profile} size={40} />
                <span className="flex-1 min-w-0">
                  <span className="block truncate font-semibold">
                    {profile.display_name}
                  </span>
                  <span className="block text-sm text-(--color-fg-secondary)">
                    /{profile.slug}
                  </span>
                </span>
                {profile.id === active?.id && <Check size={18} />}
              </Link>
            ))}
          </div>
        )}
        <Link
          to="/"
          onClick={() => setOpen(false)}
          className="inline-block mt-6 text-sm underline"
        >
          All profiles
        </Link>
      </Modal>
    </>
  );
}
