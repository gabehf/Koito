import { Link, useLoaderData, type LoaderFunctionArgs } from "react-router";
import TopTracks from "~/components/TopTracks";
import { mergeAlbums, type Album } from "api/api";
import LastPlayed from "~/components/LastPlayed";
import MediaLayout from "./MediaLayout";
import ActivityGrid from "~/components/ActivityGrid";
import { timeListenedString } from "~/utils/utils";
import InterestGraph from "~/components/InterestGraph";
import MediaItemNote from "~/components/MediaItemNote";

export async function clientLoader({ params }: LoaderFunctionArgs) {
  const res = await fetch(`/apis/web/v1/album/${params.id}`);
  if (!res.ok) {
    throw new Response("Failed to load album", { status: 500 });
  }
  const album: Album = await res.json();
  return album;
}

export default function Album() {
  const album = useLoaderData() as Album;
  const period = "all_time";

  console.log(album);

  return (
    <MediaLayout
      type="Album"
      title={album.title}
      img={album.image}
      id={album.id}
      rank={album.all_time_rank}
      musicbrainzId={album.musicbrainz_id}
      imgItemId={album.id}
      mergeFunc={mergeAlbums}
      mergeCleanerFunc={(r, id) => {
        r.artists = [];
        r.tracks = [];
        for (let i = 0; i < r.albums.length; i++) {
          if (r.albums[i].id === id) {
            r.albums.splice(i, 1);
          }
        }
        return r;
      }}
      subContent={
        <div className="flex flex-col gap-1.5 items-start">
          {album.artists.length > 0 && !album.is_various_artists && (
            <p>
              By{" "}
              {
                <span key={album.artists[0].id}>
                  <Link
                    className="hover:underline"
                    to={`/artist/${album.artists[0].id}`}
                  >
                    {album.artists[0].name}
                  </Link>
                </span>
              }
            </p>
          )}
          {album.is_various_artists && <p>By Various Artists</p>}
          {album.listen_count !== 0 && (
            <p>
              {album.listen_count} play{album.listen_count > 1 ? "s" : ""}
            </p>
          )}
          {album.time_listened !== 0 && (
            <p title={Math.floor(album.time_listened / 60 / 60) + " hours"}>
              {timeListenedString(album.time_listened)}
            </p>
          )}
          {album.first_listen > 0 && (
            <p title={new Date(album.first_listen * 1000).toLocaleString()}>
              Listening since{" "}
              {new Date(album.first_listen * 1000).toLocaleDateString()}
            </p>
          )}
        </div>
      }
    >
      <div className="flex flex-col gap-20">
        <div className="flex gap-10 md:gap-25 mt-10 flex-wrap items-center">
          <div className="flex gap-10 md:gap-25 flex-wrap lg:flex-nowrap items-start">
            <TopTracks limit={8} period={period} albumId={album.id} />
            <div className="min-w-[350px] w-2/5 max-w-[400px]">
              <LastPlayed limit={11} albumId={album.id} showNowPlaying />
            </div>
          </div>
          <div className="flex flex-col gap-10">
            <InterestGraph albumId={album.id} />
            <ActivityGrid configurable albumId={album.id} />
          </div>
        </div>
      </div>
    </MediaLayout>
  );
}
