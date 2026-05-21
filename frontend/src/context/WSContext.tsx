import React, {
  createContext,
  useContext,
  useEffect,
  useRef,
  type RefObject,
} from "react";
import { useNavigate } from "react-router-dom";
import { useAuth } from "./AuthContext";
import { ReconnectingWebSocket } from "../lib/reconnectingWebSocket";

const WSContext = createContext<RefObject<ReconnectingWebSocket | null> | null>(null);

export const useWS = () => useContext(WSContext);

export function WSProvider({ children }: { children: React.ReactNode }) {
  const wsRef = useRef<ReconnectingWebSocket | null>(null);
  const navigate = useNavigate();
  const { user } = useAuth();

  useEffect(() => {
    if (!user) return;

    const ws = new ReconnectingWebSocket("ws://localhost:8080/ws", {
      onOpen: () => {
        ws.send(JSON.stringify({ type: "auth", payload: {} }));
      },
      onMessage: (event) => {
        const data = JSON.parse(event.data);
        if (data.type === "match_found") {
          localStorage.setItem(`color_${data.gameId}`, data.payload.playerColor);
          navigate(`/play/${data.gameId}`);
        }
      },
    });
    wsRef.current = ws;

    return () => {
      ws.close();
      wsRef.current = null;
    };
  }, [navigate, user]);

  return <WSContext.Provider value={wsRef}>{children}</WSContext.Provider>;
}
