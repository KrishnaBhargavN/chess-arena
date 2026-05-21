import { useRef, useState } from "react";
import { Chess } from "chess.js";

function formatMoves(g: Chess): string {
  const moves = g.history();
  const out: string[] = [];
  for (let i = 0; i < moves.length; i += 2) {
    const num = i / 2 + 1;
    const white = moves[i];
    const black = moves[i + 1] ?? "";
    out.push(black ? `${num}. ${white} ${black}` : `${num}. ${white}`);
  }
  return out.join("  ");
}

export function useChessGame() {
  const gameRef = useRef(new Chess());

  const [fen, setFen] = useState(gameRef.current.fen());
  const [pgn, setPgn] = useState("");
  const [status, setStatus] = useState("White to move");

  const updateStatus = () => {
    const g = gameRef.current;

    const turn = g.turn() === "w" ? "White" : "Black";
    const opponent = g.turn() === "w" ? "Black" : "White";

    if (g.isCheckmate()) {
      setStatus(`Checkmate — ${opponent} wins`);
    } else if (g.isStalemate()) {
      setStatus("Stalemate — Draw");
    } else if (g.isThreefoldRepetition()) {
      setStatus("Draw by repetition");
    } else if (g.isInsufficientMaterial()) {
      setStatus("Draw by insufficient material");
    } else if (g.isDraw()) {
      setStatus("Draw");
    } else if (g.isCheck()) {
      setStatus(`${turn} is in check`);
    } else {
      setStatus(`${turn} to move`);
    }
  };

  const move = (from: string, to: string) => {
    const result = gameRef.current.move({
      from,
      to,
      promotion: "q",
    });

    if (!result) return;

    setFen(gameRef.current.fen());
    setPgn(formatMoves(gameRef.current));

    updateStatus();
  };

  const undo = () => {
    gameRef.current.undo();
    setFen(gameRef.current.fen());
    setPgn(formatMoves(gameRef.current));
    updateStatus();
  };

  const reset = () => {
    gameRef.current.reset();
    setFen(gameRef.current.fen());
    setPgn("");
    updateStatus();
  };

  return {
    game: gameRef.current,
    fen,
    pgn,
    status,
    move,
    undo,
    reset,
  };
}
