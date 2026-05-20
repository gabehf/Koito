import { Link } from "react-router";
import Image from "./primitives/Image";

interface Props {
  to: string;
  onClick: React.MouseEventHandler<HTMLAnchorElement>;
  img: string;
  imgSize?: number;
  text: string;
  subtext?: string;
}

function SearchResultItem(props: Props) {
  return (
    <Link
      to={props.to}
      className="px-3 py-2 flex gap-3 items-center hover:text-(--color-fg-secondary)"
      onClick={props.onClick}
    >
      <Image
        src={props.img}
        size={props.imgSize ? props.imgSize : 100}
        alt={props.text}
      />
      <div>
        {props.text}
        {props.subtext ? (
          <>
            <br />
            <span className="color-fg-secondary">{props.subtext}</span>
          </>
        ) : (
          ""
        )}
      </div>
    </Link>
  );
}

function SearchResultArtistItem(props: Props) {
  return (
    <Link
      to={props.to}
      className="flex gap-3 items-center w-fit"
      onClick={props.onClick}
    >
      <div className="relative border rounded-(--border-radius)">
        <img
          src={props.img}
          style={{
            borderRadius: "var(--border-radius)",
          }}
          width={150}
          height={150}
        />
        <div
          className="absolute inset-0 bg-gradient-to-t to-50% from-(--color-bg) to-transparent"
          style={{
            borderRadius: "var(--border-radius",
          }}
        />
        <div
          className="absolute inset-0 to-50% bg-gradient-to-t from-(--color-bg) to-transparent"
          style={{
            backdropFilter: "blur(0px)",
            WebkitBackdropFilter: "blur(0px)",
            maskImage: "linear-gradient(to top, black, transparent)",
            WebkitMaskImage: "linear-gradient(to top, black, transparent)",
            borderRadius: "var(--border-radius)",
          }}
        />
        <div className="absolute bottom-3 left-3">
          <h5 className="text-xl font-semibold line-clamp-2">{props.text}</h5>
        </div>
      </div>
    </Link>
  );
}

export { SearchResultArtistItem, SearchResultItem };
