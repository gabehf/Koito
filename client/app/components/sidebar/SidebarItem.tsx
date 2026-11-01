import React, { useState } from "react";
import Popup from "../Popup";

import { ProfileLink as Link } from "~/components/ProfileLink";

interface Props {
  name: string;
  to?: string;
  onClick: Function;
  children: React.ReactNode;
  modal: React.ReactNode;
  keyHint?: React.ReactNode;
  space?: number;
  externalLink?: boolean;
  /* true if the keyhint is an icon and not text */
  icon?: boolean;
}

export default function SidebarItem({
  externalLink,
  space,
  keyHint,
  name,
  to,
  children,
  modal,
  onClick,
  icon,
}: Props) {
  const classes =
    "hover:cursor-pointer hover:bg-(--color-bg-tertiary) hover:text-(--color-fg) transition duration-100 rounded-md p-2 inline-block";

  const popupInner = keyHint ? (
    <div className="flex items-center gap-2">
      <span>{name}</span>
      {icon ? (
        <div>{keyHint}</div>
      ) : (
        <kbd className="px-1 text-sm rounded bg-(--color-bg-tertiary) text-(--color-fg) border border-[var(--color-fg)]">
          {keyHint}
        </kbd>
      )}
    </div>
  ) : (
    name
  );

  return (
    <>
      <Popup position="right" space={space ?? 20} inner={popupInner}>
        {to ? (
          <Link
            aria-label={name}
            target={externalLink ? "_blank" : ""}
            className={classes}
            to={to}
          >
            {children}
          </Link>
        ) : (
          <button
            type="button"
            aria-label={name}
            className={classes}
            onClick={() => onClick()}
          >
            {children}
          </button>
        )}
      </Popup>
      {modal}
    </>
  );
}
