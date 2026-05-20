import React, {
  createContext,
  useContext,
  useEffect,
  useRef,
  type RefObject,
} from "react";
import { useNavigate } from "react-router-dom";

const WSContext = createContext<RefObject<WebSocket | null> | null>(null);

export const useWS = () => useContext(WSContext);

export function WSProvider({ children }: { children: React.ReactNode }) {
  const wsRef = useRef<WebSocket | null>(null);
  const navigate = useNavigate();

  useEffect(() => {
    if (!localStorage.getItem("playerId")) {
      localStorage.setItem("playerId", crypto.randomUUID());
    }
    const playerId = localStorage.getItem("playerId");

    const ws = new WebSocket("ws://localhost:8080/ws");

    ws.onopen = () => {
      ws.send(JSON.stringify({ type: "auth", payload: { playerId } }));
      wsRef.current = ws;
    };

    ws.onmessage = (event) => {
      const data = JSON.parse(event.data);
      if (data.type === "match_found") {
        const color: string = data.payload.playerColor;
        localStorage.setItem(playerId!, color);
        navigate(`/play/${data.gameId}`);
      }
    };

    ws.onclose = () => console.log("lobby ws disconnected");
    ws.onerror = (err) => console.error("lobby ws error:", err);

    return () => ws.close();
  }, [navigate]);

  return <WSContext.Provider value={wsRef}>{children}</WSContext.Provider>;
}
