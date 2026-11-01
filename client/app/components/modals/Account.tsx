import { logout, updateUser } from "api/api";
import { useState } from "react";
import { AsyncButton } from "../AsyncButton";
import { useAppContext } from "~/providers/AppProvider";
import SubHeader from "../primitives/SubHeader";

export default function Account() {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [confirmPw, setConfirmPw] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const {
    user,
    setUsername: setCtxUsername,
    refreshAdminSession,
  } = useAppContext();

  const logoutHandler = () => {
    setLoading(true);
    logout()
      .then(async (r) => {
        if (r.ok) {
          await refreshAdminSession();
        } else {
          r.json().then((r) => setError(r.error));
        }
      })
      .catch((err) =>
        setError(err instanceof Error ? err.message : "Request failed"),
      )
      .finally(() => setLoading(false));
  };
  const updateHandler = () => {
    setError("");
    setSuccess("");
    if (!username && !password) {
      setError("provide a username or a new password");
      return;
    }
    if (password && password !== confirmPw) {
      setError("password confirmation does not match");
      return;
    }
    if (password && Array.from(password).length < 8) {
      setError("password must be at least 8 characters long");
      return;
    }
    setError("");
    setSuccess("");
    setLoading(true);
    updateUser(username, password)
      .then((r) => {
        if (r.ok) {
          setSuccess("sucessfully updated user");
          if (username != "") {
            setCtxUsername(username);
          }
          setUsername("");
          setPassword("");
          setConfirmPw("");
        } else {
          r.json().then((r) => setError(r.error));
        }
      })
      .catch((err) =>
        setError(err instanceof Error ? err.message : "Request failed"),
      )
      .finally(() => setLoading(false));
  };

  return (
    <>
      <SubHeader>Account</SubHeader>
      <div className="flex flex-col gap-6">
        <div className="flex flex-col gap-4 items-center">
          <p>
            You're logged in as <strong>{user?.username}</strong>
          </p>
          <AsyncButton loading={loading} onClick={logoutHandler}>
            Logout
          </AsyncButton>
        </div>
        <SubHeader>Update User</SubHeader>
        <form
          action="#"
          onSubmit={(e) => e.preventDefault()}
          className="flex flex-col gap-4"
        >
          <div className="flex gap-4">
            <input
              name="koito-update-username"
              type="text"
              placeholder="Update username"
              className="w-full mx-auto fg bg rounded p-2"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
            />
          </div>
          <div className="w-fit">
            <AsyncButton loading={loading} onClick={updateHandler}>
              Submit
            </AsyncButton>
          </div>
        </form>
        <form
          action="#"
          onSubmit={(e) => e.preventDefault()}
          className="flex flex-col gap-4"
        >
          <div className="flex gap-4">
            <input
              name="koito-update-password"
              type="password"
              placeholder="Update password"
              className="w-full mx-auto fg bg rounded p-2"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
            />
            <input
              name="koito-confirm-password"
              type="password"
              placeholder="Confirm new password"
              className="w-full mx-auto fg bg rounded p-2"
              value={confirmPw}
              onChange={(e) => setConfirmPw(e.target.value)}
            />
          </div>
          <div className="w-fit">
            <AsyncButton loading={loading} onClick={updateHandler}>
              Submit
            </AsyncButton>
          </div>
        </form>
        {success != "" && <p className="success">{success}</p>}
        {error != "" && <p className="error">{error}</p>}
      </div>
    </>
  );
}
