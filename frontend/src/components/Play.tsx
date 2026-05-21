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
import type { Square } from "chess.js";
import { ReconnectingWebSocket } from "../lib/reconnectingWebSocket";
import { api } from "../context/AuthContext";

interface ServerMove {
  ply: number;
  san: string;
  from: string;
  to: string;
}

type PromotionPiece = "q" | "r" | "b" | "n";

export default function Play() {
  const { game, fen, pgn, move, reset, undo, status } = useChessGame();
  const { gameId } = useParams();
  const stored = localStorage.getItem(`color_${gameId}`);
  const orientation: Color = stored === "black" ? "black" : "white";
  const [, setTurnColor] = useState<string>("white");
  const [pendingPromotion, setPendingPromotion] = useState<{ from: string; to: string } | null>(null);
  const [gameOver, setGameOver] = useState<string | null>(null);
  const gameWsRef = useRef<ReconnectingWebSocket | null>(null);

  useEffect(() => {
    if (!gameId) return;

    const resync = async () => {
      try {
        const res = await api.get<{ moves: ServerMove[]; status: string }>(`/games/${gameId}`);
        const serverMoves = res.data.moves ?? [];
        const localCount = game.history().length;
        for (let i = localCount; i < serverMoves.length; i++) {
          move(serverMoves[i].from, serverMoves[i].to);
        }
        if (res.data.status === "finished") setGameOver("finished");
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
          move(data.payload.from, data.payload.to, data.payload.promotion ?? "q");
          setTurnColor((t) => (t === "white" ? "black" : "white"));
        }
        if (data.type === "game_over") {
          setGameOver(data.payload?.outcome ?? "finished");
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

  const isPromotion = (from: string, to: string): boolean => {
    const piece = game.get(from as Square);
    if (!piece || piece.type !== "p") return false;
    return (piece.color === "w" && to[1] === "8") || (piece.color === "b" && to[1] === "1");
  };

  const sendMove = (from: string, to: string, promotion: PromotionPiece) => {
    const prevLen = game.history().length;
    move(from, to, promotion);
    if (game.history().length === prevLen) return;

    const moveMade = game.history()[game.history().length - 1];

    gameWsRef.current?.send(
      JSON.stringify({
        type: "move",
        gameId,
        payload: { move: moveMade, from, to, promotion },
      })
    );

    setTurnColor((t) => (t === "white" ? "black" : "white"));
  };

  const onMove = (from: string, to: string) => {
    if (isPromotion(from, to)) {
      setPendingPromotion({ from, to });
      return;
    }
    sendMove(from, to, "q");
  };

  const onPromotionChoice = (piece: PromotionPiece) => {
    if (!pendingPromotion) return;
    sendMove(pendingPromotion.from, pendingPromotion.to, piece);
    setPendingPromotion(null);
  };

  const onResign = () => {
    if (!confirm("Resign this game?")) return;
    gameWsRef.current?.send(JSON.stringify({ type: "resign", gameId }));
  };

  return (
    <div style={{ display: "flex", gap: 24, padding: 24, alignItems: "flex-start" }}>
      <div style={{ flexShrink: 0, position: "relative" }}>
        <ChessBoard game={game} fen={fen} onMove={onMove} orientation={orientation} />
        {pendingPromotion && <PromotionPicker onChoice={onPromotionChoice} color={orientation} />}
        {gameOver && <GameOverOverlay outcome={gameOver} yourColor={orientation} />}
      </div>
      <div>
        <GameInfo pgn={pgn} fen={fen} status={status} />
        <GameControls
          onReset={reset}
          onUndo={undo}
          onFlip={() => {}}
          onResign={onResign}
          resignDisabled={!!gameOver}
        />
      </div>
    </div>
  );
}

function PromotionPicker({
  onChoice,
  color,
}: {
  onChoice: (p: PromotionPiece) => void;
  color: Color;
}) {
  const pieces: { id: PromotionPiece; label: string }[] = [
    { id: "q", label: color === "white" ? "♕" : "♛" },
    { id: "r", label: color === "white" ? "♖" : "♜" },
    { id: "b", label: color === "white" ? "♗" : "♝" },
    { id: "n", label: color === "white" ? "♘" : "♞" },
  ];
  return (
    <div
      style={{
        position: "absolute",
        top: "50%",
        left: "50%",
        transform: "translate(-50%, -50%)",
        background: "white",
        border: "1px solid #ccc",
        borderRadius: 8,
        boxShadow: "0 4px 16px rgba(0,0,0,0.3)",
        padding: 12,
        display: "flex",
        gap: 8,
        zIndex: 10,
      }}
    >
      {pieces.map((p) => (
        <button
          key={p.id}
          onClick={() => onChoice(p.id)}
          style={{ fontSize: 40, width: 60, height: 60, cursor: "pointer" }}
        >
          {p.label}
        </button>
      ))}
    </div>
  );
}

function GameOverOverlay({ outcome, yourColor }: { outcome: string; yourColor: Color }) {
  let text = "Game Over";
  if (outcome === "white" || outcome === "black") {
    const youWon = outcome === yourColor;
    text = youWon ? `You won (${outcome} wins)` : `You lost (${outcome} wins)`;
  } else if (outcome === "draw") {
    text = "Draw";
  }
  return (
    <div
      style={{
        position: "absolute",
        top: "50%",
        left: "50%",
        transform: "translate(-50%, -50%)",
        background: "rgba(0,0,0,0.85)",
        color: "white",
        padding: "20px 32px",
        borderRadius: 8,
        fontSize: 20,
        fontWeight: "bold",
        zIndex: 10,
      }}
    >
      {text}
    </div>
  );
}
