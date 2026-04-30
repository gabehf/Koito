import { Link } from "react-router";

interface Props {
  to: string;
  onClick: React.MouseEventHandler<HTMLAnchorElement>;
  img: string;
  text: string;
  subtext?: string;
}

export default function SearchResultItem(props: Props) {
  return (
    <Link
      to={props.to}
      className="px-3 py-2 flex gap-3 items-center hover:text-(--color-fg-secondary)"
      onClick={props.onClick}
    >
      <img src={props.img} alt={props.text} />
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
