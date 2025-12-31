import { imageUrl, type RewindStats } from "api/api";
import RewindTopItem from "./RewindTopItem";

interface Props {
  stats: RewindStats;
  includeTime: boolean;
}

export default function Rewind(props: Props) {
  const artistimg = props.stats.top_artists[0].image;
  const albumimg = props.stats.top_albums[0].image;
  const trackimg = props.stats.top_tracks[0].image;
  return (
    <div className="flex flex-col gap-10">
      <h1>{props.stats.title}</h1>
      <div className="flex gap-5">
        <div className="rewind-top-item-image">
          <img className="w-58 h-58" src={imageUrl(artistimg, "medium")} />
        </div>
        <div className="flex flex-col gap-1">
          <h4>Top Artist</h4>
          <div className="flex items-center gap-2">
            <div className="flex flex-col items-start mb-3">
              <h2>{props.stats.top_artists[0].name}</h2>
              <span className="text-(--color-fg-tertiary) -mt-3">
                {`${props.stats.top_artists[0].listen_count} plays`}
                {props.includeTime
                  ? ` (${Math.floor(
                      props.stats.top_artists[0].time_listened / 60
                    )} minutes)`
                  : ``}
              </span>
            </div>
          </div>
          {props.stats.top_artists.slice(1).map((e, i) => (
            <div className="" key={e.id}>
              {e.name}
              <span className="text-(--color-fg-tertiary)">
                {` - ${e.listen_count} plays`}
                {props.includeTime
                  ? ` (${Math.floor(e.time_listened / 60)} minutes)`
                  : ``}
              </span>
            </div>
          ))}
        </div>
      </div>
      <div className="flex gap-5">
        <div className="rewind-top-item-image">
          <img className="w-58 h-58" src={imageUrl(albumimg, "medium")} />
        </div>
        <div className="flex flex-col gap-1">
          <h4>Top Album</h4>
          <div className="flex items-center gap-2">
            <div className="flex flex-col items-start mb-3">
              <h2>{props.stats.top_albums[0].title}</h2>
              <span className="text-(--color-fg-tertiary) -mt-3">
                {`${props.stats.top_albums[0].listen_count} plays`}
                {props.includeTime
                  ? ` (${Math.floor(
                      props.stats.top_albums[0].time_listened / 60
                    )} minutes)`
                  : ``}
              </span>
            </div>
          </div>
          {props.stats.top_albums.slice(1).map((e, i) => (
            <div className="" key={e.id}>
              {e.title}
              <span className="text-(--color-fg-tertiary)">
                {` - ${e.listen_count} plays`}
                {props.includeTime
                  ? ` (${Math.floor(e.time_listened / 60)} minutes)`
                  : ``}
              </span>
            </div>
          ))}
        </div>
      </div>
      <div className="flex gap-5">
        <div className="rewind-top-item-image">
          <img className="w-58 h-58" src={imageUrl(trackimg, "medium")} />
        </div>
        <div className="flex flex-col gap-1">
          <h4>Top Track</h4>
          <div className="flex items-center gap-2">
            <div className="flex flex-col items-start mb-3">
              <h2>{props.stats.top_tracks[0].title}</h2>
              <span className="text-(--color-fg-tertiary) -mt-3">
                {`${props.stats.top_tracks[0].listen_count} plays`}
                {props.includeTime
                  ? ` (${Math.floor(
                      props.stats.top_tracks[0].time_listened / 60
                    )} minutes)`
                  : ``}
              </span>
            </div>
          </div>
          {props.stats.top_tracks.slice(1).map((e, i) => (
            <div className="" key={e.id}>
              {e.title}
              <span className="text-(--color-fg-tertiary)">
                {` - ${e.listen_count} plays`}
                {props.includeTime
                  ? ` (${Math.floor(e.time_listened / 60)} minutes)`
                  : ``}
              </span>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
