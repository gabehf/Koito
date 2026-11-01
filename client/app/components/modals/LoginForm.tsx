import { login } from "api/api";
import { useState } from "react";
import { useAppContext } from "~/providers/AppProvider";
import SubHeader from "../primitives/SubHeader";

export default function LoginForm() {
  const { refreshAdminSession } = useAppContext();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [remember, setRemember] = useState(false);

  const loginHandler = async () => {
    if (!username || !password) {
      setError("username and password are required");
      return;
    }
    setLoading(true);
    setError("");
    try {
      const response = await login(username, password, remember);
      if (!response.ok) {
        const data = await response.json();
        throw new Error(data.error || "Login failed");
      }
      await refreshAdminSession();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Login failed");
    } finally {
      setLoading(false);
    }
  };

  return (
    <>
      <SubHeader>Log In</SubHeader>
      <div className="flex flex-col items-center gap-4 w-full">
        <p>
          Logging in gives you access to <strong>admin tools</strong>, such as
          updating images, merging items, deleting items, and more.
        </p>
        <form
          action="#"
          className="flex flex-col items-center gap-4 w-3/4"
          onSubmit={(e) => {
            e.preventDefault();
            void loginHandler();
          }}
        >
          <input
            name="koito-username"
            type="text"
            placeholder="Username"
            className="w-full mx-auto fg bg rounded p-2"
            onChange={(e) => setUsername(e.target.value)}
          />
          <input
            name="koito-password"
            type="password"
            placeholder="Password"
            className="w-full mx-auto fg bg rounded p-2"
            onChange={(e) => setPassword(e.target.value)}
          />
          <div className="flex gap-2 items-center">
            <input
              type="checkbox"
              name="koito-remember"
              id="koito-remember"
              onChange={() => setRemember(!remember)}
            />
            <label htmlFor="koito-remember">Remember me</label>
          </div>
          <button
            type="submit"
            className="large-button disabled:opacity-50"
            disabled={loading}
          >
            {loading ? "Logging in…" : "Login"}
          </button>
        </form>
        <p className="error">{error}</p>
      </div>
    </>
  );
}
