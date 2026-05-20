import React, {
  createContext,
  useContext,
  useEffect,
  useRef,
  type RefObject,
} from "react";
import { useNavigate } from "react-router-dom";
import { useAuth } from "./AuthContext";

const WSContext = createContext<RefObject<WebSocket | null> | null>(null);

export const useWS = () => useContext(WSContext);

export function WSProvider({ children }: { children: React.ReactNode }) {
  const wsRef = useRef<WebSocket | null>(null);
  const navigate = useNavigate();
  const { user } = useAuth();

  useEffect(() => {
    if (!user) return;

    const ws = new WebSocket("ws://localhost:8080/ws");

    ws.onopen = () => {
      ws.send(JSON.stringify({ type: "auth", payload: {} }));
      wsRef.current = ws;
    };

    ws.onmessage = (event) => {
      const data = JSON.parse(event.data);
      if (data.type === "match_found") {
        localStorage.setItem(`color_${data.gameId}`, data.payload.playerColor);
        navigate(`/play/${data.gameId}`);
      }
    };

    ws.onclose = () => console.log("lobby ws disconnected");
    ws.onerror = (err) => console.error("lobby ws error:", err);

    return () => ws.close();
  }, [navigate, user]);

  return <WSContext.Provider value={wsRef}>{children}</WSContext.Provider>;
}
