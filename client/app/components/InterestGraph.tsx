import { useQuery } from "@tanstack/react-query";
import { apiFetch, type InterestBucket } from "api/api";
import { useTheme } from "~/hooks/useTheme";
import type { Theme } from "~/styles/themes.css";
import { Area, AreaChart, Label, XAxis } from "recharts";
import CardHeader from "./primitives/CardHeader";

interface Props {
  buckets?: number;
  artistId?: number;
  albumId?: number;
  trackId?: number;
}

const getInterest = (args: {
  buckets: number;
  artist_id: number;
  album_id: number;
  track_id: number;
}) => apiFetch<InterestBucket[]>("/apis/web/v1/interest", args);

export default function InterestGraph({
  buckets = 16,
  artistId = 0,
  albumId = 0,
  trackId = 0,
}: Props) {
  const args = {
    buckets,
    artist_id: artistId,
    album_id: albumId,
    track_id: trackId,
  };
  const { isPending, isError, data, error } = useQuery({
    queryKey: ["interest", args],
    queryFn: () => getInterest(args),
  });

  const { theme } = useTheme();
  const color = theme.primary;

  const title = "Interest over time";

  if (isPending) {
    return (
      <div className="w-[350px] sm:w-[500px]">
        <CardHeader>{title}</CardHeader>
        <p>Loading...</p>
      </div>
    );
  } else if (isError) {
    return (
      <div className="w-[350px] sm:w-[550px]">
        <CardHeader>{title}</CardHeader>
        <p className="error">Error: {error.message}</p>
      </div>
    );
  }

  // Note: I would really like to have the animation for the graph, however
  // the line graph can get weirdly clipped before the animation is done
  // so I think I just have to remove it for now.

  return (
    <div className="flex flex-col items-start">
      <CardHeader isOffset>{title}</CardHeader>
      <div className="flex flex-col items-center w-[350px] sm:w-[550px] text-[12px] p-6 card">
        <AreaChart
          style={{
            width: "100%",
            maxWidth: 450,
            overflow: "visible",
            height: "120px",
          }}
          data={data}
          margin={{ top: 20, bottom: 15 }}
        >
          <defs>
            <linearGradient id="colorGradient" x1="0" y1="0" x2="0" y2="1">
              <stop offset="5%" stopColor={color} stopOpacity={0.5} />
              <stop offset="95%" stopColor={color} stopOpacity={0} />
            </linearGradient>
          </defs>
          <Area
            dataKey="listen_count"
            type="natural"
            stroke="none"
            fill="url(#colorGradient)"
            animationDuration={0}
            animationEasing="ease-in-out"
            activeDot={false}
          />
          <Area
            dataKey="listen_count"
            type="natural"
            stroke={color}
            fill="none"
            strokeWidth={2}
            animationDuration={0}
            animationEasing="ease-in-out"
            dot={false}
            activeDot={false}
            style={{ filter: `drop-shadow(0px 0px 0px ${color})` }}
          />
        </AreaChart>
      </div>
    </div>
  );
}
