import { imageUrl, type RewindStats } from "api/api";
import RewindStatText from "./RewindStatText";
import { RewindTopItem } from "./RewindTopItem";

interface Props {
  stats: RewindStats;
  includeTime?: boolean;
}

export default function Rewind(props: Props) {
  const artistimg = props.stats.top_artists[0].image;
  const albumimg = props.stats.top_albums[0].image;
  const trackimg = props.stats.top_tracks[0].image;
  return (
    <div className="flex flex-col gap-7">
      <h1>{props.stats.title}</h1>
      <RewindTopItem
        title="Top Artist"
        imageSrc={imageUrl(artistimg, "medium")}
        items={props.stats.top_artists}
        getLabel={(a) => a.name}
        includeTime={props.includeTime}
      />

      <RewindTopItem
        title="Top Album"
        imageSrc={imageUrl(albumimg, "medium")}
        items={props.stats.top_albums}
        getLabel={(a) => a.title}
        includeTime={props.includeTime}
      />

      <RewindTopItem
        title="Top Track"
        imageSrc={imageUrl(trackimg, "medium")}
        items={props.stats.top_tracks}
        getLabel={(t) => t.title}
        includeTime={props.includeTime}
      />

      <div className="grid grid-cols-3 gap-y-5">
        <RewindStatText
          figure={`${props.stats.minutes_listened}`}
          text="Minutes listened"
        />
        <RewindStatText figure={`${props.stats.unique_tracks}`} text="Tracks" />
        <RewindStatText
          figure={`${props.stats.new_tracks}`}
          text="New tracks"
        />
        <RewindStatText figure={`${props.stats.plays}`} text="Plays" />
        <RewindStatText figure={`${props.stats.unique_albums}`} text="Albums" />
        <RewindStatText
          figure={`${props.stats.new_albums}`}
          text="New albums"
        />
        <RewindStatText
          figure={`${props.stats.avg_plays_per_day.toFixed(1)}`}
          text="Plays per day"
        />
        <RewindStatText
          figure={`${props.stats.unique_artists}`}
          text="Artists"
        />
        <RewindStatText
          figure={`${props.stats.new_artists}`}
          text="New artists"
        />
      </div>
    </div>
  );
}
