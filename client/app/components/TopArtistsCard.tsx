import {
  apiFetch,
  imageUrl,
  type PaginatedResponse,
  type Ranked,
  type Artist,
} from "api/api";
import { useQuery } from "@tanstack/react-query";

interface Props {
  period: string;
}

const getTopArtists = (args: { limit: number; period: string; page: number }) =>
  apiFetch<PaginatedResponse<Ranked<Artist>>>("/apis/web/v1/top-artists", args);

export default function TopArtistsCard({ period }: Props) {
  const args = { limit: 3, period: period, page: 0 };
  const { isPending, isError, data, error } = useQuery({
    queryKey: ["top-artists", args],
    queryFn: () => getTopArtists(args),
  });

  const header = "Top artists";

  if (isPending) {
    return (
      <div className="w-[300px]">
        <h3>{header}</h3>
        <p>Loading...</p>
      </div>
    );
  } else if (isError) {
    return (
      <div className="w-[300px]">
        <h3>{header}</h3>
        <p className="error">Error: {error.message}</p>
      </div>
    );
  }

  return (
    <div>
      <h3 className="ml-6">{header}</h3>
      <div className="max-w-[350px] border bg-(--color-bg-secondary) rounded-(--border-radius)">
        <div className="relative">
          <img
            src={imageUrl(data.items[0].item.image, "large")}
            style={{
              borderRadius: "var(--border-radius) var(--border-radius) 0 0",
            }}
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
              backdropFilter: "blur(6px)",
              WebkitBackdropFilter: "blur(6px)",
              maskImage: "linear-gradient(to top, black, transparent)",
              WebkitMaskImage: "linear-gradient(to top, black, transparent)",
              borderRadius: "var(--border-radius) var(--border-radius) 0 0",
            }}
          />
          <div className="absolute bottom-10 left-5">
            <h2 className="font-medium text-sm">{data.items[0].item.name}</h2>
            <div className="color-fg-secondary">
              {data.items[0].item.listen_count} plays
            </div>
          </div>
        </div>
        <div className="flex flex-col items-start">
          <div className="pl-6 pb-6">
            <div className="flex gap-3 items-center">
              <img
                src={imageUrl(data.items[1].item.image, "medium")}
                className="rounded-[10px] w-[125px]"
              />
              <div className="flex flex-col items-start">
                <h4 className="font-normal">{data.items[1].item.name}</h4>
                <p className="color-fg-secondary">
                  {data.items[1].item.listen_count} plays
                </p>
              </div>
            </div>
          </div>
          <div className="pl-6 pb-6">
            <div className="flex gap-3 items-center">
              <img
                src={imageUrl(data.items[2].item.image, "medium")}
                className="rounded-[10px] w-[125px]"
              />
              <div className="flex flex-col items-start">
                <h4 className="font-normal">{data.items[2].item.name}</h4>
                <p className="color-fg-secondary">
                  {data.items[2].item.listen_count} plays
                </p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
