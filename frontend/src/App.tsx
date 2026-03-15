import React from "react";
import { BrowserRouter as Router, Routes, Route } from "react-router-dom";
import Navbar from "./components/Navbar";
import Hero from "./components/Hero";
import Play from "./components/Play";
import "./App.css";

// We extract the homepage content into its own mini-component for clean routing
const Home: React.FC = () => (
  <>
    <Hero />
  </>
);

const App: React.FC = () => {
  return (
    <Router>
      <div className="app-container">
        <Navbar />
        <Routes>
          {/* Landing Page */}
          <Route path="/" element={<Home />} />
          {/* Game Page */}
          <Route path="/play" element={<Play />} />
        </Routes>
      </div>
    </Router>
  );
};

export default App;
