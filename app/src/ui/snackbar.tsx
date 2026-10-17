import {
  createContext,
  useContext,
  useRef,
  useState,
  type ReactNode,
} from "react";
import { Snackbar } from "react-native-paper";

type SnackbarOptions = {
  action?: { label: string; onPress: () => void };
  // In milliseconds. Defaults to Paper's medium duration (7 s).
  duration?: number;
};

type Message = SnackbarOptions & { id: number; text: string };

type SnackbarContextValue = {
  show: (text: string, options?: SnackbarOptions) => void;
};

const SnackbarContext = createContext<SnackbarContextValue | null>(null);

// One snackbar host for the whole app. Paper's Snackbar is only a component,
// so the queue lives here: messages show one at a time, in the order they
// were shown, and each waits until the one before it is dismissed.
export const SnackbarProvider = ({ children }: { children: ReactNode }) => {
  const [queue, setQueue] = useState<Message[]>([]);
  const nextId = useRef(0);

  const show = (text: string, options?: SnackbarOptions) => {
    const id = nextId.current++;
    setQueue((messages) => [...messages, { ...options, id, text }]);
  };

  const dismiss = (id: number) => {
    setQueue((messages) => messages.filter((message) => message.id !== id));
  };

  const current = queue[0];

  return (
    <SnackbarContext.Provider value={{ show }}>
      {children}
      {current && (
        // A new key per message remounts the Snackbar, so each one gets
        // Paper's entry animation and its own dismiss timer.
        <Snackbar
          key={current.id}
          visible
          onDismiss={() => dismiss(current.id)}
          action={current.action}
          duration={current.duration}
        >
          {current.text}
        </Snackbar>
      )}
    </SnackbarContext.Provider>
  );
};

export const useSnackbar = (): SnackbarContextValue => {
  const context = useContext(SnackbarContext);
  if (context === null) {
    throw new Error("useSnackbar must be used inside a SnackbarProvider");
  }
  return context;
};
