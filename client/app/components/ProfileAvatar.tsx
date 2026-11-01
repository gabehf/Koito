import { useState } from "react";
import { Disc3, Headphones, Music2, Radio, UserRound } from "lucide-react";
import type { ListeningProfile } from "api/api";
const icons = {
  music: Music2,
  headphones: Headphones,
  disc: Disc3,
  radio: Radio,
  user: UserRound,
};
export default function ProfileAvatar({
  profile,
  size = 36,
}: {
  profile: ListeningProfile;
  size?: number;
}) {
  const [failed, setFailed] = useState(false);
  const Icon = profile.avatar_icon
    ? icons[profile.avatar_icon as keyof typeof icons]
    : undefined;
  const initials = profile.display_name
    .trim()
    .split(/\s+/)
    .slice(0, 2)
    .map((word) => Array.from(word)[0] ?? "")
    .join("")
    .toLocaleUpperCase();
  const style = {
    width: size,
    height: size,
    minWidth: size,
    fontSize: Math.max(12, Math.round(size * 0.32)),
  };
  return (
    <span
      style={style}
      className="inline-flex items-center justify-center overflow-hidden rounded-full bg-(--color-bg-tertiary) text-(--color-fg) border border-(--color-border)"
      aria-hidden="true"
    >
      {profile.avatar_image_id && !failed ? (
        <img
          src={`/image/${encodeURIComponent(profile.avatar_image_id)}/128x128.webp`}
          className="w-full h-full object-cover"
          alt=""
          onError={() => setFailed(true)}
        />
      ) : Icon ? (
        <Icon size={size * 0.5} />
      ) : (
        initials || <UserRound size={size * 0.5} />
      )}
    </span>
  );
}
