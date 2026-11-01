import { ProfileLink as Link } from "~/components/ProfileLink";
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
      <div className="relative">
        <img
          src={props.img}
          style={{
            borderRadius: "var(--border-radius)",
          }}
          width={150}
          height={150}
        />
        <div
          className="absolute inset-0 border"
          style={{
            // eased with https://larsenwork.com/easing-gradients/
            backgroundImage: `linear-gradient(to top,
            var(--color-bg-secondary) 0%,
            color-mix(in srgb, var(--color-bg-secondary) 99%, transparent) 5%,
            color-mix(in srgb, var(--color-bg-secondary) 95%, transparent) 12%,
            color-mix(in srgb, var(--color-bg-secondary) 86%, transparent) 20%,
            color-mix(in srgb, var(--color-bg-secondary) 72%, transparent) 28%,
            color-mix(in srgb, var(--color-bg-secondary) 55%, transparent) 36%,
            color-mix(in srgb, var(--color-bg-secondary) 37%, transparent) 44%,
            color-mix(in srgb, var(--color-bg-secondary) 22%, transparent) 51%,
            color-mix(in srgb, var(--color-bg-secondary) 11%, transparent) 57%,
            color-mix(in srgb, var(--color-bg-secondary) 4%, transparent) 61%,
            color-mix(in srgb, var(--color-bg-secondary) 1%, transparent) 63.5%,
            transparent 65%
            )`,
            borderRadius: "var(--border-radius)",
          }}
        />
        <div className="absolute bottom-3 left-3 pr-2">
          <h5 className="text-xl font-semibold line-clamp-3 wrap-anywhere text-shadow-lg">
            {props.text}
          </h5>
        </div>
      </div>
    </Link>
  );
}

export { SearchResultArtistItem, SearchResultItem };
