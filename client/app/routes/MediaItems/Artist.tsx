import { useLoaderData, type LoaderFunctionArgs } from "react-router";
import TopTracks from "~/components/TopTracks";
import { mergeArtists, type Artist } from "api/api";
import LastPlayed from "~/components/LastPlayed";
import MediaLayout from "./MediaLayout";
import ArtistAlbums from "~/components/ArtistAlbums";
import ActivityGrid from "~/components/ActivityGrid";
import { timeListenedString } from "~/utils/utils";
import InterestGraph from "~/components/InterestGraph";
import MediaItemNote from "~/components/MediaItemNote";

export async function clientLoader({ params }: LoaderFunctionArgs) {
  const res = await fetch(`/apis/web/v1/artist/${params.id}`);
  if (!res.ok) {
    throw new Response("Failed to load artist", { status: 500 });
  }
  const artist: Artist = await res.json();
  return artist;
}

export default function Artist() {
  const artist = useLoaderData() as Artist;
  const period = "all_time";

  let index = artist.aliases.indexOf(artist.name);
  if (index !== -1) {
    artist.aliases.splice(index, 1);
  }

  return (
    <MediaLayout
      type="Artist"
      title={artist.name}
      img={artist.image}
      id={artist.id}
      rank={artist.all_time_rank}
      musicbrainzId={artist.musicbrainz_id}
      timeListened={artist.time_listened}
      listenCount={artist.listen_count}
      firstListen={artist.first_listen}
      imgItemId={artist.id}
      mergeFunc={mergeArtists}
      mergeCleanerFunc={(r, id) => {
        r.albums = [];
        r.tracks = [];
        for (let i = 0; i < r.artists.length; i++) {
          if (r.artists[i].id === id) {
            r.artists.splice(i, 1);
          }
        }
        return r;
      }}
      subContent={<></>}
    >
      <div className="flex flex-col gap-20">
        <div className="flex gap-10 md:gap-20 mt-10 flex-wrap items-center max-w-[1400px]">
          <div className="flex gap-10 md:gap-25 flex-wrap lg:flex-nowrap items-start">
            <TopTracks
              limit={8}
              period={period}
              artistId={artist.id}
              showSeeMore
            />
            <div className="min-w-[350px] w-2/5">
              <LastPlayed
                limit={11}
                artistId={artist.id}
                showNowPlaying
                showSeeMore
              />
            </div>
          </div>
          <div className="flex flex-wrap gap-10">
            <InterestGraph artistId={artist.id} />
            <ActivityGrid configurable artistId={artist.id} />
          </div>
        </div>
        <ArtistAlbums
          artistId={artist.id}
          header={`Albums featuring ${artist.name}`}
        />
      </div>
    </MediaLayout>
  );
}
