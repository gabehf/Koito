import { useProfile } from "~/providers/ProfileProvider";
import type { ProfileApi } from "api/api";
import { type PaginatedResponse, type Ranked, type Artist } from "api/api";
import { useQuery } from "@tanstack/react-query";
import CardHeader from "./primitives/CardHeader";
import MediaItem from "./primitives/MediaItem";

import { ProfileLink as Link } from "~/components/ProfileLink";
import { TopCardSkeleton } from "./skeletons/TopCardSkeleton";

interface Props {
  period: string;
}

const getTopArtists = (
  api: ProfileApi,
  args: { limit: number; period: string; page: number },
) =>
  api.apiFetch<PaginatedResponse<Ranked<Artist>>>(
    "/apis/web/v1/top/artists",
    args,
  );

export default function TopArtistsCard({ period }: Props) {
  const { profile, api } = useProfile();

  const numItems = 5;

  const args = { limit: numItems, period: period, page: 0 };
  const { isPending, isError, data, error } = useQuery({
    queryKey: ["profile", profile.id, "top/artists", args],
    queryFn: () => getTopArtists(api, args),
  });

  const header = "Top artists";

  if (isPending) {
    return <TopCardSkeleton header={header} numItems={numItems} />;
  } else if (isError) {
    return (
      <div className="w-[300px]">
        <CardHeader to={`/chart/top/artists?period=${period}`} isOffset>
          {header}
        </CardHeader>
        <p className="error">Error: {error.message}</p>
      </div>
    );
  }

  if (!data.items[0]) {
    return (
      <div className="w-[348px]">
        <CardHeader to={`/chart/top/artists?period=${period}`} isOffset>
          {header}
        </CardHeader>
        <p className="ml-6 mt-6">Nothing to show</p>
      </div>
    );
  }

  return (
    <div>
      <CardHeader to={`/chart/top/artists?period=${period}`} isOffset>
        {header}
      </CardHeader>
      <div className="max-w-[350px] card">
        <div className="relative">
          <img
            src={data.items[0]?.item.image?.xl}
            srcSet={`${data.items[0]?.item.image?.large} 640w, ${data.items[0]?.item.image?.xl} 1000w`}
            sizes="348px"
            style={{
              borderRadius: "var(--border-radius) var(--border-radius) 0 0",
            }}
            width={350}
            height={350}
            alt={data.items[0]?.item.name}
          />
          <div
            className="absolute inset-0 bg-linear-to-t"
            style={{
              // eased with https://larsenwork.com/easing-gradients/
              backgroundImage: `linear-gradient(to top,
              var(--color-bg-secondary) 0%,
              color-mix(in srgb, var(--color-bg-secondary) 99%, transparent) 5%,
              color-mix(in srgb, var(--color-bg-secondary) 95%, transparent) 12%,
              color-mix(in srgb, var(--color-bg-secondary) 86%, transparent) 20%,
              color-mix(in srgb, var(--color-bg-secondary) 72%, transparent) 28%,
              color-mix(in srgb, var(--color-bg-secondary) 55%, transparent) 36%,
              color-mix(in srgb, var(--color-bg-secondary) 37%, transparent) 44%,
              color-mix(in srgb, var(--color-bg-secondary) 22%, transparent) 51%,
              color-mix(in srgb, var(--color-bg-secondary) 11%, transparent) 57%,
              color-mix(in srgb, var(--color-bg-secondary) 4%, transparent) 61%,
              color-mix(in srgb, var(--color-bg-secondary) 1%, transparent) 63.5%,
              transparent 65%
              )`,
              borderRadius: "var(--border-radius) var(--border-radius) 0 0",
            }}
          />
          <div className="absolute bottom-8 left-5">
            <Link to={`/artist/${data.items[0].item.id}`}>
              <h5 className="text-3xl font-semibold line-clamp-3 wrap-anywhere text-shadow-lg">
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
