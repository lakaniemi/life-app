import { Stack } from "expo-router";
import { useTranslation } from "react-i18next";

import { HomeScreen } from "@/features/home/home-screen";

const HomeRoute = () => {
  const { t } = useTranslation();

  return (
    <>
      <Stack.Screen options={{ title: t("home.title") }} />
      <HomeScreen />
    </>
  );
};

export default HomeRoute;
