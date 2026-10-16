import { Stack } from "expo-router";

import { HomeScreen } from "@/features/home/home-screen";

export default function HomeRoute() {
  return (
    <>
      <Stack.Screen options={{ title: "Home" }} />
      <HomeScreen />
    </>
  );
}
