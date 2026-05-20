import { useQuery } from "@tanstack/react-query";
import { apiFetch, type Stats } from "api/api";
import CardHeader from "./primitives/CardHeader";

const getStats = (period: string) =>
  apiFetch<Stats>("/apis/web/v1/stats", { period });

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
      <div className="mt-6">
        <span
          className={numberClasses}
          title={Math.floor(data.minutes_listened / 60) + " hours"}
        >
          {data.minutes_listened}
        </span>{" "}
        Minutes
      </div>
      <div>
        <span className={numberClasses}>{data.listen_count}</span> Plays
      </div>
      <div>
        <span className={numberClasses}>{data.track_count}</span> Tracks
      </div>
      <div>
        <span className={numberClasses}>{data.album_count}</span> Albums
      </div>
      <div>
        <span className={numberClasses}>{data.artist_count}</span> Artists
      </div>
    </div>
  );
}

export function AllTimeStatsSkeleton() {
  const barWidths = [80, 60, 70, 50, 60];
  return (
    <div>
      <CardHeader>All time stats</CardHeader>
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
  );
}
