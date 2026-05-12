import { Chessground } from "chessground";
import { type Api } from "chessground/api";
import { useEffect, useRef } from "react";
import { Chess } from "chess.js";
import type { Color, Key } from "chessground/types";
interface Props {
  fen: string;
  onMove: (from: string, to: string) => void;
  game: Chess;
  orientation: Color;
}

function buildDests(chess: Chess): Map<Key, Key[]> {
  const dests = new Map<Key, Key[]>();

  chess.moves({ verbose: true }).forEach((m) => {
    if (!dests.has(m.from)) {
      dests.set(m.from, []);
    }

    dests.get(m.from)?.push(m.to);
  });

  return dests;
}

export default function ChessBoard({ fen, onMove, game, orientation }: Props) {
  const boardRef = useRef<HTMLDivElement>(null);
  const cgRef = useRef<Api | null>(null);
  const turnColor = game.turn() === "w" ? "white" : "black";
  const data = localStorage.getItem(localStorage.getItem("playerId")??crypto.randomUUID());
  const playerColor : Color = data === "white"? "white" : "black";
  
  useEffect(() => {
    if (!boardRef.current) return;
    const cg = Chessground(boardRef.current, {
      fen,
      turnColor,
      orientation,
      movable: {
        free: false,
        dests: buildDests(game),
        events: {
          after: onMove,
        },
        color: playerColor
      },
    });

    cgRef.current = cg;
    return () => cg.destroy();
  }, []);

  useEffect(() => {
    cgRef.current?.set({
      fen,
      turnColor,
      orientation,
      movable: {
        free: false,
        dests: buildDests(game),
      },
    });
  }, [fen, game]);
  useEffect(() => {
    cgRef.current?.set({ orientation });
  }, [orientation]);

  useEffect(() => {
    const onResize = () => cgRef.current?.redrawAll();
    window.addEventListener("resize", onResize);
    return () => window.removeEventListener("resize", onResize);
  }, []);

  return <div ref={boardRef} style={{ width: 560, height: 560 }}></div>;
}
