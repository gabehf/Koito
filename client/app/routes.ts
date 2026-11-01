import { type RouteConfig, index, route } from "@react-router/dev/routes";
export default [
  index("routes/Profiles.tsx"),
  route("admin", "routes/Admin.tsx"),
  route("theme-helper", "routes/ThemeHelper.tsx"),
  ...[
    "artist/:id",
    "album/:id",
    "track/:id",
    "chart/top/albums",
    "chart/top/artists",
    "chart/top/tracks",
    "listens",
    "rewind",
  ].map((path, i) =>
    route(path, "routes/LegacyRedirect.tsx", { id: `legacy-${i}` }),
  ),
  route(":slug", "routes/ProfileLayout.tsx", [
    index("routes/Home.tsx"),
    route("artist/:id", "routes/MediaItems/Artist.tsx"),
    route("album/:id", "routes/MediaItems/Album.tsx"),
    route("track/:id", "routes/MediaItems/Track.tsx"),
    route("chart/top/albums", "routes/Charts/AlbumChart.tsx"),
    route("chart/top/artists", "routes/Charts/ArtistChart.tsx"),
    route("chart/top/tracks", "routes/Charts/TrackChart.tsx"),
    route("listens", "routes/Charts/Listens.tsx"),
    route("rewind", "routes/RewindPage.tsx"),
  ]),
] satisfies RouteConfig;
