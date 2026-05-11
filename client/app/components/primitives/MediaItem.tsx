import { useState, type ReactNode } from "react";
import Image from "./Image";
import { Link } from "react-router";
import type { ImageList } from "api/api";

interface Props {
  image: ImageList;
  size: "sm" | "md" | "lg";
  link: string;
  title: ReactNode;
  alt: string;
  subtitle?: ReactNode;
  meta?: ReactNode;
  className?: string;
  alignTop?: boolean;
  lazy?: boolean;
}

const sizeToPx = (size: "sm" | "md" | "lg") => {
  switch (size) {
    case "sm":
      return 56;
    case "md":
      return 90;
    case "lg":
      return 125;
  }
};

const sizeToImage = (
  size: "sm" | "md" | "lg",
  img: ImageList,
  alt: string,
  lazy: boolean
): ReactNode => {
  const px = sizeToPx(size); // 56, 90, 125

  let srcset = "";
  let sizes = "";

  switch (size) {
    case "sm":
      srcset = `${img.xs} 64w, ${img.small} 128w, ${img.medium} 300w`;
      sizes = "56px";
      break;
    case "md":
      srcset = `${img.small} 128w, ${img.medium} 300w`;
      sizes = "90px";
      break;
    case "lg":
      srcset = `${img.small} 128w, ${img.medium} 300w, ${img.large} 640w`;
      sizes = "123px";
      break;
  }

  return (
    <Image
      src={img.medium}
      size={px}
      lazy={lazy}
      alt={alt}
      srcset={srcset}
      sizes={sizes}
    />
  );
};

export default function MediaItem({
  image,
  size,
  title,
  link,
  subtitle,
  meta,
  alt,
  className,
  alignTop,
  lazy,
}: Props) {
  return (
    <div
      className={`flex ${alignTop ? "items-start" : "items-center"} gap-3 ${
        className ?? ""
      }`}
    >
      <Link to={link} style={{ minWidth: sizeToPx(size) }}>
        {sizeToImage(size, image, alt, lazy || false)}
      </Link>
      <div
        className="flex flex-col items-start"
        style={alignTop ? { marginTop: 6 } : undefined}
      >
        <Link to={link} style={{ minWidth: sizeToPx(size) }}>
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
