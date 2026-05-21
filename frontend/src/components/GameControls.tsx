interface Props {
  onReset: () => void;
  onUndo: () => void;
  onFlip: () => void;
  onResign?: () => void;
  resignDisabled?: boolean;
}

export default function GameControls({ onReset, onUndo, onFlip, onResign, resignDisabled }: Props) {
  return (
    <div style={{ display: "flex", gap: 8 }}>
      <button onClick={onReset}>New Game</button>
      <button onClick={onFlip}>Flip Board</button>
      <button onClick={onUndo}>Undo</button>
      {onResign && (
        <button onClick={onResign} disabled={resignDisabled} style={{ color: "crimson" }}>
          Resign
        </button>
      )}
    </div>
  );
}
