import { useEffect, useState } from "react";
import { adminProfiles, type ProfileRelay } from "api/admin";

const field = "w-full min-w-0 fg bg rounded-md border p-3";
const button =
  "large-button rounded-md px-4 py-2 disabled:opacity-50 disabled:cursor-not-allowed";
const empty: ProfileRelay = { enabled: false, url: "", token: "" };

export default function ProfileRelayEditor({
  slug,
  busy,
  setBusy,
}: {
  slug: string;
  busy: boolean;
  setBusy: (value: boolean) => void;
}) {
  const [saved, setSaved] = useState<ProfileRelay>(empty);
  const [draft, setDraft] = useState<ProfileRelay>(empty);
  const [loaded, setLoaded] = useState(false);
  const [error, setError] = useState("");
  const [notice, setNotice] = useState("");
  useEffect(() => {
    const controller = new AbortController();
    adminProfiles
      .relay(slug, controller.signal)
      .then((data) => {
        setSaved(data);
        setDraft(data);
        setLoaded(true);
      })
      .catch((err) => {
        if (!controller.signal.aborted)
          setError(
            err instanceof Error ? err.message : "Could not load relay.",
          );
      });
    return () => controller.abort();
  }, [slug]);
  const dirty =
    draft.enabled !== saved.enabled ||
    (draft.enabled && (draft.url !== saved.url || draft.token !== saved.token));
  function change(value: Partial<ProfileRelay>) {
    setDraft((current) => ({ ...current, ...value }));
    setNotice("");
    setError("");
  }
  async function run(action: () => Promise<void>) {
    setBusy(true);
    setError("");
    setNotice("");
    try {
      await action();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Request failed.");
    } finally {
      setBusy(false);
    }
  }
  return (
    <form
      className="border rounded-lg p-4 sm:p-6 flex flex-col gap-4 min-w-0"
      onSubmit={(e) => {
        e.preventDefault();
        void run(async () => {
          if (draft.enabled) {
            const next = {
              ...draft,
              url: draft.url.trim().replace(/\/+$/, ""),
              token: draft.token.trim(),
            };
            await adminProfiles.saveRelay(slug, next.url, next.token);
            setSaved(next);
            setDraft(next);
          } else {
            await adminProfiles.removeRelay(slug);
            setSaved(empty);
            setDraft(empty);
          }
          setNotice("Relay configuration saved.");
        });
      }}
    >
      <h3 className="text-xl">ListenBrainz forwarding</h3>
      <p>
        Forward new player submissions from this profile to its external
        account. Local history is preserved.
      </p>
      {error && (
        <p role="alert" className="error break-words">
          {error}
        </p>
      )}
      {notice && <p role="status">{notice}</p>}
      {!loaded && !error && <p>Loading relay…</p>}
      {loaded && (
        <>
          <p role="status">
            {saved.enabled
              ? "Saved configuration: forwarding enabled."
              : "Saved configuration: forwarding disabled."}
            {dirty && " Unsaved changes."}
          </p>
          <fieldset disabled={busy} className="flex flex-col gap-4 min-w-0">
            <label className="flex gap-3 items-center">
              <input
                type="checkbox"
                checked={draft.enabled}
                onChange={(e) => change({ enabled: e.target.checked })}
              />
              Enable forwarding for this profile
            </label>
            <label className="flex flex-col gap-2">
              ListenBrainz server base URL
              <input
                className={field}
                type="url"
                required={draft.enabled}
                disabled={!draft.enabled}
                placeholder="https://api.listenbrainz.org/1"
                value={draft.url}
                onChange={(e) => change({ url: e.target.value })}
              />
              <span className="text-sm">
                Include /1 (or /apis/listenbrainz/1 for another Koito). Do not
                append /submit-listens.
              </span>
            </label>
            <label className="flex flex-col gap-2">
              External user token
              <input
                className={field}
                type="text"
                autoComplete="off"
                spellCheck={false}
                required={draft.enabled}
                disabled={!draft.enabled}
                value={draft.token}
                onChange={(e) => change({ token: e.target.value })}
              />
              <span className="text-sm">
                Visible to administrators. Use this user's token from the
                destination server.
              </span>
            </label>
            <p className="text-sm">
              Saving with forwarding disabled removes the saved URL and token.
              Testing checks the current fields without saving or submitting a
              listen.
            </p>
            <div className="flex flex-wrap gap-3">
              <button className={button}>Save forwarding</button>
              <button
                type="button"
                className={button}
                disabled={
                  !draft.enabled || !draft.url.trim() || !draft.token.trim()
                }
                onClick={() =>
                  void run(async () => {
                    const result = await adminProfiles.testRelay(
                      slug,
                      draft.url,
                      draft.token,
                    );
                    setNotice(
                      `Connection successful. Token valid${result.user_name ? ` for ${result.user_name}` : ""}. Configuration has not been changed by this test.`,
                    );
                  })
                }
              >
                Test connection
              </button>
            </div>
          </fieldset>
        </>
      )}
    </form>
  );
}
