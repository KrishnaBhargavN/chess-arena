import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { api, useAuth } from "../context/AuthContext";

interface GameSummary {
  id: string;
  playerA: string;
  playerB: string;
  playerAUsername?: string;
  playerBUsername?: string;
  playerAColor: string;
  playerBColor: string;
  status: string;
  createdAt: string;
  finishedAt?: string;
}

export default function GamesList() {
  const { user } = useAuth();
  const [games, setGames] = useState<GameSummary[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    api
      .get<GameSummary[]>("/games")
      .then((res) => setGames(res.data ?? []))
      .catch((err) => setError(err.message ?? "failed to load games"))
      .finally(() => setLoading(false));
  }, []);

  if (loading) return <div style={{ padding: 24 }}>Loading…</div>;
  if (error) return <div style={{ padding: 24, color: "crimson" }}>{error}</div>;
  if (games.length === 0) {
    return (
      <div style={{ padding: 24 }}>
        <h2>My Games</h2>
        <p>No games yet. <Link to="/matchmaking">Play a game</Link> to get started.</p>
      </div>
    );
  }

  return (
    <div style={{ padding: 24 }}>
      <h2>My Games</h2>
      <table style={{ width: "100%", borderCollapse: "collapse", marginTop: 16 }}>
        <thead>
          <tr style={{ textAlign: "left", borderBottom: "1px solid #ccc" }}>
            <th style={{ padding: 8 }}>Date</th>
            <th style={{ padding: 8 }}>You played</th>
            <th style={{ padding: 8 }}>Opponent</th>
            <th style={{ padding: 8 }}>Status</th>
            <th style={{ padding: 8 }}></th>
          </tr>
        </thead>
        <tbody>
          {games.map((g) => {
            const isPlayerA = g.playerA === user?.userId;
            const yourColor = isPlayerA ? g.playerAColor : g.playerBColor;
            const opponentId = isPlayerA ? g.playerB : g.playerA;
            const opponentUsername = isPlayerA ? g.playerBUsername : g.playerAUsername;
            const date = new Date(g.createdAt).toLocaleString();
            return (
              <tr key={g.id} style={{ borderBottom: "1px solid #eee" }}>
                <td style={{ padding: 8 }}>{date}</td>
                <td style={{ padding: 8, textTransform: "capitalize" }}>{yourColor}</td>
                <td style={{ padding: 8 }}>
                  <div>{opponentUsername ?? "—"}</div>
                  <div style={{ fontFamily: "monospace", fontSize: 11, color: "#888" }}>
                    {opponentId?.slice(0, 8) ?? "—"}
                  </div>
                </td>
                <td style={{ padding: 8 }}>{g.status}</td>
                <td style={{ padding: 8 }}>
                  <Link to={`/replay/${g.id}`}>Replay →</Link>
                </td>
              </tr>
            );
          })}
        </tbody>
      </table>
    </div>
  );
}
