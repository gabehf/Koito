import { Link, useLoaderData, type LoaderFunctionArgs } from "react-router";
import TopTracks from "~/components/TopTracks";
import { mergeAlbums, type Album } from "api/api";
import LastPlayed from "~/components/LastPlayed";
import MediaLayout from "./MediaLayout";
import ActivityGrid from "~/components/ActivityGrid";
import { timeListenedString } from "~/utils/utils";
import InterestGraph from "~/components/InterestGraph";
import MediaItemNote from "~/components/MediaItemNote";
import ArtistAlbums from "~/components/ArtistAlbums";

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
      timeListened={album.time_listened}
      listenCount={album.listen_count}
      firstListen={album.first_listen}
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
        <>
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
        </>
      }
    >
      <div className="flex flex-col gap-10 md:gap-12 mt-14 max-w-[1400px]">
        <div className="flex gap-10 md:gap-20 flex-wrap lg:flex-nowrap items-start">
          <TopTracks limit={8} period={period} albumId={album.id} showSeeMore />
          <div className="min-w-[350px] flex-1">
            <LastPlayed
              limit={11}
              albumId={album.id}
              showNowPlaying
              showSeeMore
            />
          </div>
        </div>
        <div className="flex flex-wrap gap-10">
          <InterestGraph type="album" id={album.id} />
          <ActivityGrid configurable albumId={album.id} />
        </div>
      </div>
    </MediaLayout>
  );
}
