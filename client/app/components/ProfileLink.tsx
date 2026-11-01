import { Link, type LinkProps } from "react-router";
import { useOptionalProfile } from "~/providers/ProfileProvider";
import { profilePath } from "../../api/profile-paths";
export function ProfileLink({ to, ...props }: LinkProps) {
  const scope = useOptionalProfile();
  const path =
    typeof to === "string" &&
    /^\/(artist|album|track|chart|listens|rewind)(\/|\?|$)/.test(to)
      ? profilePath(scope?.profile.slug ?? "", to)
      : to;
  return <Link {...props} to={path} />;
}
