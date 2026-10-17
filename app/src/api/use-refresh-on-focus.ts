import { useQueryClient } from "@tanstack/react-query";
import { useFocusEffect } from "expo-router";
import { useCallback, useRef } from "react";

// Refetches stale data when the screen comes back into focus, e.g. on going
// back to it. React Navigation keeps screens below the top of the stack
// mounted, so their queries don't refetch on their own. Only stale queries
// refetch, so quickly moving between screens doesn't. The first focus is
// skipped: the queries fetch on mount anyway.
export const useRefreshOnFocus = () => {
  const queryClient = useQueryClient();
  const isFirstFocus = useRef(true);

  useFocusEffect(
    useCallback(() => {
      if (isFirstFocus.current) {
        isFirstFocus.current = false;
        return;
      }
      void queryClient.refetchQueries({ stale: true, type: "active" });
    }, [queryClient]),
  );
};
