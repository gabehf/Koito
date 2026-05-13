import { useEffect, useState } from "react";
import { useAppContext } from "~/providers/AppProvider";
import { ExternalLink } from "lucide-react";
import { Link } from "react-router";

type VersionInfo = {
  html_url: string;
  name: string;
  body: string;
};

export default function About() {
  const { updateAvailable, currentVersion } = useAppContext();
  const [latestVersionInfo, setLatestVersionInfo] = useState<
    VersionInfo | undefined
  >();

  useEffect(() => {
    fetch("https://api.github.com/repos/gabehf/koito/releases/latest")
      .then((r) => r.json())
      .then((r) => setLatestVersionInfo(r))
      .catch((err) => console.log(err));
  }, []);

  return (
    <div>
      <div className="w-full bg p-6 rounded-sm flex flex-col items-center gap-4">
        <div className="inline-flex items-center">
          <img
            src="/web-app-manifest-192x192.png"
            alt="koito logo"
            style={{ width: 70 }}
          />
          <h5 className="text-6xl font-semibold">Koito</h5>
        </div>
        <div className="px-2 py-1 rounded-sm bg-secondary">
          Koito {currentVersion}
        </div>
        {updateAvailable && latestVersionInfo && (
          <Link
            className="text-(--color-info) text-sm mt-1 hover:cursor-pointer"
            to={latestVersionInfo.html_url}
            target="_blank"
          >
            <span className="inline-flex items-center gap-1">
              🛈 Update available - click to view release
              <ExternalLink size={14} />
            </span>
          </Link>
        )}
        {!latestVersionInfo && <div>Loading...</div>}
      </div>
    </div>
  );
}
