import "chessground/assets/chessground.base.css";
import "chessground/assets/chessground.brown.css";
import "chessground/assets/chessground.cburnett.css";
import ChessBoard from "./ChessBoard";
import GameControls from "./GameControls";
import GameInfo from "./GameInfo";
import { useChessGame } from "../hooks/useChessGame";
import { useEffect, useState } from "react";
import type { Color } from "chessground/types";
import Loading from "./MatchMaking";
import { useWS } from "../context/WSContext";
import { useParams } from "react-router-dom";

export default function Play() {
  const { game, fen, pgn, move, reset, undo, status } = useChessGame();
  const [orientation, setOrientation] = useState<Color>("white");
  const [loading, setLoading] = useState<boolean>(false);
  const ws = useWS();
  const flip = () => {
    setOrientation((o) => (o === "white" ? "black" : "white"));
  };
  const { gameId } = useParams();
  const onMove = (from: string, to: string) => {
    let prevMoveNum = game.history().length;
    console.log(`from: ${from}, to: ${to}`);

    move(from, to);
    if (prevMoveNum == game.history().length) {
      return;
    }
    const moveMade = game.history()[game.history().length - 1];
    console.log("moveMade: ", moveMade);

    ws?.current.send(
      JSON.stringify({
        type: "move",
        gameId: gameId,
        payload: {
          move: moveMade,
          from: from,
          to: to,
          playerId: localStorage.getItem("playerId"),
        },
      }),
    );
  };
  if (loading) {
    return <Loading />;
  }
  useEffect(() => {
    if (!ws?.current) return;

    ws.current.onmessage = (event) => {
      const data = JSON.parse(event.data);

      console.log("WS in Play:", data);

      if (data.type === "move") {
        move(data.payload.from, data.payload.to);
      }
    };
  }, [ws, move]);
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
        <GameControls onReset={reset} onUndo={undo} onFlip={flip} />
      </div>
    </div>
  );
}
