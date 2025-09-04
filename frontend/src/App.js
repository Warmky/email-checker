// src/frontend/App.js
import { BrowserRouter as Router, Routes, Route } from "react-router-dom";
import MainPage from "./pages/MainPage";
import ConfigViewPage from "./pages/ConfigViewPage";

function App() {
    return (
        <Router>
            <Routes>
                <Route path="/" element={<MainPage />} />
                <Route path="/config-view" element={<ConfigViewPage />} />
            </Routes>
        </Router>
    );
}

export default App;