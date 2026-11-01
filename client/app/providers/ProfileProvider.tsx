import {
  createContext,
  useContext,
  useLayoutEffect,
  useRef,
  useState,
  type ReactNode,
} from "react";
import {
  QueryClient,
  QueryClientProvider,
  useQuery,
} from "@tanstack/react-query";
import {
  createProfileApi,
  type ListeningProfile,
  type ProfileApi,
} from "api/api";

type ProfileContextValue = { profile: ListeningProfile; api: ProfileApi };
const ProfileContext = createContext<ProfileContextValue | null>(null);
export function useOptionalProfile() {
  return useContext(ProfileContext);
}
export function useProfile() {
  const value = useOptionalProfile();
  if (!value) throw new Error("This screen requires a listening profile");
  return value;
}
export function useProfileApi() {
  return useProfile().api;
}

export function ProfileProvider({
  profile,
  children,
}: {
  profile: ListeningProfile;
  children: ReactNode;
}) {
  const controller = useRef(new AbortController());
  const [client] = useState(() => new QueryClient());
  const [api] = useState(() =>
    createProfileApi(profile.slug, () => controller.current.signal),
  );
  useLayoutEffect(() => {
    // StrictMode replays effects in development. Each mounted lifetime gets a live signal.
    if (controller.current.signal.aborted)
      controller.current = new AbortController();
    return () => {
      controller.current.abort();
      void client.cancelQueries();
      client.clear();
    };
  }, [client]);
  return (
    <ProfileContext.Provider value={{ profile, api }}>
      <QueryClientProvider client={client}>{children}</QueryClientProvider>
    </ProfileContext.Provider>
  );
}

export function useFirstActivity() {
  const { profile, api } = useProfile();
  const { data } = useQuery({
    queryKey: ["profile", profile.id, "first-activity"],
    queryFn: () =>
      api.apiFetch<{ time: string }>("/apis/web/v1/first-activity"),
  });
  const date = data?.time ? new Date(data.time) : undefined;
  return date && date.getTime() > 0 ? date : undefined;
}
