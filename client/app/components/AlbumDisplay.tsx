import { imageUrl, type Album } from "api/api";
import { Link } from "react-router";

interface Props {
  album: Album;
  size: number;
}

export default function AlbumDisplay({ album, size }: Props) {
  return (
    <div className="flex gap-3" key={album.id}>
      <div>
        <Link to={`/album/${album.id}`}>
          <img
            src={imageUrl(album.image, "large")}
            alt={album.title}
            style={{ width: size }}
          />
        </Link>
      </div>
      <div className="flex flex-col items-start" style={{ width: size }}>
        <Link
          to={`/album/${album.id}`}
          className="hover:text-(--color-fg-secondary)"
        >
          <h4>{album.title}</h4>
        </Link>
        <p className="color-fg-secondary">{album.listen_count} plays</p>
      </div>
    </div>
  );
}
