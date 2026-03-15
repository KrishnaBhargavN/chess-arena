import React, { useEffect, useRef, useState, useCallback } from "react";
import { Chess, type Square } from "chess.js";
import { Chessground } from "chessground"; // <- default import (fixed)
import type { Api } from "chessground/api";
import type { Key, Color } from "chessground/types";
import "chessground/assets/chessground.base.css";
import "chessground/assets/chessground.brown.css";
import "chessground/assets/chessground.cburnett.css";
// CSS imports — add these to your main.tsx or here:
// import "chessground/assets/chessground.base.css";
// import "chessground/assets/chessground.brown.css";   ← board theme
// import "chessground/assets/chessground.cburnett.css"; ← piece set

function buildDests(chess: Chess): Map<Key, Key[]> {
  const dests = new Map<Key, Key[]>();
  chess.moves({ verbose: true }).forEach((m: any) => {
    const from = m.from as Key;
    const existing = dests.get(from);
    if (existing) existing.push(m.to as Key);
    else dests.set(from, [m.to as Key]);
  });
  return dests;
}

function toColor(chess: Chess): Color {
  return chess.turn() === "w" ? "white" : "black";
}

const Play: React.FC = () => {
  const boardRef = useRef<HTMLDivElement | null>(null);
  const cgRef = useRef<Api | null>(null);
  const gameRef = useRef<Chess>(new Chess());

  const [status, setStatus] = useState<string>("White to move");
  const [fen, setFen] = useState<string>(() => gameRef.current.fen());
  const [pgn, setPgn] = useState<string>("");
  const [orientation, setOrientation] = useState<Color>("white");

  const updateInfo = useCallback(() => {
    const g = gameRef.current;
    const turn = g.turn() === "w" ? "White" : "Black";
    const opponent = g.turn() === "w" ? "Black" : "White";

    let s: string;
    if (g.isCheckmate()) s = `Checkmate — ${opponent} wins`;
    else if (g.isStalemate()) s = "Stalemate — Draw";
    else if (g.isThreefoldRepetition()) s = "Threefold repetition — Draw";
    else if (g.isInsufficientMaterial()) s = "Insufficient material — Draw";
    else if (g.isDraw()) s = "Draw";
    else if (g.isCheck()) s = `${turn} is in check!`;
    else s = `${turn} to move`;

    setStatus(s);
    setFen(g.fen());

    // strip header lines (things in square brackets) and whitespace — leaves the moves only
    const rawPgn = g.pgn();
    const movesOnly = rawPgn.replace(/(\[.*\]\r?\n)*/g, "").trim();
    setPgn(movesOnly ? movesOnly : "");
  }, []);

  const syncBoard = useCallback(() => {
    const g = gameRef.current;
    const cg = cgRef.current;
    if (!cg) return;

    cg.set({
      fen: g.fen(),
      turnColor: toColor(g),
      movable: {
        color: toColor(g),
        dests: g.isGameOver() ? new Map() : buildDests(g),
      },
      check: g.isCheck(),
    });

    updateInfo();
  }, [updateInfo]);

  const afterMove = useCallback(
    (orig: Key, dest: Key) => {
      const g = gameRef.current;

      // detect promotion
      const piece = g.get(orig as Square);
      const isPromotion =
        piece?.type === "p" &&
        ((piece.color === "w" && dest[1] === "8") ||
          (piece.color === "b" && dest[1] === "1"));

      try {
        g.move({
          from: orig as Square,
          to: dest as Square,
          promotion: isPromotion ? "q" : undefined,
        });
      } catch {
        // shouldn't happen: moves come from chess.js, but be defensive
      } finally {
        syncBoard();
      }
    },
    [syncBoard],
  );

  useEffect(() => {
    if (!boardRef.current) return;

    const g = gameRef.current;
    const cg = Chessground(boardRef.current, {
      fen: g.fen(),
      orientation,
      turnColor: toColor(g),
      movable: {
        color: toColor(g),
        free: false,
        dests: buildDests(g),
        showDests: true,
        events: {
          after: afterMove,
        },
      },
      draggable: {
        enabled: true,
        showGhost: true,
      },
      highlight: {
        lastMove: true,
        check: true,
      },
      animation: {
        enabled: true,
        duration: 200,
      },
      premovable: {
        enabled: false,
      },
    }) as Api;

    cgRef.current = cg;

    // ensure UI reflects chess state immediately
    updateInfo();

    return () => {
      cg.destroy();
      cgRef.current = null;
    };
    // intentionally mount once
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []); // mount once only

  const handleReset = useCallback(() => {
    gameRef.current = new Chess();
    const g = gameRef.current;
    const cg = cgRef.current;

    if (!cg) {
      updateInfo();
      return;
    }

    cg.set({
      fen: "start",
      lastMove: undefined,
      check: false,
      turnColor: "white",
      orientation, // keep current orientation
      movable: {
        color: "white",
        free: false,
        dests: buildDests(g),
        showDests: true,
        events: {
          after: afterMove,
        },
      },
    });

    updateInfo();
  }, [afterMove, orientation, updateInfo]);

  const handleFlip = useCallback(() => {
    // keep React state in sync for any UI that reads orientation
    const next: Color = orientation === "white" ? "black" : "white";
    setOrientation(next);

    // chessground provides toggleOrientation()
    cgRef.current?.toggleOrientation();
  }, [orientation]);

  const handleUndo = useCallback(() => {
    gameRef.current.undo();
    syncBoard();
  }, [syncBoard]);

  return (
    <div
      style={{
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        marginTop: "40px",
        gap: "16px",
        fontFamily: "sans-serif",
      }}
    >
      <h2 style={{ margin: 0, fontSize: "18px", fontWeight: 500 }}>{status}</h2>

      <div ref={boardRef} style={{ width: "560px", height: "560px" }} />

      <div style={{ display: "flex", gap: "8px" }}>
        {(["New Game", "Flip Board", "Undo Move"] as const).map((label, i) => (
          <button
            key={label}
            onClick={[handleReset, handleFlip, handleUndo][i]}
            style={btnStyle}
          >
            {label}
          </button>
        ))}
      </div>

      {pgn && (
        <div style={infoBoxStyle}>
          <strong>PGN:</strong> {pgn}
        </div>
      )}

      <div
        style={{
          fontSize: "11px",
          color: "#aaa",
          maxWidth: "560px",
          wordBreak: "break-all",
        }}
      >
        {fen}
      </div>
    </div>
  );
};

const btnStyle: React.CSSProperties = {
  padding: "8px 18px",
  fontSize: "13px",
  cursor: "pointer",
  borderRadius: "6px",
  border: "1px solid #ccc",
  background: "#fff",
  color: "#333",
};

const infoBoxStyle: React.CSSProperties = {
  maxWidth: "560px",
  width: "100%",
  fontSize: "13px",
  color: "#555",
  wordBreak: "break-all",
  padding: "8px 12px",
  background: "#f9f9f9",
  borderRadius: "6px",
  border: "1px solid #eee",
};

export default Play;
