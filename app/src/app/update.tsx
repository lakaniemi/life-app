import { Stack } from "expo-router";
import { useTranslation } from "react-i18next";

import { InstallUpdateScreen } from "@/features/app-update/install-update-screen";

const UpdateRoute = () => {
  const { t } = useTranslation();

  return (
    <>
      <Stack.Screen options={{ title: t("appUpdate.title") }} />
      <InstallUpdateScreen />
    </>
  );
};

export default UpdateRoute;
