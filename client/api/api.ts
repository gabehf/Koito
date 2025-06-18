interface getItemsArgs {
    limit: number,
    period: string,
    page: number,
    artist_id?: number,
    album_id?: number,
    track_id?: number
}
interface getActivityArgs {
    step: string 
    range: number 
    month: number 
    year: number 
    artist_id: number
    album_id: number
    track_id: number
}

function getLastListens(args: getItemsArgs): Promise<PaginatedResponse<Listen>> {
    return fetch(`/apis/web/v1/listens?period=${args.period}&limit=${args.limit}&artist_id=${args.artist_id}&album_id=${args.album_id}&track_id=${args.track_id}&page=${args.page}`).then(r => r.json() as Promise<PaginatedResponse<Listen>>)
}

function getTopTracks(args: getItemsArgs): Promise<PaginatedResponse<Track>> {
    if (args.artist_id) {
        return fetch(`/apis/web/v1/top-tracks?period=${args.period}&limit=${args.limit}&artist_id=${args.artist_id}&page=${args.page}`).then(r => r.json() as Promise<PaginatedResponse<Track>>)
    } else if (args.album_id) {
        return fetch(`/apis/web/v1/top-tracks?period=${args.period}&limit=${args.limit}&album_id=${args.album_id}&page=${args.page}`).then(r => r.json() as Promise<PaginatedResponse<Track>>)
    } else {
        return fetch(`/apis/web/v1/top-tracks?period=${args.period}&limit=${args.limit}&page=${args.page}`).then(r => r.json() as Promise<PaginatedResponse<Track>>)
    }
}

function getTopAlbums(args: getItemsArgs): Promise<PaginatedResponse<Album>> {
    const baseUri = `/apis/web/v1/top-albums?period=${args.period}&limit=${args.limit}&page=${args.page}`
    if (args.artist_id) {
        return fetch(baseUri+`&artist_id=${args.artist_id}`).then(r => r.json() as Promise<PaginatedResponse<Album>>)
    } else {
        return fetch(baseUri).then(r => r.json() as Promise<PaginatedResponse<Album>>)
    }
}

function getTopArtists(args: getItemsArgs): Promise<PaginatedResponse<Artist>> {
    const baseUri = `/apis/web/v1/top-artists?period=${args.period}&limit=${args.limit}&page=${args.page}`
    return fetch(baseUri).then(r => r.json() as Promise<PaginatedResponse<Artist>>)
}

function getActivity(args: getActivityArgs): Promise<ListenActivityItem[]> {
    return fetch(`/apis/web/v1/listen-activity?step=${args.step}&range=${args.range}&month=${args.month}&year=${args.year}&album_id=${args.album_id}&artist_id=${args.artist_id}&track_id=${args.track_id}`).then(r => r.json() as Promise<ListenActivityItem[]>)
}

function getStats(period: string): Promise<Stats> {
    return fetch(`/apis/web/v1/stats?period=${period}`).then(r => r.json() as Promise<Stats>)
}

function search(q: string): Promise<SearchResponse> {
    q = encodeURIComponent(q)
    return fetch(`/apis/web/v1/search?q=${q}`).then(r => r.json() as Promise<SearchResponse>)
}

function imageUrl(id: string, size: string) {
    if (!id) {
        id = 'default'
    }
    return `/images/${size}/${id}`
}
function replaceImage(form: FormData): Promise<Response> {
    return fetch(`/apis/web/v1/replace-image`, {
        method: "POST",
        body: form,
    })
}

function mergeTracks(from: number, to: number): Promise<Response> {
    return fetch(`/apis/web/v1/merge/tracks?from_id=${from}&to_id=${to}`, {
        method: "POST",
    })
}
function mergeAlbums(from: number, to: number, replaceImage: boolean): Promise<Response> {
    return fetch(`/apis/web/v1/merge/albums?from_id=${from}&to_id=${to}&replace_image=${replaceImage}`, {
        method: "POST",
    })
}
function mergeArtists(from: number, to: number, replaceImage: boolean): Promise<Response> {
    return fetch(`/apis/web/v1/merge/artists?from_id=${from}&to_id=${to}&replace_image=${replaceImage}`, {
        method: "POST",
    })
}
function login(username: string, password: string, remember: boolean): Promise<Response> {
    const form = new URLSearchParams 
    form.append('username', username)
    form.append('password', password)
    form.append('remember_me', String(remember))
    return fetch(`/apis/web/v1/login`, {
        method: "POST",
        body: form,
    })
}
function logout(): Promise<Response> {
    return fetch(`/apis/web/v1/logout`, {
        method: "POST",
    })
}

function getApiKeys(): Promise<ApiKey[]> {
    return fetch(`/apis/web/v1/user/apikeys`).then((r) => r.json() as Promise<ApiKey[]>)
}
const createApiKey = async (label: string): Promise<ApiKey> => {
    const form = new URLSearchParams 
    form.append('label', label)
    const r = await fetch(`/apis/web/v1/user/apikeys`, {
        method: "POST",
        body: form,
    });
    if (!r.ok) {
        let errorMessage = `error: ${r.status}`;
        try {
            const errorData: ApiError = await r.json();
            if (errorData && typeof errorData.error === 'string') {
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
        method: "DELETE"
    })
}
function updateApiKeyLabel(id: number, label: string): Promise<Response> {
    const form = new URLSearchParams 
    form.append('id', String(id))
    form.append('label', label)
    return fetch(`/apis/web/v1/user/apikeys`, {
        method: "PATCH",
        body: form,
    })
}

function deleteItem(itemType: string, id: number): Promise<Response> {
    return fetch(`/apis/web/v1/${itemType}?id=${id}`, {
        method: "DELETE"
    })
}
function updateUser(username: string, password: string) {
    const form = new URLSearchParams 
    form.append('username', username)
    form.append('password', password)
    return fetch(`/apis/web/v1/user`, {
        method: "PATCH",
        body: form,
    })
}
function getAliases(type: string, id: number): Promise<Alias[]> {
    return fetch(`/apis/web/v1/aliases?${type}_id=${id}`).then(r => r.json() as Promise<Alias[]>)
}
function createAlias(type: string, id: number, alias: string): Promise<Response> {
    const form = new URLSearchParams 
    form.append(`${type}_id`, String(id))
    form.append('alias', alias)
    return fetch(`/apis/web/v1/aliases`, {
        method: 'POST',
        body: form,
    })
}
function deleteAlias(type: string, id: number, alias: string): Promise<Response> {
    const form = new URLSearchParams 
    form.append(`${type}_id`, String(id))
    form.append('alias', alias)
    return fetch(`/apis/web/v1/aliases/delete`, {
        method: "POST",
        body: form,
    })
}
function setPrimaryAlias(type: string, id: number, alias: string): Promise<Response> {
    const form = new URLSearchParams 
    form.append(`${type}_id`, String(id))
    form.append('alias', alias)
    return fetch(`/apis/web/v1/aliases/primary`, {
        method: "POST",
        body: form,
    })
}
function getAlbum(id: number): Promise<Album> {
    return fetch(`/apis/web/v1/album?id=${id}`).then(r => r.json() as Promise<Album>)
}

function deleteListen(listen: Listen): Promise<Response> {
    const ms = new Date(listen.time).getTime()
    const unix= Math.floor(ms / 1000); 
    return fetch(`/apis/web/v1/listen?track_id=${listen.track.id}&unix=${unix}`, {
        method: "DELETE"
    })
}
function getExport() {
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
}
type Track = {
    id: number
    title: string
    artists: SimpleArtists[]
    listen_count: number
    image: string
    album_id: number
    musicbrainz_id: string
    time_listened: number
}
type Artist = {
    id: number
    name: string
    image: string,
    aliases: string[]
    listen_count: number
    musicbrainz_id: string
    time_listened: number
    is_primary: boolean
}
type Album = {
    id: number,
    title: string
    image: string
    listen_count: number
    is_various_artists: boolean
    artists: SimpleArtists[]
    musicbrainz_id: string
    time_listened: number
}
type Alias = {
    id: number 
    alias: string 
    source: string
    is_primary: boolean
}
type Listen = {
    time: string,
    track: Track,
}
type PaginatedResponse<T> = {
    items: T[],
    total_record_count: number,
    has_next_page: boolean,
    current_page: number,
    items_per_page: number,
}
type ListenActivityItem = {
    start_time: Date,
    listens: number
}
type SimpleArtists = {
    name: string 
    id: number
}
type Stats = {
    listen_count: number 
    track_count: number 
    album_count: number 
    artist_count: number 
    hours_listened: number
}
type SearchResponse = {
    albums: Album[]
    artists: Artist[]
    tracks: Track[]
}
type User = {
    id: number
    username: string 
    role: 'user' | 'admin'
}
type ApiKey = {
    id: number
    key: string
    label: string
    created_at: Date
}
type ApiError = {
    error: string
}

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
    ApiError
}
