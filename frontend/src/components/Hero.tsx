import React from "react";
import { useNavigate } from "react-router-dom";

const Hero: React.FC = () => {
  const navigate = useNavigate();

  return (
    <header className="hero">
      <h1>
        Play Chess Online <br /> on the #1 Site!
      </h1>
      <p>
        Challenge players from around the world, solve tactics, and improve your
        game today. No downloads required.
      </p>
      <button className="btn btn-play" onClick={() => navigate("/play")}>
        Play Now
      </button>
    </header>
  );
};

export default Hero;
