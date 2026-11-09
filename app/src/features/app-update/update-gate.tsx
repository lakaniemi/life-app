import type { ReactNode } from "react";

import { useIsAppOutdated } from "@/api/compat";
import { UpdateRequiredScreen } from "@/features/app-update/update-required-screen";

// Replaces the whole app with the "Update required" screen once the API has
// turned it away, so nothing else is reachable.
export const UpdateGate = ({ children }: { children: ReactNode }) => {
  const isOutdated = useIsAppOutdated();

  return isOutdated ? <UpdateRequiredScreen /> : children;
};
