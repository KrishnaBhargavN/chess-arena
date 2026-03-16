import { useRef, useState } from "react";
import { Chess } from "chess.js";

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
    setPgn(gameRef.current.pgn());

    updateStatus();
  };

  const undo = () => {
    gameRef.current.undo();
    setFen(gameRef.current.fen());
    setPgn(gameRef.current.pgn());
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
