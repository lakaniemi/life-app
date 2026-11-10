import {
  onlineManager,
  QueryClient,
  QueryClientProvider,
} from "@tanstack/react-query";
import { act, render, screen } from "@testing-library/react-native";

import { OfflineBanner } from "@/ui/offline-banner";

const bannerText = "Offline · showing cached data";

// No garbage collection, whose timers would keep Jest from exiting.
const queryClient = new QueryClient({
  defaultOptions: { queries: { gcTime: Infinity } },
});

const renderBanner = () =>
  render(
    <QueryClientProvider client={queryClient}>
      <OfflineBanner />
    </QueryClientProvider>,
  );

afterEach(() => {
  onlineManager.setOnline(true);
  queryClient.clear();
});

test("is hidden while online", async () => {
  queryClient.setQueryData(["item"], "cached");

  await renderBanner();

  expect(screen.queryByText(bannerText)).not.toBeOnTheScreen();
});

test("shows while offline and hides when the connection returns", async () => {
  queryClient.setQueryData(["item"], "cached");
  await renderBanner();

  await act(() => onlineManager.setOnline(false));
  expect(screen.getByText(bannerText)).toBeOnTheScreen();

  await act(() => onlineManager.setOnline(true));
  expect(screen.queryByText(bannerText)).not.toBeOnTheScreen();
});

test("doesn't count a query that opts out of persistence as cached data", async () => {
  await queryClient.fetchQuery({
    queryKey: ["session-only"],
    queryFn: () => "data",
    meta: { persist: false },
  });
  await renderBanner();

  await act(() => onlineManager.setOnline(false));

  expect(screen.queryByText(bannerText)).not.toBeOnTheScreen();
});

test("is hidden while offline with nothing cached, and shows once data arrives", async () => {
  await renderBanner();

  await act(() => onlineManager.setOnline(false));
  expect(screen.queryByText(bannerText)).not.toBeOnTheScreen();

  await act(() => queryClient.setQueryData(["item"], "cached"));
  expect(screen.getByText(bannerText)).toBeOnTheScreen();
});
