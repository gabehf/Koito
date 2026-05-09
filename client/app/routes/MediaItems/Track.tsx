import { Link, useLoaderData, type LoaderFunctionArgs } from "react-router";
import { mergeTracks, type Album, type Artist, type Track } from "api/api";
import LastPlayed from "~/components/LastPlayed";
import MediaLayout from "./MediaLayout";
import ActivityGrid from "~/components/ActivityGrid";
import { timeListenedString } from "~/utils/utils";
import InterestGraph from "~/components/InterestGraph";

export async function clientLoader({ params }: LoaderFunctionArgs) {
  let res = await fetch(`/apis/web/v1/track/${params.id}`);
  if (!res.ok) {
    throw new Response("Failed to load track", { status: res.status });
  }
  const track: Track = await res.json();
  res = await fetch(`/apis/web/v1/album/${track.album_id}`);
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
  const period = "all_time";

  return (
    <MediaLayout
      type="Track"
      title={track.title}
      img={track.image}
      id={track.id}
      rank={track.all_time_rank}
      musicbrainzId={track.musicbrainz_id}
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
          {track.artists.length > 0 && (
            <p>
              By{" "}
              {track.artists.map((artist: Artist, i: number) => (
                <span key={artist.id}>
                  {i > 0 && ", "}
                  {i === track.artists.length - 1 ? "and " : ""}
                  <Link className="hover:underline" to={`/artist/${artist.id}`}>
                    {artist.name}
                  </Link>
                </span>
              ))}
            </p>
          )}
          <p>
            Appears on{" "}
            <Link className="hover:underline" to={`/album/${track.album_id}`}>
              {album.title}
            </Link>
          </p>
          {track.listen_count !== 0 && (
            <p>
              {track.listen_count} play{track.listen_count > 1 ? "s" : ""}
            </p>
          )}
          {track.time_listened !== 0 && (
            <p title={Math.floor(track.time_listened / 60 / 60) + " hours"}>
              {timeListenedString(track.time_listened)}
            </p>
          )}
          {track.first_listen > 0 && (
            <p title={new Date(track.first_listen * 1000).toLocaleString()}>
              Listening since{" "}
              {new Date(track.first_listen * 1000).toLocaleDateString()}
            </p>
          )}
        </div>
      }
    >
      <div className="flex gap-10 md:gap-25 mt-10 flex-wrap max-w-[1000px]">
        <div className="w-2/5 max-w-[400px]">
          <LastPlayed limit={11} trackId={track.id} showNowPlaying />
        </div>
        <div className="flex flex-col xl:flex-row gap-10">
          <ActivityGrid configurable trackId={track.id} />
          <InterestGraph trackId={track.id} />
        </div>
      </div>
    </MediaLayout>
  );
}
