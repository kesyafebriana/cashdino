import { useState } from "react";

export function useCurrentUser() {
  const [userId, setUserId] = useState<number>(1);
  return { userId, setUserId };
}
