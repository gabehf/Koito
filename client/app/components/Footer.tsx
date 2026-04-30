import { ExternalLinkIcon } from "lucide-react";
import pkg from "../../package.json";

export default function Footer() {
  return (
    <div className="mx-auto py-10 pt-20 color-fg-tertiary text-sm">
      <ul className="flex flex-col items-center w-sm justify-around">
        <li>Koito {import.meta.env.VITE_KOITO_VERSION || pkg.version}</li>
        <li>
          <a
            href="https://github.com/gabehf/koito"
            target="_blank"
            className="link-underline"
          >
            View the source on GitHub{" "}
            <ExternalLinkIcon className="inline mb-1" size={14} />
          </a>
        </li>
      </ul>
    </div>
  );
}
