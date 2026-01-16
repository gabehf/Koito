import { useQuery } from "@tanstack/react-query";
import { getInterest, type getInterestArgs } from "api/api";
import { useTheme } from "~/hooks/useTheme";
import type { Theme } from "~/styles/themes.css";
import { Area, AreaChart } from "recharts";
import { RechartsDevtools } from "@recharts/devtools";

function getPrimaryColor(theme: Theme): string {
  const value = theme.primary;
  const rgbMatch = value.match(
    /^rgb\(\s*(\d{1,3})\s*,\s*(\d{1,3})\s*,\s*(\d{1,3})\s*\)$/
  );
  if (rgbMatch) {
    const [, r, g, b] = rgbMatch.map(Number);
    return "#" + [r, g, b].map((n) => n.toString(16).padStart(2, "0")).join("");
  }

  return value;
}
interface Props {
  buckets?: number;
  artistId?: number;
  albumId?: number;
  trackId?: number;
}

export default function InterestGraph({
  buckets = 16,
  artistId = 0,
  albumId = 0,
  trackId = 0,
}: Props) {
  const { isPending, isError, data, error } = useQuery({
    queryKey: [
      "interest",
      {
        buckets: buckets,
        artist_id: artistId,
        album_id: albumId,
        track_id: trackId,
      },
    ],
    queryFn: ({ queryKey }) => getInterest(queryKey[1] as getInterestArgs),
  });

  const { theme } = useTheme();
  const color = getPrimaryColor(theme);

  if (isPending) {
    return (
      <div className="w-[350px] sm:w-[500px]">
        <h3>Interest over time</h3>
        <p>Loading...</p>
      </div>
    );
  } else if (isError) {
    return (
      <div className="w-[350px] sm:w-[500px]">
        <h3>Interest over time</h3>
        <p className="error">Error: {error.message}</p>
      </div>
    );
  }

  // Note: I would really like to have the animation for the graph, however
  // the line graph can get weirdly clipped before the animation is done
  // so I think I just have to remove it for now.

  return (
    <div className="flex flex-col items-start w-full max-w-[335px] sm:max-w-[500px]">
      <h3>Interest over time</h3>
      <AreaChart
        style={{
          width: "100%",
          aspectRatio: 3.5,
          maxWidth: 440,
          overflow: "visible",
        }}
        data={data}
        margin={{ top: 15, bottom: 20 }}
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
  );
}
