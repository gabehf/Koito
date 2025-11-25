import { useState } from "react";
import { Link, useLoaderData, type LoaderFunctionArgs } from "react-router";
import { mergeTracks, type Album, type Track } from "api/api";
import LastPlays from "~/components/LastPlays";
import PeriodSelector from "~/components/PeriodSelector";
import MediaLayout from "./MediaLayout";
import ActivityGrid from "~/components/ActivityGrid";
import { timeListenedString } from "~/utils/utils";

export async function clientLoader({ params }: LoaderFunctionArgs) {
  let res = await fetch(`/apis/web/v1/track?id=${params.id}`);
  if (!res.ok) {
    throw new Response("Failed to load track", { status: res.status });
  }
  const track: Track = await res.json();
  res = await fetch(`/apis/web/v1/album?id=${track.album_id}`);
  if (!res.ok) {
    throw new Response("Failed to load album for track", {
      status: res.status,
    });
  }
  const album: Album = await res.json();
  return { track: track, album: album };
}

export default function Track() {
  const { track, album } = useLoaderData();
  const [period, setPeriod] = useState("week");

  return (
    <MediaLayout
      type="Track"
      title={track.title}
      img={track.image}
      id={track.id}
      musicbrainzId={album.musicbrainz_id}
      imgItemId={track.album_id}
      mergeFunc={mergeTracks}
      mergeCleanerFunc={(r, id) => {
        r.albums = [];
        r.artists = [];
        for (let i = 0; i < r.tracks.length; i++) {
          if (r.tracks[i].id === id) {
            delete r.tracks[i];
          }
        }
        return r;
      }}
      subContent={
        <div className="flex flex-col gap-2 items-start">
          <Link to={`/album/${track.album_id}`}>appears on {album.title}</Link>
          {track.listen_count && (
            <p>
              {track.listen_count} play{track.listen_count > 1 ? "s" : ""}
            </p>
          )}
          {
            <p title={Math.floor(track.time_listened / 60 / 60) + " hours"}>
              {timeListenedString(track.time_listened)}
            </p>
          }
          {
            <p title={new Date(track.first_listen * 1000).toLocaleString()}>
              Listening since{" "}
              {new Date(track.first_listen * 1000).toLocaleDateString()}
            </p>
          }
        </div>
      }
    >
      <div className="mt-10">
        <PeriodSelector setter={setPeriod} current={period} />
      </div>
      <div className="flex flex-wrap gap-20 mt-10">
        <LastPlays limit={20} trackId={track.id} />
        <ActivityGrid trackId={track.id} configurable />
      </div>
    </MediaLayout>
  );
}
