interface Props {
  status: string;
  pgn: string;
  fen: string;
}

export default function GameInfo({ status, pgn, fen }: Props) {
  return (
    <>
      <h2>{status}</h2>

      {pgn && (
        <div>
          <strong>PGN:</strong> {pgn}
        </div>
      )}

      <div style={{ fontSize: 11 }}>{fen}</div>
    </>
  );
}
