import "chessground/assets/chessground.base.css";
import "chessground/assets/chessground.brown.css";
import "chessground/assets/chessground.cburnett.css";
import ChessBoard from "./ChessBoard";
import GameControls from "./GameControls";
import GameInfo from "./GameInfo";
import { useChessGame } from "../hooks/useChessGame";
import { useEffect, useRef, useState } from "react";
import { useParams } from "react-router-dom";
import type { Color } from "chessground/types";

export default function Play() {
  const { game, fen, pgn, move, reset, undo, status } = useChessGame();
  const stored = localStorage.getItem(
    localStorage.getItem("playerId") ?? crypto.randomUUID()
  );
  const orientation: Color = stored === "black" ? "black" : "white";
  const [, setTurnColor] = useState<string>("white");
  const gameWsRef = useRef<WebSocket | null>(null);
  const { gameId } = useParams();

  useEffect(() => {
    if (!gameId) return;

    const playerId = localStorage.getItem("playerId");
    const ws = new WebSocket("ws://localhost:8080/ws");
    gameWsRef.current = ws;

    ws.onopen = () => {
      ws.send(
        JSON.stringify({ type: "auth", payload: { playerId, gameId } })
      );
    };

    ws.onmessage = (event) => {
      const data = JSON.parse(event.data);
      if (data.type === "move") {
        move(data.payload.from, data.payload.to);
        setTurnColor((t) => (t === "white" ? "black" : "white"));
      }
    };

    ws.onclose = () => console.log("game ws disconnected");
    ws.onerror = (err) => console.error("game ws error:", err);

    return () => ws.close();
  }, [gameId]);

  const onMove = (from: string, to: string) => {
    const prevLen = game.history().length;
    move(from, to);
    if (game.history().length === prevLen) return;

    const moveMade = game.history()[game.history().length - 1];

    gameWsRef.current?.send(
      JSON.stringify({
        type: "move",
        gameId,
        payload: {
          move: moveMade,
          from,
          to,
          playerId: localStorage.getItem("playerId"),
        },
      })
    );

    setTurnColor((t) => (t === "white" ? "black" : "white"));
  };

  return (
    <div>
      <div>
        <GameInfo pgn={pgn} fen={fen} status={status} />
        <ChessBoard
          game={game}
          fen={fen}
          onMove={onMove}
          orientation={orientation}
        />
        <GameControls onReset={reset} onUndo={undo} onFlip={() => {}} />
      </div>
    </div>
  );
}
