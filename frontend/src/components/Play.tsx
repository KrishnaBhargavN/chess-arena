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
import { ReconnectingWebSocket } from "../lib/reconnectingWebSocket";
import { api } from "../context/AuthContext";

interface ServerMove {
  ply: number;
  san: string;
  from: string;
  to: string;
}

export default function Play() {
  const { game, fen, pgn, move, reset, undo, status } = useChessGame();
  const { gameId } = useParams();
  const stored = localStorage.getItem(`color_${gameId}`);
  const orientation: Color = stored === "black" ? "black" : "white";
  const [, setTurnColor] = useState<string>("white");
  const gameWsRef = useRef<ReconnectingWebSocket | null>(null);

  useEffect(() => {
    if (!gameId) return;

    const resync = async () => {
      try {
        const res = await api.get<{ moves: ServerMove[] }>(`/games/${gameId}`);
        const serverMoves = res.data.moves ?? [];
        const localCount = game.history().length;
        for (let i = localCount; i < serverMoves.length; i++) {
          move(serverMoves[i].from, serverMoves[i].to);
        }
      } catch (err) {
        console.error("resync failed:", err);
      }
    };

    const ws = new ReconnectingWebSocket("ws://localhost:8080/ws", {
      onOpen: () => {
        ws.send(JSON.stringify({ type: "auth", payload: { gameId } }));
        resync();
      },
      onMessage: (event) => {
        const data = JSON.parse(event.data);
        if (data.type === "move") {
          move(data.payload.from, data.payload.to);
          setTurnColor((t) => (t === "white" ? "black" : "white"));
        }
      },
    });
    gameWsRef.current = ws;

    return () => {
      ws.close();
      gameWsRef.current = null;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
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
        payload: { move: moveMade, from, to },
      })
    );

    setTurnColor((t) => (t === "white" ? "black" : "white"));
  };

  return (
    <div style={{ display: "flex", gap: 24, padding: 24, alignItems: "flex-start" }}>
      <div style={{ flexShrink: 0 }}>
        <ChessBoard
          game={game}
          fen={fen}
          onMove={onMove}
          orientation={orientation}
        />
      </div>
      <div>
        <GameInfo pgn={pgn} fen={fen} status={status} />
        <GameControls onReset={reset} onUndo={undo} onFlip={() => {}} />
      </div>
    </div>
  );
}
