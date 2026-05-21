interface getItemsArgs {
  limit: number;
  period: string;
  page: number;
  artist_id?: number;
  album_id?: number;
  track_id?: number;
}
interface getActivityArgs {
  step: string;
  range: number;
  month: number;
  year: number;
  artist_id: number;
  album_id: number;
  track_id: number;
}
interface timeframe {
  week?: number;
  month?: number;
  year?: number;
  from?: number;
  to?: number;
  period?: string;
}
interface getInterestArgs {
  buckets: number;
  artist_id: number;
  album_id: number;
  track_id: number;
}

async function handleJson<T>(r: Response): Promise<T> {
  if (!r.ok) {
    const err = await r.json();
    throw Error(err.error);
  }
  return (await r.json()) as T;
}

export async function apiFetch<T>(
  path: string,
  params?: Record<string, string | number | undefined>,
): Promise<T> {
  let url = path;
  if (params) {
    const searchParams = new URLSearchParams();
    for (const [key, val] of Object.entries(params)) {
      if (val !== undefined) searchParams.set(key, String(val));
    }
    const qs = searchParams.toString();
    if (qs) url += "?" + qs;
  }
  const r = await fetch(url);
  return handleJson<T>(r);
}

async function getTopAlbums(
  args: getItemsArgs,
): Promise<PaginatedResponse<Ranked<Album>>> {
  let url = `/apis/web/v1/top/albums?period=${args.period}&limit=${args.limit}&page=${args.page}`;
  if (args.artist_id) url += `&artist_id=${args.artist_id}`;

  const r = await fetch(url);
  return handleJson<PaginatedResponse<Ranked<Album>>>(r);
}

function search(q: string): Promise<SearchResponse> {
  q = encodeURIComponent(q);
  return fetch(`/apis/web/v1/search?q=${q}`).then(
    (r) => r.json() as Promise<SearchResponse>,
  );
}

function imageUrl(id: string, size: string) {
  if (!id) {
    id = "default";
  }
  return `/image/${size}/${id}`;
}
function replaceImage(
  type: string,
  id: string,
  form: FormData,
): Promise<Response> {
  return fetch(`/apis/web/v1/${type}/${id}/image`, {
    method: "PATCH",
    body: form,
  });
}

function mergeTracks(from: number, to: number): Promise<Response> {
  return fetch(`/apis/web/v1/track/${to}/merge`, {
    method: "POST",
    body: JSON.stringify({ merge_from_id: from }),
  });
}
function mergeAlbums(
  from: number,
  to: number,
  replaceImage: boolean,
): Promise<Response> {
  return fetch(`/apis/web/v1/album/${to}/merge`, {
    method: "POST",
    body: JSON.stringify({ merge_from_id: from, replace_image: replaceImage }),
  });
}
function mergeArtists(
  from: number,
  to: number,
  replaceImage: boolean,
): Promise<Response> {
  return fetch(`/apis/web/v1/artist/${to}/merge`, {
    method: "POST",
    body: JSON.stringify({ merge_from_id: from, replace_image: replaceImage }),
  });
}
function login(
  username: string,
  password: string,
  remember: boolean,
): Promise<Response> {
  return fetch(`/apis/web/v1/login`, {
    method: "POST",
    body: JSON.stringify({
      username: username,
      password: password,
      remember_me: remember,
    }),
    headers: {
      "Content-Type": "application/json",
    },
  });
}
function logout(): Promise<Response> {
  return fetch(`/apis/web/v1/logout`, {
    method: "POST",
  });
}

function getCfg(): Promise<Config> {
  return fetch(`/apis/web/v1/config`).then((r) => r.json() as Promise<Config>);
}

function submitListen(id: string, ts: Date): Promise<Response> {
  const ms = new Date(ts).getTime();
  const unix = Math.floor(ms / 1000);
  return fetch(`/apis/web/v1/listens`, {
    method: "POST",
    body: JSON.stringify({
      track_id: Number(id),
      unix: unix,
      client: "Koito Web UI",
    }),
    headers: {
      "Content-Type": "application/json",
    },
  });
}

function getApiKeys(): Promise<ApiKey[]> {
  return fetch(`/apis/web/v1/user/apikeys`).then(
    (r) => r.json() as Promise<ApiKey[]>,
  );
}
const createApiKey = async (label: string): Promise<ApiKey> => {
  const r = await fetch(`/apis/web/v1/user/apikeys`, {
    method: "POST",
    body: JSON.stringify({ label: label }),
    headers: {
      "Content-Type": "application/json",
    },
  });
  if (!r.ok) {
    let errorMessage = `error: ${r.status}`;
    try {
      const errorData: ApiError = await r.json();
      if (errorData && typeof errorData.error === "string") {
        errorMessage = errorData.error;
      }
    } catch (e) {
      console.error("unexpected api error:", e);
    }
    throw new Error(errorMessage);
  }
  const data: ApiKey = await r.json();
  return data;
};
function deleteApiKey(id: number): Promise<Response> {
  return fetch(`/apis/web/v1/user/apikeys/${id}`, {
    method: "DELETE",
  });
}
function updateApiKeyLabel(id: number, label: string): Promise<Response> {
  return fetch(`/apis/web/v1/user/apikeys/${id}`, {
    method: "PATCH",
    body: JSON.stringify({ label: label }),
    headers: {
      "Content-Type": "application/json",
    },
  });
}

function deleteItem(itemType: string, id: number): Promise<Response> {
  return fetch(`/apis/web/v1/${itemType}/${id}`, {
    method: "DELETE",
  });
}
function updateUser(username: string, password: string) {
  const form = new URLSearchParams();
  form.append("username", username);
  form.append("password", password);
  return fetch(`/apis/web/v1/user`, {
    method: "PATCH",
    body: JSON.stringify({ username: username, password: password }),
    headers: {
      "Content-Type": "application/json",
    },
  });
}
function getAliases(type: string, id: number): Promise<Alias[]> {
  return fetch(`/apis/web/v1/${type}/${id}/aliases`).then(
    (r) => r.json() as Promise<Alias[]>,
  );
}
function createAlias(
  type: string,
  id: number,
  alias: string,
): Promise<Response> {
  return fetch(`/apis/web/v1/${type}/${id}/aliases`, {
    method: "POST",
    body: JSON.stringify({ alias: alias }),
  });
}
function deleteAlias(
  type: string,
  id: number,
  alias: string,
): Promise<Response> {
  return fetch(`/apis/web/v1/${type}/${id}/aliases`, {
    method: "DELETE",
    body: JSON.stringify({ alias: alias }),
  });
}
function setPrimaryAlias(
  type: string,
  id: number,
  alias: string,
): Promise<Response> {
  return fetch(`/apis/web/v1/${type}/${id}/aliases/primary`, {
    method: "PATCH",
    body: JSON.stringify({ alias: alias }),
  });
}
function updateMbzId(
  type: string,
  id: number,
  mbzid: string,
): Promise<Response> {
  return fetch(`/apis/web/v1/${type}/${id}`, {
    method: "PATCH",
    body: JSON.stringify({ mbid: mbzid }),
  });
}
function getAlbum(id: number): Promise<Album> {
  return fetch(`/apis/web/v1/album/${id}`).then(
    (r) => r.json() as Promise<Album>,
  );
}

function deleteListen(listen: Listen): Promise<Response> {
  const ms = new Date(listen.time).getTime();
  const unix = Math.floor(ms / 1000);
  return fetch(`/apis/web/v1/listen?track_id=${listen.track.id}&unix=${unix}`, {
    method: "DELETE",
  });
}
function getExport() {}

async function getRewindStats(args: timeframe): Promise<RewindStats> {
  const r = await fetch(
    `/apis/web/v1/summary?week=${args.week}&month=${args.month}&year=${args.year}&from=${args.from}&to=${args.to}`,
  );
  return handleJson<RewindStats>(r);
}

export {
  getTopAlbums,
  search,
  replaceImage,
  mergeTracks,
  mergeAlbums,
  mergeArtists,
  imageUrl,
  login,
  logout,
  getCfg,
  deleteItem,
  updateUser,
  getAliases,
  createAlias,
  deleteAlias,
  setPrimaryAlias,
  updateMbzId,
  getApiKeys,
  createApiKey,
  deleteApiKey,
  updateApiKeyLabel,
  deleteListen,
  getAlbum,
  getExport,
  submitListen,
  getRewindStats,
};
type ImageList = {
  xs: string;
  small: string;
  medium: string;
  large: string;
  xl: string;
};
type Track = {
  id: number;
  title: string;
  artists: SimpleArtists[];
  listen_count: number;
  image: ImageList;
  album_id: number;
  musicbrainz_id: string;
  time_listened: number;
  first_listen: number;
  all_time_rank: number;
};
type SimpleTrack = {
  id: number;
  title: string;
  artists: SimpleArtists[];
  image: ImageList;
};
type Artist = {
  id: number;
  name: string;
  image: ImageList;
  aliases: string[];
  listen_count: number;
  musicbrainz_id: string;
  time_listened: number;
  first_listen: number;
  is_primary: boolean;
  all_time_rank: number;
};
type Album = {
  id: number;
  title: string;
  image: ImageList;
  listen_count: number;
  is_various_artists: boolean;
  artists: SimpleArtists[];
  musicbrainz_id: string;
  time_listened: number;
  first_listen: number;
  all_time_rank: number;
};
type Alias = {
  id: number;
  alias: string;
  source: string;
  is_primary: boolean;
};
type Listen = {
  time: string;
  track: SimpleTrack;
};
type PaginatedResponse<T> = {
  items: T[];
  total_record_count: number;
  has_next_page: boolean;
  current_page: number;
  items_per_page: number;
};
type Ranked<T> = {
  item: T;
  rank: number;
};
type ListenActivityItem = {
  start_time: Date;
  listens: number;
};
type InterestBucket = {
  bucket_start: Date;
  bucket_end: Date;
  listen_count: number;
};
type SimpleArtists = {
  name: string;
  id: number;
};
type Stats = {
  listen_count: number;
  track_count: number;
  album_count: number;
  artist_count: number;
  minutes_listened: number;
};
type SearchResponse = {
  albums: Album[];
  artists: Artist[];
  tracks: Track[];
};
type User = {
  id: number;
  username: string;
  role: "user" | "admin";
};
type ApiKey = {
  id: number;
  key: string;
  label: string;
  created_at: Date;
};
type ApiError = {
  error: string;
};
type Config = {
  default_theme: string;
};
type NowPlaying = {
  currently_playing: boolean;
  track: Track;
};
type RewindStats = {
  title: string;
  top_artists: Ranked<Artist>[];
  top_albums: Ranked<Album>[];
  top_tracks: Ranked<Track>[];
  minutes_listened: number;
  avg_minutes_listened_per_day: number;
  plays: number;
  avg_plays_per_day: number;
  unique_tracks: number;
  unique_albums: number;
  unique_artists: number;
  new_tracks: number;
  new_albums: number;
  new_artists: number;
};

export type {
  getItemsArgs,
  getActivityArgs,
  getInterestArgs,
  Track,
  Artist,
  Album,
  Listen,
  SearchResponse,
  PaginatedResponse,
  Ranked,
  ListenActivityItem,
  InterestBucket,
  User,
  Alias,
  ApiKey,
  ApiError,
  Config,
  NowPlaying,
  Stats,
  RewindStats,
  ImageList,
};
