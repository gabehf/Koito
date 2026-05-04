import {
  apiFetch,
  imageUrl,
  type PaginatedResponse,
  type Ranked,
  type Album,
} from "api/api";
import { useQuery } from "@tanstack/react-query";
import Image from "./primitives/Image";

interface Props {
  period: string;
}

const getTopAlbums = (args: { limit: number; period: string; page: number }) =>
  apiFetch<PaginatedResponse<Ranked<Album>>>("/apis/web/v1/top-albums", args);

export default function TopAlbumsCard({ period }: Props) {
  const args = { limit: 3, period: period, page: 0 };
  const { isPending, isError, data, error } = useQuery({
    queryKey: ["top-albums", args],
    queryFn: () => getTopAlbums(args),
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
            <h2 className="font-medium text-sm">{data.items[0].item.title}</h2>
            <div className="color-fg-secondary">
              {data.items[0].item.listen_count} plays
            </div>
          </div>
        </div>
        <div className="flex flex-col items-start">
          <div className="pl-6 pb-6">
            <div className="flex gap-3 items-center">
              <Image
                src={imageUrl(data.items[1].item.image, "medium")}
                size={125}
              />
              <div className="flex flex-col items-start">
                <h4 className="font-normal">{data.items[1].item.title}</h4>
                <p className="color-fg-secondary">
                  {data.items[1].item.listen_count} plays
                </p>
              </div>
            </div>
          </div>
          <div className="pl-6 pb-6">
            <div className="flex gap-3 items-center">
              <Image
                src={imageUrl(data.items[2].item.image, "medium")}
                size={125}
              />
              <div className="flex flex-col items-start">
                <h4 className="font-normal">{data.items[2].item.title}</h4>
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
