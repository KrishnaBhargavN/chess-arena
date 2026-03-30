import axios from "axios";
import { useEffect } from "react";
import { useNavigate } from "react-router-dom";

const MatchMaking: React.FC = () => {
  const navigate = useNavigate();
  useEffect(() => {
    const fetchData = async () => {
      try {
        if (!localStorage.getItem("playerId")) {
          localStorage.setItem("playerId", crypto.randomUUID());
        }
        const res = await axios.post("http://localhost:8080/matchmaking/join", {
          playerId: localStorage.getItem("playerId"),
        });
        console.log(res.data);
        if (res.data.Status === "matched") {
          navigate(`/play/${res.data.gameId}`);
        }
      } catch (err) {
        console.log(err);
      }
    };
    fetchData();
  }, []);
  return <div>Waiting for other players to join...</div>;
};

export default MatchMaking;
