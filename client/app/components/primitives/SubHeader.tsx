import { Link } from "react-router";

interface Props {
  to?: string;
  isOffset?: boolean;
  children: React.ReactNode;
}

export default function SubHeader({ isOffset, children, to }: Props) {
  const ml = isOffset ? 24 : 0;

  const classNames = "text-(--color-fg-secondary) inline-block sm:mb-3 mb-2";

  if (to) {
    return (
      <Link to={to} className={`${classNames} hover:text-(--color-fg) group`}>
        <h3 style={{ marginLeft: ml || 0 }} className="hover:cursor-pointer">
          {children}
          <span className="opacity-0 group-hover:opacity-100 transition-opacity">
            {to ? " →" : ""}
          </span>
        </h3>
      </Link>
    );
  } else {
    return (
      <h3 style={{ marginLeft: ml || 0 }} className={classNames}>
        {children}
      </h3>
    );
  }
}
