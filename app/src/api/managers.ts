import { focusManager, onlineManager } from "@tanstack/react-query";
import { addNetworkStateListener, getNetworkStateAsync } from "expo-network";
import { AppState } from "react-native";

// TanStack Query listens to the browser's focus and online events, which
// React Native doesn't have. These hand it the native equivalents, so it
// refetches stale queries when the app returns to the foreground or the
// connection comes back, and pauses them while offline.
export const connectQueryManagers = () => {
  focusManager.setEventListener((setFocused) => {
    const subscription = AppState.addEventListener("change", (state) => {
      setFocused(state === "active");
    });
    return () => subscription.remove();
  });

  onlineManager.setEventListener((setOnline) => {
    // Only isConnected, not isInternetReachable: on Android the latter needs
    // Google's connectivity check to pass, which emulators and some networks
    // fail. Wrongly offline pauses every query; wrongly online only fails
    // them with "Couldn't connect". Unknown counts as online.
    // The listener fires only on changes, hence the initial read. A change
    // that arrives before the read resolves is newer, so the read then loses.
    let hasChanged = false;
    const subscription = addNetworkStateListener((state) => {
      hasChanged = true;
      setOnline(state.isConnected !== false);
    });
    getNetworkStateAsync()
      .then((state) => {
        if (!hasChanged) {
          setOnline(state.isConnected !== false);
        }
      })
      // Without a reading, stay with TanStack Query's default: online.
      .catch(() => {});
    return () => subscription.remove();
  });
};
