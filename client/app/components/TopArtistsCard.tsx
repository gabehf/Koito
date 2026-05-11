import {
  apiFetch,
  type PaginatedResponse,
  type Ranked,
  type Artist,
} from "api/api";
import { useQuery } from "@tanstack/react-query";
import CardHeader from "./primitives/CardHeader";
import MediaItem from "./primitives/MediaItem";
import { Link } from "react-router";
import { TopCardSkeleton } from "./skeletons/TopCardSkeleton";

interface Props {
  period: string;
}

const getTopArtists = (args: { limit: number; period: string; page: number }) =>
  apiFetch<PaginatedResponse<Ranked<Artist>>>("/apis/web/v1/top-artists", args);

export default function TopArtistsCard({ period }: Props) {
  const numItems = 5;
  const imageSize = 90;

  const args = { limit: numItems, period: period, page: 0 };
  const { isPending, isError, data, error } = useQuery({
    queryKey: ["top-artists", args],
    queryFn: () => getTopArtists(args),
  });

  const header = "Top artists";

  if (isPending) {
    return <TopCardSkeleton header={header} numItems={numItems} />;
  } else if (isError) {
    return (
      <div className="w-[300px]">
        <CardHeader to={`/chart/top-artists?period=${period}`} isOffset>
          {header}
        </CardHeader>
        <p className="error">Error: {error.message}</p>
      </div>
    );
  }

  if (!data.items[0]) {
    return (
      <div className="w-[348px]">
        <CardHeader to={`/chart/top-artists?period=${period}`} isOffset>
          {header}
        </CardHeader>
        <p className="ml-6 mt-6">Nothing to show</p>
      </div>
    );
  }

  return (
    <div>
      <CardHeader to={`/chart/top-artists?period=${period}`} isOffset>
        {header}
      </CardHeader>
      <div className="max-w-[350px] card">
        <div className="relative">
          <img
            src={data.items[0]?.item.image?.large}
            style={{
              borderRadius: "var(--border-radius) var(--border-radius) 0 0",
            }}
            width={350}
            height={350}
          />
          <div
            className="absolute inset-0 bg-gradient-to-t to-50% from-(--color-bg-secondary) to-transparent"
            style={{
              borderRadius: "var(--border-radius) var(--border-radius) 0 0",
            }}
          />
          <div
            className="absolute inset-0 to-50% bg-gradient-to-t from-(--color-bg-secondary) to-transparent"
            style={{
              borderRadius: "var(--border-radius) var(--border-radius) 0 0",
            }}
          />
          <div className="absolute bottom-10 left-5">
            <Link to={`/artist/${data.items[0].item.id}`}>
              <h5 className="text-3xl font-semibold">
                {data.items[0]?.item.name}
              </h5>
            </Link>
            <div className="color-fg-secondary">
              {data.items[0]?.item.listen_count} plays
            </div>
          </div>
        </div>
        <div className="flex flex-col items-start">
          {data.items.slice(1).map((i) => (
            <div
              className="px-6 pb-6"
              key={`top_artists_card_${i.rank}_${i.item.name}`}
            >
              <MediaItem
                image={i.item.image}
                size="md"
                link={`/artist/${i.item.id}`}
                alt={i.item.name}
                title={i.item.name}
                meta={`${i.item.listen_count} plays`}
              />
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
