import {
  useQuery,
  type QueryKey,
  type UseQueryOptions,
} from "@tanstack/react-query";
import { useIsFocused } from "expo-router";

// useQuery for queries that a screen renders. React Navigation keeps the
// screens below the top of the stack mounted, and their queries would stay
// active: refetched when the app returns to the foreground, though nobody
// sees them. Unsubscribing while the screen is out of focus makes them
// inactive, and subscribing again when it regains focus refetches them if
// they're stale. Features use this instead of useQuery (lint enforces it).
export const useApiQuery = <
  TQueryFnData,
  TError = Error,
  TData = TQueryFnData,
  TQueryKey extends QueryKey = QueryKey,
>(
  options: UseQueryOptions<TQueryFnData, TError, TData, TQueryKey>,
) => {
  const isFocused = useIsFocused();
  return useQuery({ ...options, subscribed: isFocused });
};
