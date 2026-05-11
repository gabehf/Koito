import { useQuery } from "@tanstack/react-query";
import {
  apiFetch,
  type PaginatedResponse,
  type Ranked,
  type Album,
} from "api/api";
import MediaItem, { MediaItemSkeleton } from "./primitives/MediaItem";

const getArtistAlbums = (artistId: number) =>
  apiFetch<PaginatedResponse<Ranked<Album>>>("/apis/web/v1/top-albums", {
    period: "all_time",
    limit: 99,
    artist_id: artistId,
  });

interface Props {
  artistId: number;
  header: string;
}

export default function ArtistAlbums({ artistId, header }: Props) {
  const { isPending, isError, data, error } = useQuery({
    queryKey: ["artist-albums", artistId],
    queryFn: () => getArtistAlbums(artistId),
  });

  if (isPending) {
    return <ArtistAlbumsSkeleton />;
  }
  if (isError) {
    return (
      <div>
        <h3>{header}</h3>
        <p className="error">Error:{error.message}</p>
      </div>
    );
  }

  return (
    <div>
      <h3 className="mb-6">{header}</h3>
      <div className="flex flex-wrap gap-8">
        {data.items.length < 1 && "Nothing to show"}
        {data.items.map((item) => (
          <div className="w-[330px]" key={item.item.id}>
            <MediaItem
              image={item.item.image}
              size="lg"
              link={`/album/${item.item.id}`}
              alt={item.item.title}
              alignTop
              title={item.item.title}
              meta={`${item.item.listen_count} plays`}
            />
          </div>
        ))}
      </div>
    </div>
  );
}

function ArtistAlbumsSkeleton() {
  return (
    <div>
      <h3 className="mb-6">Albums featuring</h3>
      <div className="flex flex-wrap gap-8">
        {[1, 2, 3, 4, 5].map((i) => (
          <div className="w-[330px]" key={`artist_album_skeleton_${i}`}>
            <MediaItemSkeleton subtitle alignTop size="lg" />
          </div>
        ))}
      </div>
    </div>
  );
}
