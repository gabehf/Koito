import { Link } from "react-router";

interface Props {
  to?: string;
  isOffset?: boolean;
  children: React.ReactNode;
}

export default function CardHeader({ isOffset, children, to }: Props) {
  const ml = isOffset ? 24 : 0;
  if (to) {
    return (
      <Link
        to={to}
        className="text-(--color-fg-secondary) hover:text-(--color-fg) inline-block sm:mb-1 group"
      >
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
      <h3
        style={{ marginLeft: ml || 0 }}
        className="color-fg-secondary inline-block sm:mb-1"
      >
        {children}
      </h3>
    );
  }
}
