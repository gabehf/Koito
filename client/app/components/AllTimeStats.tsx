import { useQuery } from "@tanstack/react-query";
import { apiFetch, type Stats } from "api/api";
import CardHeader from "./primitives/CardHeader";

const getStats = (period: string) =>
  apiFetch<Stats>("/apis/web/v1/stats", { period });

const numberFormatter = new Intl.NumberFormat();

export default function AllTimeStats() {
  const { isPending, isError, data, error } = useQuery({
    queryKey: ["stats", "all_time"],
    queryFn: () => getStats("all_time"),
  });

  const header = "All time stats";

  if (isPending) {
    return <AllTimeStatsSkeleton />;
  } else if (isError) {
    return (
      <>
        <div>
          <h3>{header}</h3>
          <p className="error">Error: {error.message}</p>
        </div>
      </>
    );
  }

  const numberClasses = "header-font font-bold text-xl";

  return (
    <div>
      <CardHeader>{header}</CardHeader>
      <div className="flex gap-6">
        <div>
          <div className="mt-6">
            <span
              className={numberClasses}
              title={Math.floor(data.minutes_listened / 60) + " hours"}
            >
              {numberFormatter.format(data.minutes_listened)}
            </span>{" "}
            Minutes
          </div>
          <div>
            <span className={numberClasses}>
              {numberFormatter.format(data.listen_count)}
            </span>{" "}
            Plays
          </div>
          <div>
            <span className={numberClasses}>
              {numberFormatter.format(data.track_count)}
            </span>{" "}
            Tracks
          </div>
          <div>
            <span className={numberClasses}>
              {numberFormatter.format(data.album_count)}
            </span>{" "}
            Albums
          </div>
          <div>
            <span className={numberClasses}>
              {numberFormatter.format(data.artist_count)}
            </span>{" "}
            Artists
          </div>
        </div>
        <div>
          <div className="mt-6">
            <span className={numberClasses}>{data.days_active}</span> Days
            Active
          </div>
          <div>
            <span className={numberClasses}>{data.longest_streak}</span> Longest
            Streak
          </div>
          <div>
            <span className={numberClasses}>
              {data.avg_daily_plays.toFixed(1)}
            </span>{" "}
            Avg. Daily Plays
          </div>
          <div>
            <span className={numberClasses}>
              {data.tracks_per_artist.toFixed(1)}
            </span>{" "}
            Avg. Tracks / Artist
          </div>
          <div>
            <span className={numberClasses}>
              {data.albums_per_artist.toFixed(1)}
            </span>{" "}
            Avg. Albums / Artist
          </div>
        </div>
      </div>
    </div>
  );
}

export function AllTimeStatsSkeleton() {
  const barWidths = [80, 60, 70, 50, 60];
  return (
    <div>
      <CardHeader>All time stats</CardHeader>
      <div className="flex gap-6">
        <div className="mt-6 flex flex-col gap-2">
          {barWidths.map((w, i) => (
            <div key={i} className="flex items-center gap-2">
              <div
                className="h-5 bg-secondary animate-pulse rounded-(--border-radius)"
                style={{ width: 40 }}
              />
              <div
                className="h-3 bg-secondary animate-pulse rounded-(--border-radius)"
                style={{ width: w }}
              />
            </div>
          ))}
        </div>
        <div className="mt-6 flex flex-col gap-2">
          {barWidths.map((w, i) => (
            <div key={i} className="flex items-center gap-2">
              <div
                className="h-5 bg-secondary animate-pulse rounded-(--border-radius)"
                style={{ width: 40 }}
              />
              <div
                className="h-3 bg-secondary animate-pulse rounded-(--border-radius)"
                style={{ width: w }}
              />
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
