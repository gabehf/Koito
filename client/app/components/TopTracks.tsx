import { useProfile } from "~/providers/ProfileProvider";
import type { ProfileApi } from "api/api";
import { useQuery } from "@tanstack/react-query";
import { type PaginatedResponse, type Ranked, type Track } from "api/api";
import { ProfileLink as Link } from "~/components/ProfileLink";
import TopItemList, { TopItemListSkeleton } from "./TopItemList";
import CardHeader from "./primitives/CardHeader";

interface Props {
  limit: number;
  period: string;
  className?: string;
  showSeeMore?: boolean;
  artistId?: Number;
  albumId?: Number;
}

const getTopTracks = (
  api: ProfileApi,
  args: {
    limit: number;
    period: string;
    artist_id?: number;
    album_id?: number;
    page: number;
  },
) =>
  api.apiFetch<PaginatedResponse<Ranked<Track>>>(
    "/apis/web/v1/top/tracks",
    args,
  );

const TopTracks = (props: Props) => {
  const { profile, api } = useProfile();
  const args = {
    limit: props.limit,
    period: props.period,
    artist_id: props.artistId as number | undefined,
    album_id: props.albumId as number | undefined,
    page: 0,
  };
  const { isPending, isError, data, error } = useQuery({
    queryKey: ["profile", profile.id, "top/tracks", args],
    queryFn: () => getTopTracks(api, args),
  });

  const header = "Top tracks";

  if (isPending) {
    return (
      <div className={props.className}>
        <CardHeader isOffset>{header}</CardHeader>
        <div className="mt-7">
          <TopItemListSkeleton count={props.limit} ranked type="track" />
        </div>
      </div>
    );
  } else if (isError) {
    return (
      <div className={props.className}>
        <CardHeader isOffset>{header}</CardHeader>
        <p className="error">Error: {error.message}</p>
      </div>
    );
  }
  if (!data.items) return;

  let params = "";
  params += props.artistId ? `&artist_id=${props.artistId}` : "";
  params += props.albumId ? `&album_id=${props.albumId}` : "";

  const slug = `/chart/top/tracks?period=${props.period}&limit=100${params}`;

  if (!data.items[0]) {
    return (
      <div className={props.className}>
        <CardHeader
          isOffset
          to={`/chart/top/tracks?period=${props.period}${params}`}
        >
          {header}
        </CardHeader>
        <p className="ml-6 mt-6">Nothing to show</p>
      </div>
    );
  }

  return (
    <div className={props.className}>
      <CardHeader
        isOffset
        to={`/chart/top/tracks?period=${props.period}${params}`}
      >
        {header}
      </CardHeader>
      <div className="mt-7">
        <TopItemList
          ranked
          type="track"
          data={data}
          separators
          slug={slug}
          showSeeMore={props.showSeeMore}
        />
      </div>
    </div>
  );
};

export default TopTracks;
