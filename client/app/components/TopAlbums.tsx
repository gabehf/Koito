import { useProfile } from "~/providers/ProfileProvider";
import type { ProfileApi } from "api/api";
import { useQuery } from "@tanstack/react-query";
import { type PaginatedResponse, type Ranked, type Album } from "api/api";

import { ProfileLink as Link } from "~/components/ProfileLink";
import TopItemList from "./TopItemList";

interface Props {
  limit: number;
  period: string;
  artistId?: Number;
}

const getTopAlbums = (
  api: ProfileApi,
  args: {
    limit: number;
    period: string;
    artist_id?: number;
    page: number;
  },
) =>
  api.apiFetch<PaginatedResponse<Ranked<Album>>>(
    "/apis/web/v1/top/albums",
    args,
  );

export default function TopAlbums(props: Props) {
  const { profile, api } = useProfile();

  const args = {
    limit: props.limit,
    period: props.period,
    artist_id: props.artistId as number | undefined,
    page: 0,
  };
  const { isPending, isError, data, error } = useQuery({
    queryKey: ["profile", profile.id, "top/albums", args],
    queryFn: () => getTopAlbums(api, args),
  });

  const header = "Top albums";

  const slug = `/chart/top/albums?period=${props.period}&artist_id=${props.artistId}&limit=${props.limit}`;

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
      <h3 className="hover:underline">
        <Link
          to={`/chart/top/albums?period=${props.period}${
            props.artistId ? `&artist_id=${props.artistId}` : ""
          }`}
        >
          {header}
        </Link>
      </h3>
      <div className="max-w-[300px]">
        <TopItemList type="album" data={data} slug={slug} />
        {data.items.length < 1 ? "Nothing to show" : ""}
      </div>
    </div>
  );
}
