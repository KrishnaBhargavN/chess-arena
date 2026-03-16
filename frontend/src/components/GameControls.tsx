interface Props {
  onReset: () => void;
  onUndo: () => void;
  onFlip: () => void;
}

export default function GameControls({ onReset, onUndo, onFlip }: Props) {
  return (
    <div style={{ display: "flex", gap: 8 }}>
      <button onClick={onReset}>New Game</button>
      <button onClick={onFlip}>Flip Board</button>
      <button onClick={onUndo}>Undo</button>
    </div>
  );
}
