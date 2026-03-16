import "chessground/assets/chessground.base.css";
import "chessground/assets/chessground.brown.css";
import "chessground/assets/chessground.cburnett.css";
import ChessBoard from "./ChessBoard";
import GameControls from "./GameControls";
import GameInfo from "./GameInfo";
import { useChessGame } from "../hooks/useChessGame";
import { useEffect, useState } from "react";
import type { Color } from "chessground/types";
import Loading from "./Loading";

export default function Play() {
  const { game, fen, pgn, move, reset, undo, status } = useChessGame();
  const [orientation, setOrientation] = useState<Color>("white");
  const [loading, setLoading] = useState<boolean>(false);
  const flip = () => {
    setOrientation((o) => (o === "white" ? "black" : "white"));
  };
  if (loading) {
    return <Loading />;
  }
  useEffect(() => {}, []);
  return (
    <div>
      <div>
        <GameInfo pgn={pgn} fen={fen} status={status} />

        <ChessBoard
          game={game}
          fen={fen}
          onMove={(from: string, to: string) => move(from, to)}
          orientation={orientation}
        />
        <GameControls onReset={reset} onUndo={undo} onFlip={flip} />
      </div>
    </div>
  );
}
