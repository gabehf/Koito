import { useState, type ReactNode } from "react";
import Image from "./Image";
import { Link } from "react-router";

interface Props {
  image: string;
  imageSize: number;
  link: string;
  title: ReactNode;
  subtitle?: ReactNode;
  meta?: ReactNode;
  className?: string;
  alignTop?: boolean;
}

export default function MediaItem({
  image,
  imageSize,
  title,
  link,
  subtitle,
  meta,
  className,
  alignTop,
}: Props) {
  return (
    <div
      className={`flex ${alignTop ? "items-start" : "items-center"} gap-3 ${
        className ?? ""
      }`}
    >
      <Link to={link} style={{ minWidth: imageSize }}>
        <Image src={image} size={imageSize} />
      </Link>
      <div
        className="flex flex-col items-start"
        style={alignTop ? { marginTop: 6 } : undefined}
      >
        <Link to={link} style={{ minWidth: imageSize }}>
          <div className="line-clamp-2 hover:text-(--color-fg-secondary)">
            {title}
          </div>
        </Link>
        {subtitle !== undefined && (
          <div className="text-[12px] sm:text-[14px]">{subtitle}</div>
        )}
        {meta !== undefined && (
          <div className="color-fg-secondary text-[12px] sm:text-[14px]">
            {meta}
          </div>
        )}
      </div>
    </div>
  );
}
