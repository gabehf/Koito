import { Link } from "react-router";
import ArtistLinks from "./ArtistLinks";
import MediaItem, { MediaItemSkeleton } from "./primitives/MediaItem";
import {
  type Album,
  type Artist,
  type Track,
  type PaginatedResponse,
  type Ranked,
} from "api/api";

type Item = Album | Track | Artist;

interface Props<T extends Ranked<Item>> {
  data: PaginatedResponse<T>;
  slug: string;
  startIndex?: number;
  separators?: ConstrainBoolean;
  ranked?: boolean;
  type: "album" | "track" | "artist";
  className?: string;
  showSeeMore?: boolean;
}

export default function TopItemList<T extends Ranked<Item>>({
  data,
  slug,
  separators,
  startIndex,
  type,
  className,
  ranked,
  showSeeMore,
}: Props<T>) {
  if (startIndex) data.items.splice(0, startIndex - 1);
  return (
    <div className={`flex flex-col gap-1 ${className} min-w-[200px]`}>
      {data.items.map((item, index) => {
        const key = `${type}-${item.item.id}`;
        return (
          <div key={key} style={{ fontSize: 12 }} className="mb-0.5">
            <ItemCard
              ranked={ranked}
              rank={item.rank}
              item={item.item}
              type={type}
              key={type + item.item.id}
            />
            {separators && index !== data.items.length - 1 && (
              <div className="border-b border-(--color-bg-tertiary) mt-2" />
            )}
          </div>
        );
      })}
      {showSeeMore && data.has_next_page && (
        <div className="flex items-center w-full mt-2">
          <Link
            to={slug}
            className="text-[13px] sm:text-[15px] inline-block w-fit mx-auto text-(--color-fg-secondary) hover:text-(--color-fg) hover:cursor-pointer"
          >
            SEE MORE →
          </Link>
        </div>
      )}
    </div>
  );
}

function ItemCard({
  item,
  type,
  rank,
  ranked,
}: {
  item: Item;
  type: "album" | "track" | "artist";
  rank: number;
  ranked?: boolean;
}) {
  const itemClasses = `flex items-center gap-2`;

  switch (type) {
    case "album": {
      const album = item as Album;

      return (
        <table className="sm:text-[15px] text-[13px] border-collapse">
          <tbody>
            <tr>
              {ranked && (
                <td className="pr-3">
                  <div
                    className={`color-fg-secondary text-end ${
                      rank === 1 && "color-primary"
                    }`}
                  >
                    {rank.toString().padStart(2, "0")}
                  </div>
                </td>
              )}
              <td className="pr-3 py-1 w-full">
                <MediaItem
                  className="gap-2"
                  image={album.image}
                  link={`/album/${album.id}`}
                  size="sm"
                  title={album.title}
                  alt={album.title}
                  meta={
                    album.is_various_artists ? (
                      "Various Artists"
                    ) : (
                      <ArtistLinks artists={[album.artists[0]]} />
                    )
                  }
                  lazy
                />
              </td>
              <td className="min-w-[75px]">
                <div className="color-fg-secondary text-end">
                  {album.listen_count} plays
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      );
    }
    case "track": {
      const track = item as Track;

      return (
        <table className="sm:text-[15px] text-[13px] border-collapse">
          <tbody>
            <tr>
              {ranked && (
                <td className="pr-3">
                  <div
                    className={`color-fg-secondary text-end ${
                      rank === 1 && "color-primary"
                    }`}
                  >
                    {rank.toString().padStart(2, "0")}
                  </div>
                </td>
              )}
              <td className="pr-3 py-1 w-full">
                <MediaItem
                  className="gap-2"
                  image={track.image}
                  link={`/track/${track.id}`}
                  size="sm"
                  title={track.title}
                  alt={track.title}
                  subtitle={<ArtistLinks artists={track.artists} />}
                  lazy
                />
              </td>
              <td className="min-w-[75px]">
                <div className="color-fg-secondary text-end">
                  {track.listen_count} plays
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      );
    }
    case "artist": {
      const artist = item as Artist;

      return (
        <table className="sm:text-[15px] text-[13px] border-collapse">
          <tbody>
            <tr>
              {ranked && (
                <td className="pr-3">
                  <div
                    className={`color-fg-secondary text-end ${
                      rank === 1 && "color-primary"
                    }`}
                  >
                    {rank.toString().padStart(2, "0")}
                  </div>
                </td>
              )}
              <td className="pr-3 py-1 w-full">
                <MediaItem
                  className="gap-2"
                  image={artist.image}
                  size="sm"
                  link={`/artist/${artist.id}`}
                  title={artist.name}
                  alt={artist.name}
                  lazy
                />
              </td>
              <td className="min-w-[75px]">
                <div className="color-fg-secondary text-end">
                  {artist.listen_count} plays
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      );
    }
  }
}

interface SkeletonProps {
  count?: number;
  ranked?: boolean;
  separators?: boolean;
  type: "album" | "track" | "artist";
  className?: string;
}

export function TopItemListSkeleton({
  count = 5,
  ranked,
  type,
  className,
}: SkeletonProps) {
  return (
    <div className={`flex flex-col gap-1 ${className} min-w-[350px]`}>
      {Array.from({ length: count }).map((_, i) => (
        <div key={i} style={{ fontSize: 12 }} className="mb-3">
          <table className="sm:text-[15px] text-[13px] border-collapse">
            <tbody>
              <tr>
                {ranked && (
                  <td className="pr-3">
                    <div className="w-5 h-3 bg-secondary animate-pulse rounded-(--border-radius)" />
                  </td>
                )}
                <td className="pr-3 py-1 w-full">
                  <MediaItemSkeleton
                    size="sm"
                    className="gap-2"
                    subtitle={type === "track"}
                    meta={type === "album"}
                  />
                </td>
                <td className="min-w-[75px]">
                  <div className="flex justify-end">
                    <div className="w-14 h-3 bg-secondary animate-pulse rounded-(--border-radius)" />
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      ))}
    </div>
  );
}
