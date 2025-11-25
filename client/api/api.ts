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

async function handleJson<T>(r: Response): Promise<T> {
  if (!r.ok) {
    const err = await r.json();
    throw Error(err.error);
  }
  return (await r.json()) as T;
}
async function getLastListens(
  args: getItemsArgs
): Promise<PaginatedResponse<Listen>> {
  const r = await fetch(
    `/apis/web/v1/listens?period=${args.period}&limit=${args.limit}&artist_id=${args.artist_id}&album_id=${args.album_id}&track_id=${args.track_id}&page=${args.page}`
  );
  return handleJson<PaginatedResponse<Listen>>(r);
}

async function getTopTracks(
  args: getItemsArgs
): Promise<PaginatedResponse<Track>> {
  let url = `/apis/web/v1/top-tracks?period=${args.period}&limit=${args.limit}&page=${args.page}`;

  if (args.artist_id) url += `&artist_id=${args.artist_id}`;
  else if (args.album_id) url += `&album_id=${args.album_id}`;

  const r = await fetch(url);
  return handleJson<PaginatedResponse<Track>>(r);
}

async function getTopAlbums(
  args: getItemsArgs
): Promise<PaginatedResponse<Album>> {
  let url = `/apis/web/v1/top-albums?period=${args.period}&limit=${args.limit}&page=${args.page}`;
  if (args.artist_id) url += `&artist_id=${args.artist_id}`;

  const r = await fetch(url);
  return handleJson<PaginatedResponse<Album>>(r);
}

async function getTopArtists(
  args: getItemsArgs
): Promise<PaginatedResponse<Artist>> {
  const url = `/apis/web/v1/top-artists?period=${args.period}&limit=${args.limit}&page=${args.page}`;
  const r = await fetch(url);
  return handleJson<PaginatedResponse<Artist>>(r);
}

async function getActivity(
  args: getActivityArgs
): Promise<ListenActivityItem[]> {
  const r = await fetch(
    `/apis/web/v1/listen-activity?step=${args.step}&range=${args.range}&month=${args.month}&year=${args.year}&album_id=${args.album_id}&artist_id=${args.artist_id}&track_id=${args.track_id}`
  );
  return handleJson<ListenActivityItem[]>(r);
}

async function getStats(period: string): Promise<Stats> {
  const r = await fetch(`/apis/web/v1/stats?period=${period}`);

  return handleJson<Stats>(r);
}

function search(q: string): Promise<SearchResponse> {
  q = encodeURIComponent(q);
  return fetch(`/apis/web/v1/search?q=${q}`).then(
    (r) => r.json() as Promise<SearchResponse>
  );
}

function imageUrl(id: string, size: string) {
  if (!id) {
    id = "default";
  }
  return `/images/${size}/${id}`;
}
function replaceImage(form: FormData): Promise<Response> {
  return fetch(`/apis/web/v1/replace-image`, {
    method: "POST",
    body: form,
  });
}

function mergeTracks(from: number, to: number): Promise<Response> {
  return fetch(`/apis/web/v1/merge/tracks?from_id=${from}&to_id=${to}`, {
    method: "POST",
  });
}
function mergeAlbums(
  from: number,
  to: number,
  replaceImage: boolean
): Promise<Response> {
  return fetch(
    `/apis/web/v1/merge/albums?from_id=${from}&to_id=${to}&replace_image=${replaceImage}`,
    {
      method: "POST",
    }
  );
}
function mergeArtists(
  from: number,
  to: number,
  replaceImage: boolean
): Promise<Response> {
  return fetch(
    `/apis/web/v1/merge/artists?from_id=${from}&to_id=${to}&replace_image=${replaceImage}`,
    {
      method: "POST",
    }
  );
}
function login(
  username: string,
  password: string,
  remember: boolean
): Promise<Response> {
  const form = new URLSearchParams();
  form.append("username", username);
  form.append("password", password);
  form.append("remember_me", String(remember));
  return fetch(`/apis/web/v1/login`, {
    method: "POST",
    body: form,
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
  const form = new URLSearchParams();
  form.append("track_id", id);
  const ms = new Date(ts).getTime();
  const unix = Math.floor(ms / 1000);
  form.append("unix", unix.toString());
  return fetch(`/apis/web/v1/listen`, {
    method: "POST",
    body: form,
  });
}

function getApiKeys(): Promise<ApiKey[]> {
  return fetch(`/apis/web/v1/user/apikeys`).then(
    (r) => r.json() as Promise<ApiKey[]>
  );
}
const createApiKey = async (label: string): Promise<ApiKey> => {
  const form = new URLSearchParams();
  form.append("label", label);
  const r = await fetch(`/apis/web/v1/user/apikeys`, {
    method: "POST",
    body: form,
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
  return fetch(`/apis/web/v1/user/apikeys?id=${id}`, {
    method: "DELETE",
  });
}
function updateApiKeyLabel(id: number, label: string): Promise<Response> {
  const form = new URLSearchParams();
  form.append("id", String(id));
  form.append("label", label);
  return fetch(`/apis/web/v1/user/apikeys`, {
    method: "PATCH",
    body: form,
  });
}

function deleteItem(itemType: string, id: number): Promise<Response> {
  return fetch(`/apis/web/v1/${itemType}?id=${id}`, {
    method: "DELETE",
  });
}
function updateUser(username: string, password: string) {
  const form = new URLSearchParams();
  form.append("username", username);
  form.append("password", password);
  return fetch(`/apis/web/v1/user`, {
    method: "PATCH",
    body: form,
  });
}
function getAliases(type: string, id: number): Promise<Alias[]> {
  return fetch(`/apis/web/v1/aliases?${type}_id=${id}`).then(
    (r) => r.json() as Promise<Alias[]>
  );
}
function createAlias(
  type: string,
  id: number,
  alias: string
): Promise<Response> {
  const form = new URLSearchParams();
  form.append(`${type}_id`, String(id));
  form.append("alias", alias);
  return fetch(`/apis/web/v1/aliases`, {
    method: "POST",
    body: form,
  });
}
function deleteAlias(
  type: string,
  id: number,
  alias: string
): Promise<Response> {
  const form = new URLSearchParams();
  form.append(`${type}_id`, String(id));
  form.append("alias", alias);
  return fetch(`/apis/web/v1/aliases/delete`, {
    method: "POST",
    body: form,
  });
}
function setPrimaryAlias(
  type: string,
  id: number,
  alias: string
): Promise<Response> {
  const form = new URLSearchParams();
  form.append(`${type}_id`, String(id));
  form.append("alias", alias);
  return fetch(`/apis/web/v1/aliases/primary`, {
    method: "POST",
    body: form,
  });
}
function getAlbum(id: number): Promise<Album> {
  return fetch(`/apis/web/v1/album?id=${id}`).then(
    (r) => r.json() as Promise<Album>
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

function getNowPlaying(): Promise<NowPlaying> {
  return fetch("/apis/web/v1/now-playing").then((r) => r.json());
}

export {
  getLastListens,
  getTopTracks,
  getTopAlbums,
  getTopArtists,
  getActivity,
  getStats,
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
  getApiKeys,
  createApiKey,
  deleteApiKey,
  updateApiKeyLabel,
  deleteListen,
  getAlbum,
  getExport,
  submitListen,
  getNowPlaying,
};
type Track = {
  id: number;
  title: string;
  artists: SimpleArtists[];
  listen_count: number;
  image: string;
  album_id: number;
  musicbrainz_id: string;
  time_listened: number;
  first_listen: number;
};
type Artist = {
  id: number;
  name: string;
  image: string;
  aliases: string[];
  listen_count: number;
  musicbrainz_id: string;
  time_listened: number;
  first_listen: number;
  is_primary: boolean;
};
type Album = {
  id: number;
  title: string;
  image: string;
  listen_count: number;
  is_various_artists: boolean;
  artists: SimpleArtists[];
  musicbrainz_id: string;
  time_listened: number;
  first_listen: number;
};
type Alias = {
  id: number;
  alias: string;
  source: string;
  is_primary: boolean;
};
type Listen = {
  time: string;
  track: Track;
};
type PaginatedResponse<T> = {
  items: T[];
  total_record_count: number;
  has_next_page: boolean;
  current_page: number;
  items_per_page: number;
};
type ListenActivityItem = {
  start_time: Date;
  listens: number;
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

export type {
  getItemsArgs,
  getActivityArgs,
  Track,
  Artist,
  Album,
  Listen,
  SearchResponse,
  PaginatedResponse,
  ListenActivityItem,
  User,
  Alias,
  ApiKey,
  ApiError,
  Config,
  NowPlaying,
  Stats,
};
