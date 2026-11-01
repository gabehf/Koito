import test from "node:test";
import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import { compileTypeScript } from "./typescript.mjs";
const source = await readFile(
  new URL("../api/profile-paths.ts", import.meta.url),
  "utf8",
);
const pathsURL = `data:text/javascript,${encodeURIComponent(compileTypeScript(source))}`;
const { profilePath, profileRequestPath, switchedProfilePath } = await import(
  pathsURL
);
const apiSource = (
  await readFile(new URL("../api/api.ts", import.meta.url), "utf8")
).replace('"./profile-paths"', JSON.stringify(pathsURL));
const { createProfileApi, updateUser } = await import(
  `data:text/javascript,${encodeURIComponent(compileTypeScript(apiSource))}`
);

test("profile reads and private actions always name their destination", () => {
  assert.equal(
    profileRequestPath("everton", "/apis/web/v1/stats?period=all_time"),
    "/apis/web/v1/profiles/everton/stats?period=all_time",
  );
  for (const [path, method] of [
    ["/listens", "POST"],
    ["/listens?track_id=1", "DELETE"],
    ["/export", "GET"],
    ["/user/apikeys/2", "DELETE"],
  ]) {
    assert.match(
      profileRequestPath("first", "/apis/web/v1" + path, method),
      /^\/apis\/web\/v1\/profiles\/first\//,
    );
    assert.throws(() =>
      profileRequestPath(undefined, "/apis/web/v1" + path, method),
    );
  }
  assert.equal(
    profileRequestPath("first", "/apis/web/v1/stats", "get"),
    "/apis/web/v1/profiles/first/stats",
  );
  assert.equal(
    profileRequestPath(undefined, "/apis/web/v1/login", "POST"),
    "/apis/web/v1/login",
  );
  assert.equal(
    profileRequestPath("first", "/apis/web/v1/album/3", "PATCH"),
    "/apis/web/v1/album/3",
  );
});
test("switch preserves general screens and periods, but drops person-specific media filters", () => {
  assert.equal(
    switchedProfilePath(
      "second",
      "/first/chart/top/artists",
      "?period=year&year=2025&page=3&artist_id=9",
      "first",
    ),
    "/second/chart/top/artists?period=year&year=2025",
  );
  assert.equal(
    switchedProfilePath(
      "second",
      "/first/track/9",
      "?period=all_time&track_id=9",
      "first",
    ),
    "/second?period=all_time",
  );
  assert.equal(
    switchedProfilePath(
      "second",
      "/first/rewind",
      "?year=2024&month=3",
      "first",
    ),
    "/second/rewind?year=2024&month=3",
  );
  assert.equal(switchedProfilePath("second", "/", "?next=/track/1"), "/second");
  assert.equal(
    profilePath("first", "https://example.com"),
    "https://example.com",
  );
  assert.equal(profilePath("first", "/artist/4"), "/first/artist/4");
});
test("captured API instances keep their destination even after another profile is selected", async () => {
  const original = globalThis.fetch;
  const requests = [];
  globalThis.fetch = async (url, init) => {
    requests.push({ url, signal: init.signal });
    return Response.json({ items: [] });
  };
  try {
    const controller = new AbortController();
    const first = createProfileApi("first", () => controller.signal);
    const second = createProfileApi("second");
    await first.search("one");
    await second.search("two");
    await first.deleteListen({ time: new Date(0), track: { id: 5 } });
    assert.match(requests[0].url, /\/profiles\/first\/search/);
    assert.match(requests[1].url, /\/profiles\/second\/search/);
    assert.match(requests[2].url, /\/profiles\/first\/listens/);
    controller.abort();
    assert.equal(requests[0].signal.aborted, true);
    assert.equal(requests[2].signal.aborted, true);
  } finally {
    globalThis.fetch = original;
  }
});

test("admin account changes omit unchanged credentials", async () => {
  const originalFetch = globalThis.fetch;
  const bodies = [];
  globalThis.fetch = async (_url, init) => {
    bodies.push(JSON.parse(init.body));
    return new Response(null, { status: 204 });
  };
  try {
    await updateUser("renamed-admin", "");
    await updateUser("", "new-password");
    assert.deepEqual(bodies, [
      { username: "renamed-admin" },
      { password: "new-password" },
    ]);
  } finally {
    globalThis.fetch = originalFetch;
  }
});
