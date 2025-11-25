import { useQuery } from "@tanstack/react-query";
import { getStats, type Stats, type ApiError } from "api/api";

export default function AllTimeStats() {
  const { isPending, isError, data, error } = useQuery({
    queryKey: ["stats", "all_time"],
    queryFn: ({ queryKey }) => getStats(queryKey[1]),
  });

  if (isPending) {
    return (
      <div className="w-[200px]">
        <h2>All Time Stats</h2>
        <p>Loading...</p>
      </div>
    );
  } else if (isError) {
    return (
      <>
        <div>
          <h2>All Time Stats</h2>
          <p className="error">Error: {error.message}</p>
        </div>
      </>
    );
  }

  const numberClasses = "header-font font-bold text-xl";

  return (
    <div>
      <h2>All Time Stats</h2>
      <div>
        <span
          className={numberClasses}
          title={Math.floor(data.minutes_listened / 60) + " hours"}
        >
          {data.minutes_listened}
        </span>{" "}
        Minutes Listened
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
