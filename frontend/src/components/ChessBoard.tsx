import { Chessground } from "chessground";
import { type Api } from "chessground/api";
import { useEffect, useRef } from "react";
import { Chess } from "chess.js";
import type { Color, Key } from "chessground/types";
interface Props {
  fen: string;
  onMove?: (from: string, to: string) => void;
  game?: Chess;
  orientation: Color;
  viewOnly?: boolean;
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

export default function ChessBoard({ fen, onMove, game, orientation, viewOnly }: Props) {
  const boardRef = useRef<HTMLDivElement>(null);
  const cgRef = useRef<Api | null>(null);
  const turnColor = game && game.turn() === "b" ? "black" : "white";

  useEffect(() => {
    if (!boardRef.current) return;
    const cg = Chessground(boardRef.current, {
      fen,
      turnColor,
      orientation,
      viewOnly: !!viewOnly,
      movable: viewOnly
        ? { free: false }
        : {
            free: false,
            dests: game ? buildDests(game) : new Map(),
            events: { after: onMove },
            color: orientation,
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
      movable: viewOnly
        ? { free: false }
        : {
            free: false,
            dests: game ? buildDests(game) : new Map(),
          },
    });
  }, [fen, game, viewOnly]);
  useEffect(() => {
    if (viewOnly) return;
    cgRef.current?.set({ orientation, movable: { color: orientation } });
  }, [orientation, viewOnly]);

  useEffect(() => {
    const onResize = () => cgRef.current?.redrawAll();
    window.addEventListener("resize", onResize);
    return () => window.removeEventListener("resize", onResize);
  }, []);

  return <div ref={boardRef} style={{ width: 560, height: 560, flexShrink: 0 }}></div>;
}
