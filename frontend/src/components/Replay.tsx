import "chessground/assets/chessground.base.css";
import "chessground/assets/chessground.brown.css";
import "chessground/assets/chessground.cburnett.css";
import { useEffect, useMemo, useState } from "react";
import { useParams } from "react-router-dom";
import { Chess } from "chess.js";
import type { Color } from "chessground/types";
import ChessBoard from "./ChessBoard";
import { api, useAuth } from "../context/AuthContext";

interface MoveRecord {
  ply: number;
  san: string;
  uci: string;
  from: string;
  to: string;
  by?: string;
  timestamp: string;
}

interface GameData {
  id: string;
  playerA: string;
  playerB: string;
  playerAColor: string;
  playerBColor: string;
  status: string;
  initial_fen: string;
  moves: MoveRecord[];
}

export default function Replay() {
  const { gameId } = useParams();
  const { user } = useAuth();
  const [data, setData] = useState<GameData | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [currentPly, setCurrentPly] = useState(0);

  useEffect(() => {
    if (!gameId) return;
    api
      .get<GameData>(`/games/${gameId}`)
      .then((res) => {
        setData(res.data);
        setCurrentPly(res.data.moves?.length ?? 0);
      })
      .catch((err) => setError(err.message ?? "failed to load game"));
  }, [gameId]);

  const fen = useMemo(() => {
    if (!data) return new Chess().fen();
    const chess = new Chess();
    for (let i = 0; i < currentPly && i < data.moves.length; i++) {
      try {
        chess.move(data.moves[i].san);
      } catch {
        break;
      }
    }
    return chess.fen();
  }, [data, currentPly]);

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (!data) return;
      if (e.key === "ArrowLeft") setCurrentPly((p) => Math.max(0, p - 1));
      if (e.key === "ArrowRight") setCurrentPly((p) => Math.min(data.moves.length, p + 1));
      if (e.key === "Home") setCurrentPly(0);
      if (e.key === "End") setCurrentPly(data.moves.length);
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [data]);

  if (error) return <div style={{ padding: 24, color: "crimson" }}>{error}</div>;
  if (!data) return <div style={{ padding: 24 }}>Loading…</div>;

  const isPlayerA = data.playerA === user?.userId;
  const yourColor = isPlayerA ? data.playerAColor : data.playerBColor;
  const orientation: Color = yourColor === "black" ? "black" : "white";

  const totalPly = data.moves.length;
  const pairs: { ply: number; white?: MoveRecord; black?: MoveRecord }[] = [];
  for (let i = 0; i < totalPly; i += 2) {
    pairs.push({
      ply: i / 2 + 1,
      white: data.moves[i],
      black: data.moves[i + 1],
    });
  }

  const moveButtonStyle = (ply: number): React.CSSProperties => ({
    background: currentPly === ply + 1 ? "#cce5ff" : "transparent",
    border: "none",
    padding: "2px 6px",
    cursor: "pointer",
    fontFamily: "monospace",
    fontSize: 14,
  });

  return (
    <div style={{ display: "flex", gap: 24, padding: 24, alignItems: "flex-start" }}>
      <div style={{ flexShrink: 0 }}>
        <ChessBoard fen={fen} orientation={orientation} viewOnly />
      </div>
      <div style={{ minWidth: 280 }}>
        <h3 style={{ margin: 0 }}>Replay</h3>
        <div style={{ fontSize: 13, color: "#666", marginBottom: 12 }}>
          Ply {currentPly} / {totalPly} · status: {data.status}
        </div>

        <div style={{ display: "flex", gap: 8, marginBottom: 16 }}>
          <button onClick={() => setCurrentPly(0)} disabled={currentPly === 0}>⏮ Start</button>
          <button onClick={() => setCurrentPly((p) => Math.max(0, p - 1))} disabled={currentPly === 0}>← Prev</button>
          <button onClick={() => setCurrentPly((p) => Math.min(totalPly, p + 1))} disabled={currentPly === totalPly}>Next →</button>
          <button onClick={() => setCurrentPly(totalPly)} disabled={currentPly === totalPly}>End ⏭</button>
        </div>

        <div style={{ maxHeight: 400, overflowY: "auto", border: "1px solid #ddd", padding: 8 }}>
          {pairs.length === 0 && <div style={{ color: "#888" }}>No moves played.</div>}
          {pairs.map(({ ply, white, black }) => (
            <div key={ply} style={{ display: "flex", gap: 8, alignItems: "center" }}>
              <span style={{ width: 28, color: "#888" }}>{ply}.</span>
              {white && (
                <button style={moveButtonStyle(white.ply)} onClick={() => setCurrentPly(white.ply + 1)}>
                  {white.san}
                </button>
              )}
              {black && (
                <button style={moveButtonStyle(black.ply)} onClick={() => setCurrentPly(black.ply + 1)}>
                  {black.san}
                </button>
              )}
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
