import {
  apiFetch,
  type PaginatedResponse,
  type Ranked,
  type Album,
} from "api/api";
import { useQuery } from "@tanstack/react-query";
import Image from "./primitives/Image";
import CardHeader from "./primitives/CardHeader";
import ArtistLinks from "./ArtistLinks";
import MediaItem from "./primitives/MediaItem";
import { Link } from "react-router";

interface Props {
  period: string;
}

const getTopAlbums = (args: { limit: number; period: string; page: number }) =>
  apiFetch<PaginatedResponse<Ranked<Album>>>("/apis/web/v1/top-albums", args);

export default function TopAlbumsCard({ period }: Props) {
  const numItems = 5;
  const imageSize = 90;

  const args = { limit: numItems, period: period, page: 0 };
  const { isPending, isError, data, error } = useQuery({
    queryKey: ["top-albums", args],
    queryFn: () => getTopAlbums(args),
  });

  const header = "Top albums";

  if (isPending) {
    return (
      <div className="w-[300px]">
        <CardHeader to={`/chart/top-albums?period=${period}`} isOffset>
          {header}
        </CardHeader>
        <p>Loading...</p>
      </div>
    );
  } else if (isError) {
    return (
      <div className="w-[300px]">
        <CardHeader to={`/chart/top-albums?period=${period}`} isOffset>
          {header}
        </CardHeader>
        <p className="error">Error: {error.message}</p>
      </div>
    );
  }

  if (!data.items[0]) {
    return (
      <div className="w-[348px]">
        <CardHeader to={`/chart/top-albums?period=${period}`} isOffset>
          {header}
        </CardHeader>
        <p className="ml-6 mt-6">Nothing to show</p>
      </div>
    );
  }

  return (
    <div>
      <CardHeader to={`/chart/top-albums?period=${period}`} isOffset>
        {header}
      </CardHeader>
      <div className="max-w-[350px] border bg-(--color-bg-secondary) rounded-(--border-radius)">
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
              backdropFilter: "blur(0px)",
              WebkitBackdropFilter: "blur(0px)",
              maskImage: "linear-gradient(to top, black, transparent)",
              WebkitMaskImage: "linear-gradient(to top, black, transparent)",
              borderRadius: "var(--border-radius) var(--border-radius) 0 0",
            }}
          />
          <div className="absolute bottom-10 left-5">
            <Link to={`/album/${data.items[0].item.id}`}>
              <h5 className="text-3xl font-semibold">
                {data.items[0]?.item.title}
              </h5>
            </Link>
            <div>
              <ArtistLinks
                artists={
                  data.items[0]?.item.artists
                    ? [data.items[0]?.item.artists[0]]
                    : [{ id: 0, name: "Unknown Artist" }]
                }
              />
              <div className="color-fg-secondary">
                {data.items[0]?.item.listen_count} plays
              </div>
            </div>
          </div>
        </div>
        <div className="flex flex-col items-start">
          {data.items.slice(1).map((i) => (
            <div
              className="px-6 pb-6"
              key={`top_albums_card_${i.rank}_${i.item.title}`}
            >
              <MediaItem
                image={i.item.image.medium}
                imageSize={imageSize}
                link={`/album/${i.item.id}`}
                title={i.item.title}
                subtitle={<ArtistLinks artists={[i.item.artists[0]]} />}
                meta={`${i.item.listen_count} plays`}
              />
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
