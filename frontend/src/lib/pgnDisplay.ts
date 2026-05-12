import type { Chess } from "chess.js";

/** PGN date tag: YYYY.MM.DD */
export function pgnDateTag(d = new Date()): string {
  const y = d.getFullYear();
  const m = String(d.getMonth() + 1).padStart(2, "0");
  const day = String(d.getDate()).padStart(2, "0");
  return `${y}.${m}.${day}`;
}

function resultForPosition(game: Chess): string {
  if (game.isCheckmate()) {
    return game.turn() === "w" ? "0-1" : "1-0";
  }
  if (
    game.isStalemate() ||
    game.isDraw() ||
    game.isThreefoldRepetition() ||
    game.isInsufficientMaterial()
  ) {
    return "1/2-1/2";
  }
  return "*";
}

/** Sets readable seven-tag roster and a Result that matches the current position. */
export function syncPgnHeaders(game: Chess): void {
  game.setHeader("Event", "Casual game");
  game.setHeader("Site", "go-chess");
  game.setHeader("Date", pgnDateTag());
  game.setHeader("Round", "-");
  game.setHeader("White", "White");
  game.setHeader("Black", "Black");
  game.setHeader("Result", resultForPosition(game));
}

/** Multi-line PGN suitable for display or export. */
export function formatPgnForDisplay(game: Chess): string {
  syncPgnHeaders(game);
  return game.pgn({ newline: "\n", maxWidth: 72 });
}
