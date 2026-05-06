import { Link } from "react-router";
import ArtistLinks from "./ArtistLinks";
import MediaItem from "./primitives/MediaItem";
import {
  imageUrl,
  type Album,
  type Artist,
  type Track,
  type PaginatedResponse,
  type Ranked,
} from "api/api";

type Item = Album | Track | Artist;

interface Props<T extends Ranked<Item>> {
  data: PaginatedResponse<T>;
  startIndex?: number;
  separators?: ConstrainBoolean;
  ranked?: boolean;
  type: "album" | "track" | "artist";
  className?: string;
}

export default function TopItemList<T extends Ranked<Item>>({
  data,
  separators,
  startIndex,
  type,
  className,
  ranked,
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
        <div style={{ fontSize: 12 }} className={itemClasses}>
          {ranked && (
            <div className="w-7 color-fg-secondary text-end">{rank}</div>
          )}
          <Link to={`/album/${album.id}`}>
            <img
              loading="lazy"
              src={imageUrl(album.image, "medium")}
              alt={album.title}
              className="max-w-[56px] rounded-lg border-1 border-(--color-bg-tertiary)"
            />
          </Link>
          <div>
            <Link
              to={`/album/${album.id}`}
              className="hover:text-(--color-fg-secondary)"
            >
              <span style={{ fontSize: 14 }}>{album.title}</span>
            </Link>
            <br />
            {album.is_various_artists ? (
              <span className="color-fg-secondary">Various Artists</span>
            ) : (
              <div>
                <ArtistLinks
                  artists={
                    album.artists
                      ? [album.artists[0]]
                      : [{ id: 0, name: "Unknown Artist" }]
                  }
                />
              </div>
            )}
            <div className="color-fg-secondary">{album.listen_count} plays</div>
          </div>
        </div>
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
              <td className="pr-3 py-1 w-[250px] sm:w-[325px]">
                <MediaItem
                  className="gap-2"
                  image={imageUrl(track.image, "medium")}
                  link={`/track/${track.id}`}
                  imageSize={56}
                  title={<Link to={`/track/${track.id}`}>{track.title}</Link>}
                  subtitle={<ArtistLinks artists={track.artists} />}
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
        <div style={{ fontSize: 12 }} className={itemClasses}>
          {ranked && <div className="w-7 text-end">{rank}</div>}
          <Link
            className={
              itemClasses + " mt-1 mb-[6px] hover:text-(--color-fg-secondary)"
            }
            to={`/artist/${artist.id}`}
          >
            <img
              loading="lazy"
              src={imageUrl(artist.image, "small")}
              alt={artist.name}
              className="min-w-[48px]"
            />
            <div>
              <span style={{ fontSize: 14 }}>{artist.name}</span>
              <div className="color-fg-secondary">
                {artist.listen_count} plays
              </div>
            </div>
          </Link>
        </div>
      );
    }
  }
}
