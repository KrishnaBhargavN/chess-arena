import { useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { api } from "../context/AuthContext";

const MatchMaking: React.FC = () => {
  const navigate = useNavigate();

  useEffect(() => {
    const fetchData = async () => {
      try {
        const res = await api.post("/matchmaking/join");
        if (res.data.Status === "matched") {
          localStorage.setItem(`color_${res.data.gameId}`, res.data.playerColor);
          navigate(`/play/${res.data.gameId}`);
        }
      } catch (err) {
        console.log(err);
      }
    };
    fetchData();
  }, []);

  return <div style={{ padding: 24, textAlign: "center", fontSize: "1.2rem" }}>Waiting for other players to join...</div>;
};

export default MatchMaking;
