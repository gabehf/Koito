import { login } from "api/api";
import { useEffect, useState } from "react";
import { AsyncButton } from "../AsyncButton";

export default function LoginForm() {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [remember, setRemember] = useState(false);

  const loginHandler = () => {
    if (username && password) {
      setLoading(true);
      login(username, password, remember)
        .then((r) => {
          if (r.status >= 200 && r.status < 300) {
            window.location.reload();
          } else {
            r.json().then((r) => setError(r.error));
          }
        })
        .catch((err) => setError(err));
      setLoading(false);
    } else if (username || password) {
      setError("username and password are required");
    }
  };

  return (
    <>
      <h3>Log In</h3>
      <div className="flex flex-col items-center gap-4 w-full">
        <p>
          Logging in gives you access to <strong>admin tools</strong>, such as
          updating images, merging items, deleting items, and more.
        </p>
        <form
          action="#"
          className="flex flex-col items-center gap-4 w-3/4"
          onSubmit={(e) => e.preventDefault()}
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
            <label htmlFor="kotio-remember">Remember me</label>
          </div>
          <AsyncButton loading={loading} onClick={loginHandler}>
            Login
          </AsyncButton>
        </form>
        <p className="error">{error}</p>
      </div>
    </>
  );
}
