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
      timeListened={track.time_listened}
      listenCount={track.listen_count}
      firstListen={track.first_listen}
      imgItemId={track.album_id}
      mergeFunc={mergeTracks}
      mergeCleanerFunc={(r, id) => {
        r.albums = [];
        r.artists = [];
        for (let i = 0; i < r.tracks.length; i++) {
          if (r.tracks[i].id === id) {
            r.tracks.splice(i, 1);
          }
        }
        return r;
      }}
      subContent={
        <>
          {track.artists.length > 0 && (
            <p>
              By{" "}
              {track.artists.map((artist: Artist, i: number) => (
                <span key={artist.id}>
                  {i > 0 && i !== track.artists.length - 1 && ", "}
                  {i === track.artists.length - 1 && track.artists.length > 1
                    ? " and "
                    : ""}
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
        </>
      }
    >
      <div className="flex flex-col gap-10 md:gap-20 mt-14">
        <div className="min-w-[350px] flex-1">
          <LastPlayed limit={11} trackId={track.id} showNowPlaying />
        </div>
        <div className="flex flex-wrap gap-10">
          <InterestGraph type="track" id={track.id} />
          <ActivityGrid configurable trackId={track.id} />
        </div>
      </div>
    </MediaLayout>
  );
}
