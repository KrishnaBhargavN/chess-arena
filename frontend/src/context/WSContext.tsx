import { eventPosition } from "chessground/util";
import React, {
  createContext,
  useContext,
  useEffect,
  useRef,
  type RefObject,
} from "react";
import { useNavigate } from "react-router-dom";
import { useChessGame } from "../hooks/useChessGame";

const WSContext = createContext<RefObject<WebSocket> | null>(null);

export const useWS = () => useContext(WSContext);

export function WSProvider({ children }: { children: React.ReactNode }) {
  const wsRef = useRef<WebSocket>(null);
  const navigate = useNavigate();
  const ws = new WebSocket("ws://localhost:8080/ws");
  const { move } = useChessGame();
  useEffect(() => {
    if (!localStorage.getItem("playerId")) {
      localStorage.setItem("playerId", crypto.randomUUID());
    }
    const playerId = localStorage.getItem("playerId");

    ws.onopen = () => {
      console.log("WS connected");

      ws.send(
        JSON.stringify({
          type: "auth",
          payload: { playerId },
        }),
      );

      ws.onmessage = (event) => {
        const data = JSON.parse(event.data);
        console.log(data);

        if (data.type === "match_found") {
          navigate(`/play/${data.gameId}`);
        }
      };

      ws.onclose = () => {
        console.log("ws disconnected");
      };

      ws.onerror = (err) => {
        console.error("WS error: ", err);
      };

      wsRef.current = ws;

      return () => {
        ws.close();
      };
    };
  }, []);

  return <WSContext.Provider value={wsRef}>{children}</WSContext.Provider>;
}
