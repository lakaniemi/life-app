import type { ReactNode } from "react";

import { useIsAppOutdated } from "@/api/compat";
import { InstallUpdateScreen } from "@/features/app-update/install-update-screen";

// Replaces the whole app with the update screen once the API has turned it
// away, so nothing else is reachable.
export const UpdateGate = ({ children }: { children: ReactNode }) => {
  const isOutdated = useIsAppOutdated();

  return isOutdated ? <InstallUpdateScreen required /> : children;
};
