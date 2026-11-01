import { useEffect, useState, type FormEvent } from "react";
import { Link, useSearchParams } from "react-router";
import { adminProfiles } from "api/admin";
import type { ApiKey, ListeningProfile } from "api/api";
import ProfileAvatar from "./ProfileAvatar";
import ProfileRelayEditor from "./ProfileRelayEditor";

const field = "w-full min-w-0 fg bg rounded-md border p-3";
const button =
  "large-button rounded-md px-4 py-2 disabled:opacity-50 disabled:cursor-not-allowed";
const section = "border rounded-lg p-4 sm:p-6 flex flex-col gap-4 min-w-0";
function message(error: unknown) {
  return error instanceof Error ? error.message : "Request failed";
}
export function suggestSlug(name: string) {
  return name
    .normalize("NFD")
    .replace(/[\u0300-\u036f]/g, "")
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-|-$/g, "")
    .slice(0, 64)
    .replace(/-$/, "");
}
function IconSelect({
  value,
  onChange,
}: {
  value: string;
  onChange: (value: string) => void;
}) {
  return (
    <label className="flex flex-col gap-2">
      Fallback icon
      <select
        className={field}
        value={value}
        onChange={(e) => onChange(e.target.value)}
      >
        <option value="">Name initials</option>
        {["music", "headphones", "disc", "radio", "user"].map((icon) => (
          <option key={icon} value={icon}>
            {icon}
          </option>
        ))}
      </select>
    </label>
  );
}
function PhotoInput({ onChange }: { onChange: (file?: File) => void }) {
  return (
    <label className="flex flex-col gap-2">
      Photo (optional, up to 5 MiB)
      <input
        className={field}
        type="file"
        accept="image/*"
        onChange={(e) => onChange(e.target.files?.[0])}
      />
    </label>
  );
}
function IdentityFields({
  name,
  slug,
  setName,
  setSlug,
}: {
  name: string;
  slug: string;
  setName: (value: string) => void;
  setSlug: (value: string) => void;
}) {
  return (
    <>
      <label className="flex flex-col gap-2">
        Display name
        <input
          className={field}
          required
          maxLength={200}
          value={name}
          onChange={(e) => setName(e.target.value)}
        />
      </label>
      <label className="flex flex-col gap-2">
        URL name (slug)
        <input
          className={field}
          required
          pattern="[a-z0-9]+(-[a-z0-9]+)*"
          maxLength={64}
          value={slug}
          onChange={(e) => setSlug(e.target.value)}
        />
        <span className="text-sm">
          Lowercase letters, numbers and hyphens. Technical route names are
          reserved.
        </span>
      </label>
    </>
  );
}
export default function AdminProfiles({
  embedded = false,
  onNavigate,
}: {
  embedded?: boolean;
  onNavigate?: () => void;
} = {}) {
  const [profiles, setProfiles] = useState<ListeningProfile[]>([]);
  const [params, setParams] = useSearchParams();
  const [localProfileID, setLocalProfileID] = useState<number>();
  const selectProfile = (id?: number) => {
    if (embedded) setLocalProfileID(id);
    else setParams(id ? { profile: String(id) } : {});
  };
  const [loaded, setLoaded] = useState(false);
  const [error, setError] = useState("");
  const [name, setName] = useState("");
  const [slug, setSlug] = useState("");
  const [slugEdited, setSlugEdited] = useState(false);
  const [icon, setIcon] = useState("");
  const [photo, setPhoto] = useState<File>();
  const [busy, setBusy] = useState(false);
  const [notice, setNotice] = useState("");
  const selected = profiles.find((p) =>
    embedded ? p.id === localProfileID : String(p.id) === params.get("profile"),
  );
  useEffect(() => {
    const controller = new AbortController();
    adminProfiles
      .list(controller.signal)
      .then((data) => {
        setProfiles(data);
        setLoaded(true);
      })
      .catch((err) => {
        if (!controller.signal.aborted) setError(message(err));
      });
    return () => controller.abort();
  }, []);
  const choose = (id?: number) => {
    setNotice("");
    setPhoto(undefined);
    selectProfile(id);
  };
  const update = (p: ListeningProfile) =>
    setProfiles((items) => items.map((item) => (item.id === p.id ? p : item)));
  async function create(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setBusy(true);
    setError("");
    setNotice("");
    try {
      let profile = await adminProfiles.create(name.trim(), slug, icon);
      setProfiles((items) => [...items, profile]);
      selectProfile(profile.id);
      setName("");
      setSlug("");
      setSlugEdited(false);
      setNotice(
        "Profile created with its initial API key. Copy the key below to configure your player.",
      );
      if (photo) {
        try {
          profile = await adminProfiles.avatar(profile.slug, icon, photo);
          update(profile);
        } catch (err) {
          setNotice(
            `Profile and key created. Photo upload failed: ${message(err)}. Retry in Avatar.`,
          );
        }
      }
      setPhoto(undefined);
      setIcon("");
    } catch (err) {
      setError(message(err));
    } finally {
      setBusy(false);
    }
  }
  return (
    <section className="flex flex-col gap-6 mt-10">
      <h2 className="text-2xl">Listening profiles</h2>
      <p>
        Public profiles do not have passwords. API keys send scrobbles only to
        their linked profile.
      </p>
      {error && (
        <p role="alert" className="error">
          {error}
        </p>
      )}
      {notice && <p role="status">{notice}</p>}
      {!loaded && !error && <p>Loading profiles…</p>}
      {loaded && (
        <>
          <label className="flex flex-col gap-2">
            Manage profile
            <select
              disabled={busy}
              className={field}
              value={selected?.id ?? ""}
              onChange={(e) => choose(Number(e.target.value) || undefined)}
            >
              <option value="">Create a profile</option>
              {profiles.map((p) => (
                <option key={p.id} value={p.id}>
                  {p.display_name} (/{p.slug})
                </option>
              ))}
            </select>
          </label>
          {selected ? (
            <ProfileEditor
              key={`${selected.id}:${selected.slug}`}
              profile={selected}
              onNavigate={onNavigate}
              busy={busy}
              setBusy={setBusy}
              onUpdate={update}
              onDelete={() => {
                setProfiles((items) =>
                  items.filter((p) => p.id !== selected.id),
                );
                choose();
              }}
            />
          ) : (
            <form className={section} onSubmit={create}>
              <h3 className="text-xl">Create profile</h3>
              {profiles.length === 0 && (
                <p>No profiles yet. Create one to begin listening.</p>
              )}
              <IdentityFields
                name={name}
                slug={slug}
                setName={(value) => {
                  setName(value);
                  if (!slugEdited) setSlug(suggestSlug(value));
                }}
                setSlug={(value) => {
                  setSlugEdited(true);
                  setSlug(value);
                }}
              />
              <IconSelect value={icon} onChange={setIcon} />
              <PhotoInput key={profiles.length} onChange={setPhoto} />
              <p>
                One API key will be created automatically. No listener login is
                needed.
              </p>
              <button className={button} disabled={busy}>
                {busy ? "Creating…" : "Create profile"}
              </button>
            </form>
          )}
        </>
      )}
    </section>
  );
}
function ProfileEditor({
  profile,
  onNavigate,
  onUpdate,
  onDelete,
  busy,
  setBusy,
}: {
  onNavigate?: () => void;
  busy: boolean;
  setBusy: (value: boolean) => void;
  profile: ListeningProfile;
  onUpdate: (p: ListeningProfile) => void;
  onDelete: () => void;
}) {
  const [name, setName] = useState(profile.display_name);
  const [slug, setSlug] = useState(profile.slug);
  const [icon, setIcon] = useState(profile.avatar_icon || "");
  const [photo, setPhoto] = useState<File>();
  const [keys, setKeys] = useState<ApiKey[]>([]);
  const [keysLoaded, setKeysLoaded] = useState(false);
  const [label, setLabel] = useState("");
  const [error, setError] = useState("");
  const [notice, setNotice] = useState("");
  const [deleteCount, setDeleteCount] = useState<number>();
  const [confirmName, setConfirmName] = useState("");
  const [format, setFormat] = useState("koito");
  const [file, setFile] = useState<File>();
  const [importConfirmed, setImportConfirmed] = useState(false);
  useEffect(() => {
    const controller = new AbortController();
    adminProfiles
      .keys(profile.slug, controller.signal)
      .then((data) => {
        setKeys(data);
        setKeysLoaded(true);
      })
      .catch((err) => {
        if (!controller.signal.aborted) setError(message(err));
      });
    return () => controller.abort();
  }, [profile.slug]);
  async function run(action: () => Promise<void>) {
    setBusy(true);
    setError("");
    setNotice("");
    try {
      await action();
    } catch (err) {
      setError(message(err));
    } finally {
      setBusy(false);
    }
  }
  const endpoint =
    typeof window === "undefined"
      ? "/apis/listenbrainz/1/submit-listens"
      : `${window.location.origin}/apis/listenbrainz/1/submit-listens`;
  return (
    <div className="flex flex-col gap-6">
      <div className="flex gap-3 items-center min-w-0">
        <ProfileAvatar key={profile.avatar_image_id} profile={profile} />
        <Link
          className="underline break-all"
          to={`/${profile.slug}`}
          onClick={onNavigate}
        >
          {profile.display_name} — /{profile.slug}
        </Link>
      </div>
      {error && (
        <p role="alert" className="error break-words">
          {error}
        </p>
      )}
      {notice && <p role="status">{notice}</p>}
      <fieldset disabled={busy} className="flex flex-col gap-6 min-w-0">
        <form
          className={section}
          onSubmit={(e) => {
            e.preventDefault();
            void run(async () => {
              onUpdate(
                await adminProfiles.update(profile.slug, name.trim(), slug),
              );
              setNotice("Profile saved.");
            });
          }}
        >
          <h3 className="text-xl">Profile details</h3>
          <IdentityFields
            name={name}
            slug={slug}
            setName={setName}
            setSlug={setSlug}
          />
          <p>
            Changing the URL name keeps the history and API keys. The previous
            URL will return 404.
          </p>
          <button className={button}>Save profile</button>
        </form>
        <form
          className={section}
          onSubmit={(e) => {
            e.preventDefault();
            void run(async () => {
              onUpdate(await adminProfiles.avatar(profile.slug, icon, photo));
              setPhoto(undefined);
              setNotice("Avatar saved.");
            });
          }}
        >
          <h3 className="text-xl">Avatar</h3>
          <IconSelect value={icon} onChange={setIcon} />
          <PhotoInput
            key={profile.avatar_image_id ?? "none"}
            onChange={setPhoto}
          />
          <p>
            Photos take priority over icons. Saving without a photo replaces the
            current photo with the chosen icon or initials.
          </p>
          <div className="flex flex-wrap gap-3">
            <button className={button}>Save avatar</button>
            <button
              type="button"
              className={button}
              onClick={() =>
                void run(async () => {
                  onUpdate(await adminProfiles.clearAvatar(profile.slug));
                  setIcon("");
                  setPhoto(undefined);
                  setNotice("Avatar removed; using initials.");
                })
              }
            >
              Use initials
            </button>
          </div>
        </form>
        <ProfileRelayEditor slug={profile.slug} busy={busy} setBusy={setBusy} />
        <section className={section}>
          <h3 className="text-xl">API keys for {profile.display_name}</h3>
          <p>
            Revoke a key to stop future submissions. Existing history is
            preserved.
          </p>
          <CopyValue label="Scrobbling endpoint" value={endpoint} />
          <p className="text-sm">
            For a ListenBrainz server/base URL setting, use{" "}
            <code className="break-all">
              {endpoint.replace(/\/submit-listens$/, "")}
            </code>
            . Requests use <code>Authorization: Token YOUR_KEY</code>.
          </p>
          {!keysLoaded && <p>Loading keys…</p>}
          {keysLoaded && keys.length === 0 && (
            <p>No active keys. Create one to receive scrobbles.</p>
          )}
          {keys.map((key) => (
            <KeyRow
              key={key.id}
              item={key}
              disabled={busy}
              onRename={(value) =>
                run(async () => {
                  await adminProfiles.labelKey(profile.slug, key.id, value);
                  setKeys((items) =>
                    items.map((item) =>
                      item.id === key.id ? { ...item, label: value } : item,
                    ),
                  );
                })
              }
              onRevoke={() =>
                run(async () => {
                  await adminProfiles.revokeKey(profile.slug, key.id);
                  setKeys((items) =>
                    items.filter((item) => item.id !== key.id),
                  );
                  setNotice("Key revoked. History preserved.");
                })
              }
            />
          ))}
          <form
            className="flex flex-col sm:flex-row gap-3"
            onSubmit={(e) => {
              e.preventDefault();
              void run(async () => {
                const key = await adminProfiles.createKey(
                  profile.slug,
                  label.trim(),
                );
                setKeys((items) => [...items, key]);
                setLabel("");
              });
            }}
          >
            <label className="flex-1 min-w-0">
              New key label
              <input
                className={field}
                required
                maxLength={100}
                placeholder="Navidrome or Phone"
                value={label}
                onChange={(e) => setLabel(e.target.value)}
              />
            </label>
            <button className={button}>Create key</button>
          </form>
        </section>
        <form
          className={section}
          onSubmit={(e) => {
            e.preventDefault();
            if (!file) return;
            void run(async () => {
              await adminProfiles.importFile(profile.slug, format, file);
              setImportConfirmed(false);
              setNotice(`Import completed for ${profile.display_name}.`);
            });
          }}
        >
          <h3 className="text-xl">Import history</h3>
          <p>
            Destination: <strong>{profile.display_name}</strong> (/
            {profile.slug}, ID {profile.id}). To use another destination, choose
            it in “Manage profile”. Names inside the file do not change the
            destination.
          </p>
          <label className="flex flex-col gap-2">
            Format
            <select
              className={field}
              value={format}
              onChange={(e) => {
                setFormat(e.target.value);
                setImportConfirmed(false);
              }}
            >
              {["koito", "listenbrainz", "spotify", "lastfm", "maloja"].map(
                (v) => (
                  <option key={v} value={v}>
                    {v}
                  </option>
                ),
              )}
            </select>
          </label>
          <label className="flex flex-col gap-2">
            History file (up to 100 MiB; ListenBrainz uses a ZIP export)
            <input
              className={field}
              type="file"
              required
              accept={format === "listenbrainz" ? ".zip" : ".json"}
              onChange={(e) => {
                setFile(e.target.files?.[0]);
                setImportConfirmed(false);
              }}
            />
          </label>
          <label className="flex gap-3 items-start">
            <input
              type="checkbox"
              checked={importConfirmed}
              onChange={(e) => setImportConfirmed(e.target.checked)}
            />
            I confirm this history belongs to {profile.display_name}.
          </label>
          <p>
            Keep this page open until completion. Large imports can take time.
            If a file fails midway, some records may already have been saved.
          </p>
          <button className={button} disabled={!importConfirmed || !file}>
            {busy ? "Working…" : `Import into ${profile.display_name}`}
          </button>
          <a
            className="underline"
            href={`/apis/web/v1/profiles/${encodeURIComponent(profile.slug)}/export`}
            download
          >
            Export this profile’s history
          </a>
        </form>
        <section className={section}>
          <h3 className="text-xl">Delete profile</h3>
          <p>
            Deletes this profile, its keys, listens and avatar. Shared artists,
            albums and tracks remain. The administrator account remains
            available even if you delete the last profile.
          </p>
          {deleteCount === undefined ? (
            <button
              className={button}
              onClick={() =>
                void run(async () => {
                  const details = await adminProfiles.details(profile.slug);
                  setDeleteCount(details.listen_count);
                  setConfirmName("");
                })
              }
            >
              Review deletion
            </button>
          ) : (
            <>
              <p>
                Delete <strong>{profile.display_name}</strong> and{" "}
                <strong>{deleteCount}</strong> listens permanently?
              </p>
              <label className="flex flex-col gap-2">
                Type the display name to confirm
                <input
                  className={field}
                  autoComplete="off"
                  value={confirmName}
                  onChange={(e) => setConfirmName(e.target.value)}
                />
              </label>
              <div className="flex flex-wrap gap-3">
                <button
                  className={`${button} text-(--color-error)`}
                  disabled={confirmName !== profile.display_name}
                  onClick={() =>
                    void run(async () => {
                      try {
                        await adminProfiles.remove(
                          profile.slug,
                          confirmName,
                          deleteCount,
                        );
                        onDelete();
                      } catch (err) {
                        setDeleteCount(undefined);
                        setConfirmName("");
                        throw err;
                      }
                    })
                  }
                >
                  Permanently delete profile
                </button>
                <button
                  className={button}
                  onClick={() => setDeleteCount(undefined)}
                >
                  Cancel
                </button>
              </div>
            </>
          )}
        </section>
      </fieldset>
    </div>
  );
}
function CopyValue({
  label,
  value,
  secret = false,
}: {
  label: string;
  value: string;
  secret?: boolean;
}) {
  const [revealed, setRevealed] = useState(!secret);
  const [status, setStatus] = useState("");
  return (
    <div className="flex flex-col gap-2 min-w-0">
      <span>{label}</span>
      <code className="bg rounded p-3 break-all select-all">
        {revealed ? value : "••••••••••••••••"}
      </code>
      <div className="flex flex-wrap gap-2">
        {secret && (
          <button
            type="button"
            className={button}
            onClick={() => setRevealed(!revealed)}
          >
            {revealed ? "Hide key" : "Reveal key"}
          </button>
        )}
        <button
          type="button"
          className={button}
          onClick={async () => {
            try {
              await navigator.clipboard.writeText(value);
              setStatus("Copied.");
            } catch {
              setRevealed(true);
              setStatus("Select the displayed value and copy it manually.");
            }
          }}
        >
          Copy {secret ? "key" : "endpoint"}
        </button>
        <span role="status">{status}</span>
      </div>
    </div>
  );
}
function KeyRow({
  item,
  onRename,
  onRevoke,
  disabled,
}: {
  item: ApiKey;
  onRename: (label: string) => Promise<void>;
  onRevoke: () => Promise<void>;
  disabled: boolean;
}) {
  const [label, setLabel] = useState(item.label);
  const [confirm, setConfirm] = useState(false);
  return (
    <div className="border rounded-md p-3 flex flex-col gap-3 min-w-0">
      <form
        className="flex flex-col sm:flex-row gap-2"
        onSubmit={(e) => {
          e.preventDefault();
          void onRename(label.trim());
        }}
      >
        <label className="flex-1 min-w-0">
          Key label
          <input
            className={field}
            required
            maxLength={100}
            value={label}
            onChange={(e) => setLabel(e.target.value)}
          />
        </label>
        <button className={button} disabled={disabled || label === item.label}>
          Save label
        </button>
      </form>
      <CopyValue
        label={item.label || `Key ${item.id}`}
        value={item.key}
        secret
      />
      {confirm ? (
        <div className="flex flex-wrap gap-3 items-center">
          <span>Revoke “{item.label}”? Player submissions will stop.</span>
          <button
            type="button"
            className={button}
            onClick={() => void onRevoke()}
          >
            Confirm revocation
          </button>
          <button
            type="button"
            className={button}
            onClick={() => setConfirm(false)}
          >
            Cancel
          </button>
        </div>
      ) : (
        <button
          type="button"
          className={`${button} self-start`}
          onClick={() => setConfirm(true)}
        >
          Revoke key
        </button>
      )}
    </div>
  );
}
