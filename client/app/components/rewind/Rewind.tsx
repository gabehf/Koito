import { type RewindStats } from "api/api";
import RewindStatText from "./RewindStatText";
import { RewindTopItem } from "./RewindTopItem";

interface Props {
  stats: RewindStats;
  includeTime?: boolean;
}

export default function Rewind(props: Props) {
  const artistimg = props.stats.top_artists[0]?.item.image;
  const albumimg = props.stats.top_albums[0]?.item.image;
  const trackimg = props.stats.top_tracks[0]?.item.image;
  if (
    !props.stats.top_artists[0] ||
    !props.stats.top_albums[0] ||
    !props.stats.top_tracks[0]
  ) {
    return <p>Not enough data exists to create a Rewind for this period :(</p>;
  }
  return (
    <div className="flex flex-col gap-7 card pt-4 sm:pt-6 pb-8 sm:pb-12 px-4 sm:px-8 min-w-[350px]">
      <div className="w-full text-start shrink-0 sm:ml-6">
        <span
          className="
            relative inline-block
            text-2xl md:text-4xl font-semibold
          "
        >
          <span
            className="
              absolute inset-0
              sm:-translate-x-6 translate-y-9 sm:translate-y-11
              bg-(--color-primary)
              z-0
              h-0.5
            "
            aria-hidden
          />
          <h5 className="mb-2">{props.stats.title}</h5>
        </span>
      </div>
      <RewindTopItem
        title="Top Artist"
        imageSrc={artistimg?.medium}
        items={props.stats.top_artists}
        getLabel={(a) => a.name}
        includeTime={props.includeTime}
      />

      <RewindTopItem
        title="Top Album"
        imageSrc={albumimg?.medium}
        items={props.stats.top_albums}
        getLabel={(a) => a.title}
        includeTime={props.includeTime}
      />

      <RewindTopItem
        title="Top Track"
        imageSrc={trackimg?.medium}
        items={props.stats.top_tracks}
        getLabel={(t) => t.title}
        includeTime={props.includeTime}
      />

      <div className="grid grid-cols-1 sm:grid-cols-3 gap-y-5">
        <RewindStatText
          figure={`${props.stats.minutes_listened}`}
          text="Minutes"
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
          text="Plays / day"
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
