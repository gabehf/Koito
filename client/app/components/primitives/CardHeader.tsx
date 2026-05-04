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
        className="text-(--color-fg-secondary) hover:text-(--color-fg)"
      >
        <h3 style={{ marginLeft: ml || 0 }} className="hover:cursor-pointer">
          {children}
        </h3>
      </Link>
    );
  } else {
    return (
      <h3 style={{ marginLeft: ml || 0 }} className="color-fg-secondary">
        {children}
      </h3>
    );
  }
}
