import type { ListeningProfile, ApiKey } from "./api";

export interface ProfileRelay {
  enabled: boolean;
  url: string;
  token: string;
}

const base = "/apis/web/v1/profiles";
export async function adminRequest<T>(
  path: string,
  init: RequestInit = {},
): Promise<T> {
  const response = await fetch(base + path, {
    ...init,
    credentials: "same-origin",
  });
  if (!response.ok) {
    const error = await response.json().catch(() => null);
    throw new Error(error?.error || `Request failed (${response.status})`);
  }
  return response.status === 204 ? (undefined as T) : response.json();
}
const json = (method: string, body: unknown): RequestInit => ({
  method,
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify(body),
});
const profilePath = (slug: string) => `/${encodeURIComponent(slug)}`;
export const adminProfiles = {
  list: (signal?: AbortSignal) =>
    adminRequest<ListeningProfile[]>("", { signal }),
  create: (display_name: string, slug: string, avatar_icon: string) =>
    adminRequest<ListeningProfile>(
      "",
      json("POST", { display_name, slug, avatar_icon }),
    ),
  update: (slug: string, display_name: string, nextSlug: string) =>
    adminRequest<ListeningProfile>(
      profilePath(slug),
      json("PATCH", { display_name, slug: nextSlug }),
    ),
  details: (slug: string, signal?: AbortSignal) =>
    adminRequest<{ profile: ListeningProfile; listen_count: number }>(
      profilePath(slug) + "/management",
      { signal },
    ),
  remove: (slug: string, display_name: string, listen_count: number) =>
    adminRequest<void>(
      profilePath(slug),
      json("DELETE", { display_name, listen_count }),
    ),
  avatar: (slug: string, icon: string, file?: File) => {
    const body = new FormData();
    body.set("icon", icon);
    if (file) body.set("image", file);
    return adminRequest<ListeningProfile>(profilePath(slug) + "/avatar", {
      method: "PUT",
      body,
    });
  },
  clearAvatar: (slug: string) =>
    adminRequest<ListeningProfile>(profilePath(slug) + "/avatar", {
      method: "DELETE",
    }),
  relay: (slug: string, signal?: AbortSignal) =>
    adminRequest<ProfileRelay>(profilePath(slug) + "/relay", {
      signal,
      cache: "no-store",
    }),
  saveRelay: (slug: string, url: string, token: string) =>
    adminRequest<void>(
      profilePath(slug) + "/relay",
      json("PUT", { url, token }),
    ),
  removeRelay: (slug: string) =>
    adminRequest<void>(profilePath(slug) + "/relay", { method: "DELETE" }),
  testRelay: (slug: string, url: string, token: string) =>
    adminRequest<{ valid: boolean; user_name: string }>(
      profilePath(slug) + "/relay/test",
      json("POST", { url, token }),
    ),
  keys: (slug: string, signal?: AbortSignal) =>
    adminRequest<ApiKey[]>(profilePath(slug) + "/user/apikeys", { signal }),
  createKey: (slug: string, label: string) =>
    adminRequest<ApiKey>(
      profilePath(slug) + "/user/apikeys",
      json("POST", { label }),
    ),
  labelKey: (slug: string, id: number, label: string) =>
    adminRequest<void>(
      profilePath(slug) + `/user/apikeys/${id}`,
      json("PATCH", { label }),
    ),
  revokeKey: (slug: string, id: number) =>
    adminRequest<void>(profilePath(slug) + `/user/apikeys/${id}`, {
      method: "DELETE",
    }),
  importFile: (slug: string, format: string, file: File) => {
    const body = new FormData();
    body.set("format", format);
    body.set("file", file);
    return adminRequest<void>(profilePath(slug) + "/import", {
      method: "POST",
      body,
    });
  },
};
