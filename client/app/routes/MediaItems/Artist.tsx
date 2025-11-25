import { useState } from "react";
import { useLoaderData, type LoaderFunctionArgs } from "react-router";
import TopTracks from "~/components/TopTracks";
import { mergeArtists, type Artist } from "api/api";
import LastPlays from "~/components/LastPlays";
import PeriodSelector from "~/components/PeriodSelector";
import MediaLayout from "./MediaLayout";
import ArtistAlbums from "~/components/ArtistAlbums";
import ActivityGrid from "~/components/ActivityGrid";
import { timeListenedString } from "~/utils/utils";

export async function clientLoader({ params }: LoaderFunctionArgs) {
  const res = await fetch(`/apis/web/v1/artist?id=${params.id}`);
  if (!res.ok) {
    throw new Response("Failed to load artist", { status: 500 });
  }
  const artist: Artist = await res.json();
  return artist;
}

export default function Artist() {
  const artist = useLoaderData() as Artist;
  const [period, setPeriod] = useState("week");

  // remove canonical name from alias list
  console.log(artist.aliases);
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
      musicbrainzId={artist.musicbrainz_id}
      imgItemId={artist.id}
      mergeFunc={mergeArtists}
      mergeCleanerFunc={(r, id) => {
        r.albums = [];
        r.tracks = [];
        for (let i = 0; i < r.artists.length; i++) {
          if (r.artists[i].id === id) {
            delete r.artists[i];
          }
        }
        return r;
      }}
      subContent={
        <div className="flex flex-col gap-2 items-start">
          {artist.listen_count && (
            <p>
              {artist.listen_count} play{artist.listen_count > 1 ? "s" : ""}
            </p>
          )}
          {
            <p title={Math.floor(artist.time_listened / 60 / 60) + " hours"}>
              {timeListenedString(artist.time_listened)}
            </p>
          }
          {
            <p title={new Date(artist.first_listen * 1000).toLocaleString()}>
              Listening since{" "}
              {new Date(artist.first_listen * 1000).toLocaleDateString()}
            </p>
          }
        </div>
      }
    >
      <div className="mt-10">
        <PeriodSelector setter={setPeriod} current={period} />
      </div>
      <div className="flex flex-col gap-20">
        <div className="flex gap-15 mt-10 flex-wrap">
          <LastPlays limit={20} artistId={artist.id} />
          <TopTracks limit={8} period={period} artistId={artist.id} />
          <ActivityGrid configurable artistId={artist.id} />
        </div>
        <ArtistAlbums period={period} artistId={artist.id} name={artist.name} />
      </div>
    </MediaLayout>
  );
}
