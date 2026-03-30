import React from "react";
import { BrowserRouter as Router, Routes, Route } from "react-router-dom";
import Navbar from "./components/Navbar";
import Home from "./components/Home";
import Play from "./components/Play";
import "./App.css";
import { WSProvider } from "./context/WSContext";
import MatchMaking from "./components/MatchMaking";

const App: React.FC = () => {
  return (
    <Router>
      <WSProvider>
        <div className="app-container">
          <Navbar />
          <Routes>
            <Route path="/" element={<Home />} />
            <Route path="/play/:gameId" element={<Play />} />
            <Route path="/matchmaking" element={<MatchMaking />} />
          </Routes>
        </div>
      </WSProvider>
    </Router>
  );
};

export default App;
