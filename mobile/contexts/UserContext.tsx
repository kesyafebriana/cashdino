import {
  createContext,
  useContext,
  useState,
  useEffect,
  useCallback,
  type ReactNode,
} from "react";
import AsyncStorage from "@react-native-async-storage/async-storage";
import {
  fetchUsers,
  fetchBanner,
  type UserInfo,
  type BannerResponse,
} from "@/services/api";

interface UserContextValue {
  currentUser: UserInfo | null;
  users: UserInfo[];
  switchUser: (user: UserInfo) => void;
  banner: BannerResponse | null;
  rankChange: number | null;
  bannerLoading: boolean;
  refreshBanner: () => void;
  loading: boolean;
  error: string | null;
  retry: () => void;
}

const UserContext = createContext<UserContextValue>({
  currentUser: null,
  users: [],
  switchUser: () => {},
  banner: null,
  rankChange: null,
  bannerLoading: true,
  refreshBanner: () => {},
  loading: true,
  error: null,
  retry: () => {},
});

function parseRankNumber(rankDisplay: string): number | null {
  if (rankDisplay === "99+") return null;
  const match = rankDisplay.match(/^#(\d+)$/);
  return match ? parseInt(match[1], 10) : null;
}

function rankStorageKey(userId: string, challengeId: string): string {
  return `rank:${userId}:${challengeId}`;
}

export function UserProvider({ children }: { children: ReactNode }) {
  const [users, setUsers] = useState<UserInfo[]>([]);
  const [currentUser, setCurrentUser] = useState<UserInfo | null>(null);
  const [banner, setBanner] = useState<BannerResponse | null>(null);
  const [rankChange, setRankChange] = useState<number | null>(null);
  const [bannerLoading, setBannerLoading] = useState(true);
  const [bannerRefreshKey, setBannerRefreshKey] = useState(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const loadUsers = useCallback(() => {
    setLoading(true);
    setError(null);
    fetchUsers(["james", "olivia", "nathan", "oscar"])
      .then((data) => {
        setUsers(data);
        if (data.length > 0) {
          setCurrentUser(data[0]);
        }
      })
      .catch((err) => {
        console.warn("Failed to fetch users:", err.message);
        setError(err.message);
      })
      .finally(() => setLoading(false));
  }, []);

  useEffect(() => {
    loadUsers();
  }, [loadUsers]);

  const refreshBanner = useCallback(() => {
    setBannerRefreshKey((k) => k + 1);
  }, []);

  // Fetch banner whenever currentUser changes or refresh is triggered
  useEffect(() => {
    if (!currentUser) return;
    setBannerLoading(true);
    setBanner(null);
    setRankChange(null);

    fetchBanner(currentUser.id)
      .then(async (data) => {
        setBanner(data);

        const currentRank = parseRankNumber(data.rank_display);
        const key = rankStorageKey(currentUser.id, data.challenge_id);

        try {
          const prev = await AsyncStorage.getItem(key);
          if (prev !== null && currentRank !== null) {
            const prevRank = parseInt(prev, 10);
            // Positive = moved up (rank number decreased), negative = moved down
            const change = prevRank - currentRank;
            setRankChange(change !== 0 ? change : null);
          }
          // Save current rank for next comparison
          if (currentRank !== null) {
            await AsyncStorage.setItem(key, String(currentRank));
          }
        } catch {
          // AsyncStorage error — ignore, rank change just won't show
        }
      })
      .catch(() => setBanner(null))
      .finally(() => setBannerLoading(false));
  }, [currentUser, bannerRefreshKey]);

  const switchUser = (user: UserInfo) => {
    setCurrentUser(user);
  };

  return (
    <UserContext.Provider
      value={{
        currentUser,
        users,
        switchUser,
        banner,
        rankChange,
        bannerLoading,
        refreshBanner,
        loading,
        error,
        retry: loadUsers,
      }}
    >
      {children}
    </UserContext.Provider>
  );
}

export function useUser() {
  return useContext(UserContext);
}
