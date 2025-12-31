import { type RouteConfig, index, route } from "@react-router/dev/routes";

export default [
  index("routes/Home.tsx"),
  route("/artist/:id", "routes/MediaItems/Artist.tsx"),
  route("/album/:id", "routes/MediaItems/Album.tsx"),
  route("/track/:id", "routes/MediaItems/Track.tsx"),
  route("/chart/top-albums", "routes/Charts/AlbumChart.tsx"),
  route("/chart/top-artists", "routes/Charts/ArtistChart.tsx"),
  route("/chart/top-tracks", "routes/Charts/TrackChart.tsx"),
  route("/listens", "routes/Charts/Listens.tsx"),
  route("/rewind", "routes/RewindPage.tsx"),
  route("/theme-helper", "routes/ThemeHelper.tsx"),
] satisfies RouteConfig;
