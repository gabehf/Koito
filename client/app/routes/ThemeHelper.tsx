import { useState } from "react";
import { AsyncButton } from "../components/AsyncButton";
import AllTimeStats from "~/components/AllTimeStats";
import ActivityGrid from "~/components/ActivityGrid";
import LastPlayed from "~/components/LastPlayed";
import TopAlbums from "~/components/TopAlbums";
import TopArtists from "~/components/TopArtists";
import TopTracks from "~/components/TopTracks";
import { useTheme } from "~/hooks/useTheme";
import { type Theme } from "~/styles/themes.css";
import TopArtistsCard from "~/components/TopArtistsCard";
import TopAlbumsCard from "~/components/TopAlbumsCard";

export default function ThemeHelper() {
  const initialTheme = {
    bg: "#1e1816",
    bgSecondary: "#2f2623",
    bgTertiary: "#453733",
    fg: "#f8f3ec",
    fgSecondary: "#d6ccc2",
    fgTertiary: "#b4a89c",
    primary: "#f5a97f",
    primaryDim: "#d88b65",
    accent: "#f9db6d",
    accentDim: "#d9bc55",
    error: "#e26c6a",
    warning: "#f5b851",
    success: "#8fc48f",
    info: "#87b8dd",
  };

  const [custom, setCustom] = useState(
    JSON.stringify(initialTheme, null, "  ")
  );
  const { setCustomTheme } = useTheme();

  const handleCustomTheme = () => {
    console.log(custom);
    try {
      const theme = JSON.parse(custom) as Theme;
      console.log(theme);
      setCustomTheme(theme);
    } catch (err) {
      console.log(err);
    }
  };

  const homeItems = 10;

  return (
    <div className="mt-10 flex flex-col gap-10 items-center">
      <div className="flex gap-5">
        <AllTimeStats />
        <ActivityGrid />
      </div>
      <div className="container justify-center flex flex-wrap gap-10">
        {/*<TopArtists period="all_time" limit={homeItems} />
        <TopAlbums period="all_time" limit={homeItems} />*/}
        <TopArtistsCard period="all_time" />
        <TopAlbumsCard period="all_time" />
        <TopTracks period="all_time" limit={homeItems} />
        <LastPlayed limit={5} />
      </div>
      <div className="flex gap-10">
        <div className="flex flex-col items-center gap-3 bg-secondary p-5 rounded-lg">
          <textarea
            name="custom-theme"
            onChange={(e) => setCustom(e.target.value)}
            id="custom-theme-input"
            className="bg-(--color-bg) w-[300px] p-5 h-full rounded-md"
            value={custom}
          />
          <AsyncButton onClick={handleCustomTheme}>Submit</AsyncButton>
        </div>
        <div className="flex flex-col gap-6 bg-secondary p-10 rounded-lg">
          <div className="flex flex-col gap-4 items-center">
            <p>
              You"re logged in as <strong>Example User</strong>
            </p>
            <AsyncButton loading={false} onClick={() => {}}>
              Logout
            </AsyncButton>
          </div>
          <div className="flex gap-4">
            <input
              name="koito-update-username"
              type="text"
              placeholder="Update username"
              className="w-full mx-auto fg bg rounded p-2"
            />
            <AsyncButton loading={false} onClick={() => {}}>
              Submit
            </AsyncButton>
          </div>
          <div className="flex gap-4">
            <input
              name="koito-update-password"
              type="password"
              placeholder="Update password"
              className="w-full mx-auto fg bg rounded p-2"
            />
            <input
              name="koito-confirm-password"
              type="password"
              placeholder="Confirm password"
              className="w-full mx-auto fg bg rounded p-2"
            />
            <AsyncButton loading={false} onClick={() => {}}>
              Submit
            </AsyncButton>
          </div>
          <div className="flex gap-2 mt-3">
            <input
              type="checkbox"
              name="reverse-merge-order"
              onChange={() => {}}
            />
            <label htmlFor="reverse-merge-order">Example checkbox</label>
          </div>
          <p className="success">successfully displayed example text</p>
          <p className="error">this is an example of error text</p>
          <p className="info">here is an informational example</p>
          <p className="warning">heed this warning, traveller</p>
        </div>
      </div>
    </div>
  );
}
