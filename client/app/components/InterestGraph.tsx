import { useQuery } from "@tanstack/react-query";
import { apiFetch, type InterestBucket } from "api/api";
import { useTheme } from "~/hooks/useTheme";
import { Area, AreaChart } from "recharts";
import CardHeader from "./primitives/CardHeader";

interface Props {
  buckets?: number;
  type: string;
  id: number;
}

const getInterest = (args: { buckets: number; type: string; id: number }) =>
  apiFetch<InterestBucket[]>(
    `/apis/web/v1/${args.type.toLowerCase()}/${args.id}/interest`,
    args,
  );

export default function InterestGraph({ buckets = 16, type, id }: Props) {
  const args = {
    buckets,
    type: type,
    id: id,
  };
  const { isPending, isError, data, error } = useQuery({
    queryKey: ["interest", args],
    queryFn: () => getInterest(args),
  });

  const { theme } = useTheme();
  const color = theme.primary;

  const title = "Interest over time";

  if (isPending) {
    return <InterestGraphSkeleton />;
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
      <div className="flex flex-col items-center w-[350px] sm:w-[514px] md:w-[550px] h-[175px] text-[12px] p-6 card">
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

export function InterestGraphSkeleton() {
  return (
    <div className="flex flex-col items-start">
      <CardHeader isOffset>Interest over time</CardHeader>
      <div className="flex flex-col items-center w-[350px] sm:w-[550px] h-[175px] text-[12px] p-6 card">
        <div className="w-full max-w-[450px] h-[120px] relative overflow-hidden rounded-(--border-radius) flex justify-around items-center">
          <div className="w-full h-3/4 bg rounded-(--border-radius) animate-pulse" />
        </div>
      </div>
    </div>
  );
}
