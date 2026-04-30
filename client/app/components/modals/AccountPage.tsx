import { useAppContext } from "~/providers/AppProvider";
import LoginForm from "./LoginForm";
import Account from "./Account";

export default function AuthForm() {
  const { user } = useAppContext();

  return <>{user ? <Account /> : <LoginForm />}</>;
}
